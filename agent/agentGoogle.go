package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/tools"
	"github.com/doptime/eloevo/utils"
	"github.com/samber/lo"
	"golang.design/x/clipboard"
	"google.golang.org/genai"
)

// AgentGoogle is responsible for interacting with a Google GenAI model,
// handling function calls, and managing state.
type AgentGoogle struct {
	SharedMemory map[string]any
	Models       []*models.Model // Assumes models.Model is adapted to hold *genai.Client

	Prompt         *template.Template
	Tools          *genai.Tool
	toolsCallbacks map[string]func(Param interface{}, CallMemory map[string]any) error

	// Configuration for message handling
	msgToMemKey      string
	msgDeFile        string
	msgToFile        string
	msgContentToFile string

	// Configuration for data persistence and external sources
	redisKey         string
	fieldReaderFunc  FieldReaderFunc
	msgFromCliboard  bool
	memDeCliboardKey string

	copyPromptOnly bool
	CallBack       func(ctx context.Context, inputs string) error

	// Mutex to ensure thread-safe execution of tool calls
	ToolCallRunningMutext *sync.Mutex
}

// NewAgentGoogle creates a new instance of AgentGoogle.
func NewAgentGoogle(prompt *template.Template, tools ...tool.ToolInterface) (a *AgentGoogle) {
	a = &AgentGoogle{
		Models:         []*models.Model{models.ModelDefault}, // Ensure ModelDefault is configured for GenAI
		Prompt:         prompt,
		toolsCallbacks: make(map[string]func(Param interface{}, CallMemory map[string]any) error),
		SharedMemory:   make(map[string]any),
	}
	a.WithTools(tools...)
	return a
}

// WithToolCallMutextRun enables a mutex for serializing tool calls.
func (a *AgentGoogle) WithToolCallMutextRun() *AgentGoogle {
	a.ToolCallRunningMutext = &sync.Mutex{}
	return a
}

// WithTools adds a set of tools (function declarations) to the agent.
// Note: This now creates a new agent instance to maintain immutability.
func (a *AgentGoogle) WithTools(tools ...tool.ToolInterface) *AgentGoogle {
	// Create a shallow copy of the agent
	ret := a.Clone()
	if ret.Tools == nil {
		ret.Tools = &genai.Tool{}
	}

	for _, t := range tools {
		ret.Tools.FunctionDeclarations = append(ret.Tools.FunctionDeclarations, t.GoogleGenaiTool())
		ret.toolsCallbacks[t.Name()] = t.HandleCallback
	}
	return ret
}

// --- Fluent configuration methods ---

func (a *AgentGoogle) WithMsgToMem(memoryKey string) *AgentGoogle {
	a.msgToMemKey = memoryKey
	return a
}

func (a *AgentGoogle) WithMsgDeFile(filename string) *AgentGoogle {
	a.msgDeFile = filename
	return a
}

func (a *AgentGoogle) WithMsgToFile(filename string) *AgentGoogle {
	a.msgToFile = filename
	return a
}

func (a *AgentGoogle) WithMsgContentToFile(filename string) *AgentGoogle {
	a.msgContentToFile = filename
	return a
}

func (a *AgentGoogle) ShareMemoryUpdate(MemoryCacheKey string, param interface{}) {
	if len(MemoryCacheKey) == 0 {
		return
	}
	a.SharedMemory[MemoryCacheKey] = param
}

func (a *AgentGoogle) WithContent2RedisHash(Key string, f FieldReaderFunc) *AgentGoogle {
	b := a.Clone()
	b.redisKey = Key
	b.fieldReaderFunc = f
	return b
}

// Clone creates a shallow copy of the agent.
func (a *AgentGoogle) Clone() *AgentGoogle {
	b := *a
	b.toolsCallbacks = make(map[string]func(Param interface{}, CallMemory map[string]any) error, len(a.toolsCallbacks))
	for k, v := range a.toolsCallbacks {
		b.toolsCallbacks[k] = v
	}
	*b.Tools = *a.Tools
	return &b
}

func (a *AgentGoogle) WithMsgDeClipboard() *AgentGoogle {
	a.msgFromCliboard = true
	return a
}

func (a *AgentGoogle) WithMemDeClipboard(memoryKey string) *AgentGoogle {
	a.memDeCliboardKey = memoryKey
	return a
}

func (a *AgentGoogle) WithModels(Model ...*models.Model) *AgentGoogle {
	a.Models = Model
	return a
}

func (a *AgentGoogle) WithCallback(callback func(ctx context.Context, inputs string) error) *AgentGoogle {
	a.CallBack = callback
	return a
}

func (a *AgentGoogle) CopyPromptOnly() *AgentGoogle {
	a.copyPromptOnly = true
	return a
}

// ExeResponse processes the function calls from a genai model's response.
func (a *AgentGoogle) ExeResponse(params map[string]any, resp *genai.GenerateContentResponse) (err error) {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil // No content to process
	}

	toolCallParts := lo.Filter(resp.Candidates[0].Content.Parts, func(part genai.ContentPart, _ int) bool {
		_, ok := part.(genai.FunctionCall)
		return ok
	})

	if len(toolCallParts) == 0 {
		return nil // No function calls in the response
	}

	ToolCallHash := make(map[uint64]bool)
	for _, part := range toolCallParts {
		toolcall := part.(genai.FunctionCall)

		// Convert args from map[string]interface{} to a consistent format for hashing
		argsBytes, jsonErr := json.Marshal(toolcall.Args)
		if jsonErr != nil {
			log.Printf("Warning: could not marshal tool call arguments for hashing: %v", jsonErr)
			continue
		}

		hash, _ := utils.GetCanonicalHash(string(argsBytes))
		if _, ok := ToolCallHash[hash]; ok {
			continue // Skip redundant tool call
		}
		ToolCallHash[hash] = true

		callback, ok := a.toolsCallbacks[toolcall.Name]
		if !ok {
			return fmt.Errorf("error: function '%s' not found in FunctionMap", toolcall.Name)
		}

		// Execute the callback
		execute := func() {
			err := callback(toolcall.Args, params)
			if err != nil {
				log.Printf("Error executing tool '%s': %v", toolcall.Name, err)
			}
		}

		if a.ToolCallRunningMutext != nil {
			a.ToolCallRunningMutext.Lock()
			execute()
			a.ToolCallRunningMutext.Unlock()
		} else {
			execute()
		}
	}
	return nil
}

// CallWithResponseString processes a string assumed to be a JSON representation of a genai response.
// This function is highly dependent on the string format and may be brittle.
func (a *AgentGoogle) CallWithResponseString(content string) (err error) {
	var params = make(map[string]any)
	for k, v := range a.SharedMemory {
		params[k] = v
	}

	// This is a significant change. We can't just convert a string to a genai.GenerateContentResponse.
	// We'll assume the string is the *text content* of the response, not the full response object.
	// To handle function calls, the full response object is needed.
	// A better approach would be to simulate a response if needed.
	// For now, this function might have limited utility unless the string contains function call JSON.
	log.Println("Warning: CallWithResponseString has limited functionality with the GenAI library.")
	log.Println("It cannot process function calls from a simple string.")

	// If you want to process text content from a string:
	if a.CallBack != nil {
		a.CallBack(context.Background(), content)
	}
	return nil
}

// Call executes the main logic: renders a prompt, sends it to the GenAI model, and processes the response.
func (a *AgentGoogle) Call(ctx context.Context, memories ...map[string]any) (err error) {
	// 1. Prepare parameters for the prompt template
	var params = make(map[string]any)
	for k, v := range a.SharedMemory {
		params[k] = v
	}
	for _, memory := range memories {
		for k, v := range memory {
			params[k] = v
		}
	}
	params["ThisAgentGoogle"] = a // Add self-reference to memory

	if a.memDeCliboardKey != "" {
		textbytes := clipboard.Read(clipboard.FmtText)
		if len(textbytes) == 0 {
			fmt.Println("no data in clipboard")
			return nil
		}
		params[a.memDeCliboardKey] = string(textbytes)
	}

	// 2. Render the prompt
	var promptBuffer bytes.Buffer
	if err := a.Prompt.Execute(&promptBuffer, params); err != nil {
		return fmt.Errorf("error rendering prompt: %w", err)
	}
	promptContent := promptBuffer.String()

	if a.copyPromptOnly {
		fmt.Println("Copying prompt to clipboard:", promptContent)
		clipboard.Write(clipboard.FmtText, []byte(promptContent))
		return nil
	}

	// 3. Select a model and initialize the GenAI client
	model := models.LoadbalancedPick(a.Models...)
	params["Model"] = model
	// Assumes model.Client is a *genai.Client
	if model.Client == nil {
		return fmt.Errorf("model '%s' has a nil genai.Client", model.Name)
	}
	gm := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: model.ApiKey,

		Model: model.Name,
	})

	// 4. Configure the generation settings
	gm.GenerationConfig = genai.GenerationConfig{}
	if model.Temperature > 0 {
		gm.Temperature = model.Temperature
	}
	if model.TopP > 0 {
		gm.TopP = model.TopP
	}
	if model.TopK > 0 {
		gm.N = model.TopK
	}
	if len(a.Tools.FunctionDeclarations) > 0 {
		gm.Tools = a.Tools
	}

	// 5. Send the request to the model
	timestart := time.Now()
	log.Printf("Sending request to model %s...", model.Name)
	resp, err := gm.GenerateContent(ctx, genai.Text(promptContent))
	if err != nil {
		model.ResponseTime(time.Since(timestart)) // Record time even on failure
		return fmt.Errorf("error generating content for model %s: %w", model.Name, err)
	}
	model.ResponseTime(time.Since(timestart))

	// 6. Process the response
	if resp == nil || len(resp.Candidates) == 0 {
		log.Println("Received an empty response from the model.")
		return nil
	}
	if resp.Candidates[0].FinishReason == genai.FinishReasonStop {
		log.Println("Model finished generation.")
	}

	// Extract and log the text content from the response
	var responseText strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText.WriteString(string(txt))
		}
	}
	fullResponseText := responseText.String()
	log.Printf("Model Response Text: %s", fullResponseText)

	// 7. Handle callbacks and memory updates
	if a.CallBack != nil {
		a.CallBack(ctx, fullResponseText)
	}
	if a.msgToMemKey != "" && len(memories) > 0 {
		memories[0][a.msgToMemKey] = fullResponseText
	}

	if a.redisKey != "" && a.fieldReaderFunc != nil && fullResponseText != "" {
		if field := a.fieldReaderFunc(fullResponseText); field != "" {
			tools.SaveToRedisHashKey(&tools.RedisHashKeyFieldValue{Key: a.redisKey, Field: field, Value: fullResponseText})
		}
	}

	// 8. Execute any function calls returned by the model
	return a.ExeResponse(params, resp)
}

// GetResponse is a deprecated placeholder. The logic is now inside Call.
// Kept for compatibility if other parts of the codebase use it, but should be removed.
func (a *AgentGoogle) GetResponse(client *genai.Client, req interface{}) (*genai.GenerateContentResponse, error) {
	// This function's signature is problematic as it mixes concerns.
	// The new `Call` method handles this flow correctly.
	return nil, fmt.Errorf("GetResponse is deprecated; use the Call method instead")
}

// --- Helper for converting old tool format to new genai.Schema format ---
// You will need to implement this based on your `tool.ToolInterface` and its `OaiTool()` method.
// This is a conceptual example.
func convertParamsToSchema(params interface{}) *genai.Schema {
	// Assuming params is a JSON-like structure (e.g., map[string]interface{})
	// that describes the OpenAI function parameters.
	pBytes, err := json.Marshal(params)
	if err != nil {
		return nil
	}

	var schema genai.Schema
	if err := json.Unmarshal(pBytes, &schema); err != nil {
		log.Printf("Failed to convert OpenAI params to genai.Schema: %v", err)
		return nil
	}
	return &schema
}

// You need to add a GenaiTool() method to your ToolInterface
// Example:
/*
type ToolInterface interface {
    Name() string
    HandleCallback(Param interface{}, CallMemory map[string]any) error
    OaiTool() *SomeOpenAIStructure // Old method
    GenaiTool() *genai.Tool        // New method
}
*/

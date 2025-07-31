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
	Models       []*models.Model // Assumes models.Model is adapted to hold genai client info

	Prompt *template.Template
	Tools  *genai.Tool
	//【注意】Go的map是引用类型，Clone时需要深拷贝
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
		Models:         []*models.Model{models.ModelDefault}, // Ensure ModelDefault is configured
		Prompt:         prompt,
		toolsCallbacks: make(map[string]func(Param interface{}, CallMemory map[string]any) error),
		SharedMemory:   make(map[string]any),
		Tools:          &genai.Tool{}, // Initialize Tools to avoid nil pointer
	}
	// Note: WithTools now returns a new instance, so we re-assign it.
	return a.WithTools(tools...)
}

// WithToolCallMutextRun enables a mutex for serializing tool calls.
func (a *AgentGoogle) WithToolCallMutextRun() *AgentGoogle {
	a.ToolCallRunningMutext = &sync.Mutex{}
	return a
}

// WithTools adds a set of tools (function declarations) to the agent.
// This method returns a new agent instance to maintain immutability.
func (a *AgentGoogle) WithTools(tools ...tool.ToolInterface) *AgentGoogle {
	// Create a shallow copy of the agent
	ret := a.Clone()

	for _, t := range tools {
		// Append to the new agent's tool set
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

// Clone creates a new AgentGoogle instance with copied values.
func (a *AgentGoogle) Clone() *AgentGoogle {
	b := *a // Shallow copy of the struct itself

	// Deep copy the toolsCallbacks map
	b.toolsCallbacks = make(map[string]func(Param interface{}, CallMemory map[string]any) error, len(a.toolsCallbacks))
	for k, v := range a.toolsCallbacks {
		b.toolsCallbacks[k] = v
	}

	// Deep copy the Tools struct to ensure slice independence
	if a.Tools != nil {
		newTools := &genai.Tool{}
		// Create a new slice with the same capacity and copy elements
		newTools.FunctionDeclarations = make([]*genai.FunctionDeclaration, len(a.Tools.FunctionDeclarations))
		copy(newTools.FunctionDeclarations, a.Tools.FunctionDeclarations)
		b.Tools = newTools
	}

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

	// Filter for function call parts
	// toolCallParts := lo.Filter(resp.Candidates[0].Content.Parts, func(part *genai.Part, _ int) bool {
	// 	return part.FunctionCall != nil && part.FunctionCall.Name != "" && part.FunctionCall.Args != nil
	// })
	toolCalls := lo.Map(resp.Candidates[0].Content.Parts, func(candidate *genai.Part, _ int) *genai.FunctionCall {
		return candidate.FunctionCall
	})

	if len(toolCalls) == 0 {
		return nil // No function calls in the response
	}

	ToolCallHash := make(map[uint64]bool)
	for _, toolcall := range toolCalls {
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
			// It's better to log a warning and continue, rather than failing the entire process
			log.Printf("Warning: function '%s' not found in FunctionMap, skipping.", toolcall.Name)
			continue
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

// CallWithResponseString processes a simple string. Note: This cannot handle function calls.
func (a *AgentGoogle) CallWithResponseString(content string) (err error) {
	log.Println("CallWithResponseString invoked. Note: This function only processes the text content and cannot execute function calls.")

	if a.CallBack != nil {
		return a.CallBack(context.Background(), content)
	}
	return nil
}

// Call executes the main logic: renders a prompt, sends it to the GenAI model, and processes the response.
func (a *AgentGoogle) Call(ctx context.Context, memories ...map[string]any) (err error) {
	// 1. Prepare parameters for the prompt template
	params := make(map[string]any)
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
		err := clipboard.Init()
		if err != nil {
			return fmt.Errorf("error initializing clipboard: %w", err)
		}
		fmt.Println("Copying prompt to clipboard:", promptContent)
		clipboard.Write(clipboard.FmtText, []byte(promptContent))
		return nil
	}

	// 3. Select a model and initialize the GenAI client
	modelInfo := models.LoadbalancedPick(a.Models...)
	params["Model"] = modelInfo
	if modelInfo.ApiKey == "" {
		return fmt.Errorf("model '%s' has an empty API key", modelInfo.Name)
	}

	// 【修正】使用 option.WithAPIKey 正确创建客户端
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: modelInfo.ApiKey,
		HTTPOptions: genai.HTTPOptions{
			BaseURL: modelInfo.BaseURL,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating genai client: %w", err)
	}

	// 4. 【修正】在 GenerativeModel 实例上配置生成参数
	GenerationConfig := genai.GenerateContentConfig{}
	if modelInfo.Temperature != 0 {
		GenerationConfig.Temperature = &modelInfo.Temperature
	}
	if modelInfo.TopP != 0 {
		GenerationConfig.TopP = &modelInfo.TopP
	}
	if modelInfo.TopK != 0 {
		GenerationConfig.TopK = &modelInfo.TopK
	}
	if a.Tools != nil && len(a.Tools.FunctionDeclarations) > 0 {
		GenerationConfig.Tools = append(GenerationConfig.Tools, a.Tools)
	}

	// 5. Send the request to the model
	timestart := time.Now()
	log.Printf("Sending request to model %s...", modelInfo.Name)
	msg := genai.Text(promptContent)

	resp, err := client.Models.GenerateContent(context.Background(), modelInfo.Name, msg, &GenerationConfig)
	if err != nil {
		modelInfo.ResponseTime(time.Since(timestart)) // Record time even on failure
		return fmt.Errorf("error generating content for model %s: %w", modelInfo.Name, err)
	}
	modelInfo.ResponseTime(time.Since(timestart))

	// 6. Process the response
	if resp == nil || len(resp.Candidates) == 0 {
		log.Println("Received an empty response from the model.")
		return nil
	}

	// Extract and log the text content from the response
	var responseText strings.Builder
	if resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				responseText.WriteString(part.Text)
			}
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

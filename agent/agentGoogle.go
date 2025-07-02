package agent

import (
	"bytes"
	"context"
	"fmt"
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
	genai "google.golang.org/genai"
)

// GoalProposer is responsible for proposing goals using an genai model,
// handling function calls, and managing callbacks.
type AgentGoogle struct {
	SharedMemory map[string]any
	Models       []*models.Model

	Prompt              *template.Template
	Tools               []genai.Tool
	toolsCallbacks      map[string]func(Param interface{}, CallMemory map[string]any) error
	msgToMemKey         string
	msgDeFile           string
	msgToFile           string
	msgContentToFile    string
	redisKey            string
	fieldReaderFunc     FieldReaderFunc
	msgFromCliboard     bool
	memDeCliboardKey    string
	functioncallParsers []func(resp genai.ChatCompletionResponse) (toolCalls []*FunctionCall)

	copyPromptOnly bool
	CallBack       func(ctx context.Context, inputs string) error

	ToolCallRunningMutext interface{}
}

func NewAgentGoogle(prompt *template.Template, tools ...tool.ToolInterface) (a *AgentGoogle) {
	a = &AgentGoogle{
		Models:         []*models.Model{models.ModelDefault},
		Prompt:         prompt,
		toolsCallbacks: map[string]func(Param interface{}, CallMemory map[string]any) error{},
		SharedMemory:   map[string]any{},
	}
	a.WithTools(tools...)
	return a
}
func (a *AgentGoogle) WithToolCallMutextRun() *AgentGoogle {
	a.ToolCallRunningMutext = &sync.Mutex{}
	return a
}
func (a *AgentGoogle) WithTools(tools ...tool.ToolInterface) (ret *AgentGoogle) {
	ret = &AgentGoogle{}
	*ret = *a
	for _, tool := range tools {
		ret.Tools = append(ret.Tools, *tool.OaiTool())
		ret.toolsCallbacks[tool.Name()] = tool.HandleCallback
	}
	return ret
}
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
	var b AgentGoogle = *a
	b.redisKey = Key
	b.fieldReaderFunc = f
	return &b
}
func (a *AgentGoogle) Clone() *AgentGoogle {
	var b AgentGoogle = *a
	b.toolsCallbacks = map[string]func(Param interface{}, CallMemory map[string]any) error{}
	for k, v := range a.toolsCallbacks {
		b.toolsCallbacks[k] = v
	}
	b.Tools = append([]genai.Tool{}, a.Tools...)

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

func (a *AgentGoogle) ExeResponse(params map[string]any, resp genai.ChatCompletionResponse) (err error) {
	var toolCalls []*FunctionCall
	for _, parser := range a.functioncallParsers {
		toolCalls = append(toolCalls, parser(resp)...)
	}
	ToolCallHash := map[uint64]bool{}
	for _, toolcall := range toolCalls {
		//skip redundant toolcall
		hash, _ := utils.GetCanonicalHash(toolcall.Arguments)
		if _, ok := ToolCallHash[hash]; ok {
			continue
		}
		ToolCallHash[hash] = true

		_tool, ok := a.toolsCallbacks[toolcall.Name]
		if ok {
			if a.ToolCallRunningMutext != nil {
				a.ToolCallRunningMutext.(*sync.Mutex).Lock()
				_tool(toolcall.Arguments, params)
				a.ToolCallRunningMutext.(*sync.Mutex).Unlock()
			} else {
				_tool(toolcall.Arguments, params)
			}
		} else if !ok {
			return fmt.Errorf("error: function not found in FunctionMap")
		}
	}
	return nil

}
func (a *AgentGoogle) CallWithResponseString(content string) (err error) {
	var params = map[string]any{}
	for k, v := range a.SharedMemory {
		params[k] = v
	}
	resp, err := utils.StringToResponse(content)
	if err != nil {
		return err
	}
	return a.ExeResponse(params, resp)
}

// ProposeGoals generates goals based on the provided file contents.
// It renders the prompt, sends a request to the genai model, and processes the response.
func (a *AgentGoogle) Call(ctx context.Context, memories ...map[string]any) (err error) {
	// Render the prompt with the provided files content and available functions
	var params = map[string]any{}
	for k, v := range a.SharedMemory {
		params[k] = v
	}
	for _, memory := range memories {
		for k, v := range memory {
			params[k] = v
		}
	}
	params["ThisAgentGoogle"] = a // add self reference to memory
	if a.memDeCliboardKey != "" {
		textbytes := clipboard.Read(clipboard.FmtText)
		if len(textbytes) == 0 {
			fmt.Println("no data in clipboard")
			return nil
		}
		params[a.memDeCliboardKey] = string(textbytes)
	}
	var promptBuffer bytes.Buffer
	if err := a.Prompt.Execute(&promptBuffer, params); err != nil {
		fmt.Printf("Error rendering prompt: %v\n", err)
		return err
	}

	//model might be changed by other process
	model := models.LoadbalancedPick(a.Models...)
	params["Model"] = model
	// Create the chat completion request with function calls enabled
	req := genai.ChatCompletionRequest{
		Model: model.Name,
		Messages: []genai.ChatCompletionMessage{
			{
				Role:    genai.ChatMessageRoleUser,
				Content: promptBuffer.String(),
			},
		},
		TopP:        model.TopP,
		Temperature: model.Temperature,
	}
	if model.Temperature > 0 {
		req.Temperature = model.Temperature
	}
	if model.TopP > 0 {
		req.TopP = model.TopP
	}
	if model.TopK > 0 {
		req.N = model.TopK
	}
	if len(a.Tools) > 0 {
		if model.ToolInPrompt != nil {
			model.ToolInPrompt.WithToolcallSysMsg(a.Tools, &req)
		} else {
			req.Tools = a.Tools
		}
	}

	if a.copyPromptOnly {
		msg := strings.Join(lo.Map(req.Messages, func(m genai.ChatCompletionMessage, _ int) string { return m.Content }), "\n")
		fmt.Println("copy prompt to clipboard", msg)
		clipboard.Write(clipboard.FmtText, []byte(msg))
		return nil
	}
	timestart := time.Now()
	resp, err := a.GetResponse(model.Client, req)
	if err == nil {
		model.ResponseTime(time.Since(timestart))
		reqMesseges := lo.Map(req.Messages, func(m genai.ChatCompletionMessage, _ int) string {
			return m.Content
		})
		resmesseges := lo.Map(resp.Choices, func(c genai.ChatCompletionChoice, _ int) string {
			return c.Message.Content
		})
		toolCalls := lo.Map(resp.Choices, func(c genai.ChatCompletionChoice, _ int) any {
			return c.Message.FunctionCall
		})
	}
	fmt.Println("resp:", resp)
	if err != nil {
		fmt.Println("Error creating chat completion:", err)
		fmt.Println("req:", req.Messages[0].Content)
		return err
	}
	if a.CallBack != nil {
		a.CallBack(ctx, resp.Choices[0].Message.Content)
	}
	if a.msgToMemKey != "" && len(memories) > 0 {
		memories[0][a.msgToMemKey] = resp.Choices[0].Message.Content
	}

	if a.redisKey != "" && a.fieldReaderFunc != nil && len(resp.Choices) > 0 {
		if field := a.fieldReaderFunc(resp.Choices[0].Message.Content); field != "" {
			tools.SaveToRedisHashKey(&tools.RedisHashKeyFieldValue{Key: a.redisKey, Field: field, Value: resp.Choices[0].Message.Content})
		}
	}
	return a.ExeResponse(params, resp)
}

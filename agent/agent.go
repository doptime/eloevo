package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/doptime/eloevo/memory"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/tools"
	"github.com/doptime/eloevo/utils"
	openai "github.com/sashabaranov/go-openai"
	"golang.design/x/clipboard"
)

type FileToMem struct {
	File string `json:"file"`
	Mem  string `json:"mem"`
}

// GoalProposer is responsible for proposing goals using an OpenAI model,
// handling function calls, and managing callbacks.
type Agent struct {
	Model               *models.Model
	Prompt              *template.Template
	Tools               []openai.Tool
	toolsCallbacks      map[string]func(Param interface{}, CallMemory map[string]any) error
	msgToMemKey         string
	fileToMem           *FileToMem
	msgDeFile           string
	msgToFile           string
	msgContentToFile    string
	redisKey            string
	fieldReaderFunc     FieldReaderFunc
	msgDeCliboard       bool
	memDeCliboardKey    string
	functioncallParsers []func(resp openai.ChatCompletionResponse) (toolCalls []*FunctionCall)

	copyPromptOnly bool
	CallBack       func(ctx context.Context, inputs string) error

	ToolCallRunningMutext interface{}
}

func NewAgent(prompt *template.Template, tools ...tool.ToolInterface) (a *Agent) {
	a = &Agent{
		Model:          models.ModelDefault,
		Prompt:         prompt,
		toolsCallbacks: map[string]func(Param interface{}, CallMemory map[string]any) error{},
	}
	a.WithTools(tools...)
	a.WithToolcallParser(nil)
	return a
}
func (a *Agent) WithToolCallLocked() *Agent {
	a.ToolCallRunningMutext = &sync.Mutex{}
	return a
}
func (a *Agent) WithTools(tools ...tool.ToolInterface) *Agent {
	for _, tool := range tools {
		a.Tools = append(a.Tools, *tool.OaiTool())
		a.toolsCallbacks[tool.Name()] = tool.HandleCallback
	}
	return a
}
func (a *Agent) WithFileToMem(filename, memoryKey string) *Agent {
	a.fileToMem = &FileToMem{File: filename, Mem: memoryKey}
	return a
}
func (a *Agent) WithMsgToMem(memoryKey string) *Agent {
	a.msgToMemKey = memoryKey
	return a
}
func (a *Agent) WithMsgDeFile(filename string) *Agent {
	a.msgDeFile = filename
	return a
}
func (a *Agent) WithMsgToFile(filename string) *Agent {
	a.msgToFile = filename
	return a
}
func (a *Agent) WithMsgContentToFile(filename string) *Agent {
	a.msgContentToFile = filename
	return a
}

type FieldReaderFunc func(content string) (field string)

func (a *Agent) WithContent2RedisHash(Key string, f FieldReaderFunc) *Agent {
	var b Agent = *a
	b.redisKey = Key
	b.fieldReaderFunc = f
	return &b
}
func (a *Agent) Clone() *Agent {
	var b Agent = *a
	b.toolsCallbacks = map[string]func(Param interface{}, CallMemory map[string]any) error{}
	for k, v := range a.toolsCallbacks {
		b.toolsCallbacks[k] = v
	}
	b.Tools = append([]openai.Tool{}, a.Tools...)

	return &b
}
func (a *Agent) WithMsgDeClipboard() *Agent {
	a.msgDeCliboard = true
	return a
}
func (a *Agent) WithMemDeClipboard(memoryKey string) *Agent {
	a.memDeCliboardKey = memoryKey
	return a
}
func (a *Agent) WithModel(Model *models.Model) *Agent {
	a.Model = Model
	return a
}

func (a *Agent) WithCallback(callback func(ctx context.Context, inputs string) error) *Agent {
	a.CallBack = callback
	return a
}
func (a *Agent) CopyPromptOnly() *Agent {
	a.copyPromptOnly = true
	return a
}
func (a *Agent) withToolcallSysMsg(req *openai.ChatCompletionRequest) {

	if len(a.Tools) == 0 {
		return
	}
	ToolCallMsg, err := template.New("ToolCallMsg").Parse(`
# Tools

You may call one or more functions to assist with the user query.

You are provided with function signatures within <tools></tools> XML tags:

<tools>
	{{range $ind, $val := .Tools}}
		{{$val}}
	{{end}}
</tools>

For each function call, return a json object with function name and arguments within <tool_call></tool_call> XML tags:
<tool_call>
{\"name\": <function-name>, \"arguments\": <args-json-object>}
</tool_call>
`)
	if err != nil {
		fmt.Println("Error parsing ToolCallMsg template:", err)
		return
	}
	ToolStr := []string{}
	for _, v := range a.Tools {
		toolBytes, _ := json.Marshal(v)
		ToolStr = append(ToolStr, string(toolBytes))
	}

	var promptBuffer bytes.Buffer
	if err := ToolCallMsg.Execute(&promptBuffer, map[string]any{"Tools": ToolStr}); err == nil {
		msgToolCall := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleSystem, Content: promptBuffer.String()}
		req.Messages = append([]openai.ChatCompletionMessage{msgToolCall}, req.Messages...)
	}
}

// ProposeGoals generates goals based on the provided file contents.
// It renders the prompt, sends a request to the OpenAI model, and processes the response.
func (a *Agent) Call(ctx context.Context, memories ...map[string]any) (err error) {
	// Render the prompt with the provided files content and available functions
	var params = map[string]any{}
	for k, v := range memory.SharedMemory {
		params[k] = v
	}
	for _, memory := range memories {
		for k, v := range memory {
			params[k] = v
		}
	}
	if a.memDeCliboardKey != "" {
		textbytes := clipboard.Read(clipboard.FmtText)
		if len(textbytes) == 0 {
			fmt.Println("no data in clipboard")
			return nil
		}
		params[a.memDeCliboardKey] = string(textbytes)
	}
	if a.fileToMem != nil {
		resp, err := utils.FileToResponse(a.fileToMem.File)
		if err == nil {
			params[a.fileToMem.Mem] = resp.Choices[0].Message.Content
		}
	}

	var promptBuffer bytes.Buffer
	if err := a.Prompt.Execute(&promptBuffer, params); err != nil {
		fmt.Printf("Error rendering prompt: %v\n", err)
		return err
	}

	//model might be changed by other process
	model := a.Model
	if md, ok := params["Model"]; ok {
		model = md.(*models.Model)
	}
	// Create the chat completion request with function calls enabled
	req := openai.ChatCompletionRequest{
		Model: model.Name,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
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
		if model.ToolInPrompt {
			a.withToolcallSysMsg(&req)
		} else {
			req.Tools = a.Tools
		}
	}

	if a.copyPromptOnly {
		fmt.Println("copy prompt to clipboard", req.Messages[0].Content)
		clipboard.Write(clipboard.FmtText, []byte(req.Messages[0].Content))
		return nil
	}
	timestart := time.Now()
	resp, err := a.GetResponse(model.Client, req)
	if err == nil {
		model.ResponseTime(time.Since(timestart))
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

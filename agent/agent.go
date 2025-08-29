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
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
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
	SharedMemory           map[string]any
	Models                 []*models.Model
	AbbreviateContextModel *models.Model

	Prompt              *template.Template
	Tools               []openai.Tool
	ToolInSystemPrompt  bool
	ToolInUserPrompt    bool
	toolsCallbacks      map[string]func(Param interface{}, CallMemory map[string]any) error
	msgToMemKey         string
	msgDeFile           string
	msgToFile           string
	msgContentToFile    string
	redisKey            string
	fieldReaderFunc     FieldReaderFunc
	msgFromCliboard     bool
	memDeCliboardKey    string
	functioncallParsers []func(resp openai.ChatCompletionResponse) (toolCalls []*FunctionCall)

	copyPromptOnly bool
	CallBack       func(ctx context.Context, inputs string) error

	ToolCallRunningMutext interface{}
}

func NewAgent(prompt *template.Template, tools ...tool.ToolInterface) (a *Agent) {
	a = &Agent{
		Models:         []*models.Model{models.ModelDefault},
		Prompt:         prompt,
		toolsCallbacks: map[string]func(Param interface{}, CallMemory map[string]any) error{},
		SharedMemory:   map[string]any{},
	}
	a.WithTools(tools...)
	a.WithToolcallParser(nil)
	return a
}
func (a *Agent) WithToolCallMutextRun() *Agent {
	a.ToolCallRunningMutext = &sync.Mutex{}
	return a
}
func (a *Agent) WithTools(tools ...tool.ToolInterface) (ret *Agent) {
	ret = &Agent{}
	*ret = *a
	for _, tool := range tools {
		ret.Tools = append(ret.Tools, *tool.OaiTool())
		ret.toolsCallbacks[tool.Name()] = tool.HandleCallback
	}
	return ret
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

func (a *Agent) ShareMemoryUpdate(MemoryCacheKey string, param interface{}) {
	if len(MemoryCacheKey) == 0 {
		return
	}
	a.SharedMemory[MemoryCacheKey] = param
}

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
	a.msgFromCliboard = true
	return a
}
func (a *Agent) WithMemDeClipboard(memoryKey string) *Agent {
	a.memDeCliboardKey = memoryKey
	return a
}
func (a *Agent) WithModels(Model ...*models.Model) *Agent {
	a.Models = Model
	return a
}
func (a *Agent) WithAbbreviateContext(Model *models.Model) *Agent {
	a.AbbreviateContextModel = Model
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

type QAPaire struct {
	Time      time.Time
	Model     string
	Question  any
	Response  any
	ToolCalls any
}

var keyQA = redisdb.NewListKey[*QAPaire](redisdb.Opt.Rds("Catalogs"))

func (a *Agent) ExeResponse(params map[string]any, resp openai.ChatCompletionResponse) (err error) {
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
func (a *Agent) CallWithResponseString(content string) (err error) {
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
// It renders the prompt, sends a request to the OpenAI model, and processes the response.
func (a *Agent) Call(ctx context.Context, memories ...map[string]any) (err error) {
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
	params["ThisAgent"] = a // add self reference to memory
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
	fmt.Printf("Requesting prompt: %v\n", promptBuffer.String())

	//model might be changed by other process
	model, ok := params["Model"].(*models.Model)
	if !ok || model == nil {
		model = models.LoadbalancedPick(a.Models...)
		params["Model"] = models.LoadbalancedPick(a.Models...)
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
	if a.AbbreviateContextModel != nil {
		messege := promptBuffer.String()
		AbbreviatedMessege := AbbreviateContext(messege, a.AbbreviateContextModel)
		req.Messages[0].Content = AbbreviatedMessege

	}
	if len(a.Tools) > 0 {
		if model.ToolInPrompt != nil {
			model.ToolInPrompt.WithToolcallSysMsg(a.Tools, &req)
		} else {
			req.Tools = a.Tools
		}
	}

	if a.copyPromptOnly {
		msg := strings.Join(lo.Map(req.Messages, func(m openai.ChatCompletionMessage, _ int) string { return m.Content }), "\n")
		err := clipboard.Init()
		if err != nil {
			return fmt.Errorf("error initializing clipboard: %w", err)
		}
		//remove \x00 in the msg
		var sb strings.Builder
		for _, r := range msg {
			if r != '\x00' {
				sb.WriteRune(r)
			}
		}
		fmt.Println("copy prompt to clipboard", msg)
		msg = sb.String()
		clipboard.Write(clipboard.FmtText, []byte(msg))
		return nil
	}
	timestart := time.Now()
	resp, err := a.GetResponse(model.Client, req)
	if err == nil {
		model.ResponseTime(time.Since(timestart))
		reqMesseges := lo.Map(req.Messages, func(m openai.ChatCompletionMessage, _ int) string {
			return m.Content
		})
		resmesseges := lo.Map(resp.Choices, func(c openai.ChatCompletionChoice, _ int) string {
			return c.Message.Content
		})
		toolCalls := lo.Map(resp.Choices, func(c openai.ChatCompletionChoice, _ int) any {
			return c.Message.FunctionCall
		})
		keyQA.LPush(&QAPaire{Time: time.Now(), Model: model.Name, Question: reqMesseges, Response: resmesseges, ToolCalls: toolCalls})
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

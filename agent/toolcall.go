package agent

import (
	"encoding/json"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Process each choice in the response
type FunctionCall struct {
	Name string `json:"name,omitempty"`
	// call function with arguments in JSON format
	Arguments any `json:"arguments,omitempty"`
}

func parseOneToolcall(toolcallString string) (toolCalls *FunctionCall) {
	tool := FunctionCall{Name: "", Arguments: map[string]any{}}
	//openai.FunctionCall 中的Arguments是string类型.直接unmrshal 会报错
	err := json.Unmarshal([]byte(toolcallString), &tool)
	if err == nil && tool.Name != "" && tool.Arguments != nil {
		return &tool
	}
	return nil
}
func ToolcallParserDefault(resp openai.ChatCompletionResponse) (toolCalls []*FunctionCall) {
	for _, choice := range resp.Choices {
		for _, toolcall := range choice.Message.ToolCalls {
			functioncall := &FunctionCall{
				Name:      toolcall.Function.Name,
				Arguments: toolcall.Function.Arguments,
			}
			toolCalls = append(toolCalls, functioncall)
		}
	}
	if len(toolCalls) == 0 && len(resp.Choices) > 0 {
		rsp := resp.Choices[0].Message.Content

		rsp = strings.ReplaceAll(rsp, "tool_code>", "tool_call>")
		rsp = strings.ReplaceAll(rsp, "<tool>", "<tool_call>")
		rsp = strings.ReplaceAll(rsp, "</tools>", "<tool_call>")
		rsp = strings.ReplaceAll(rsp, "</tool_call>", "<tool_call>")
		//json tool call
		rsp = strings.ReplaceAll(rsp, "```json\n", "<tool_call>")
		rsp = strings.ReplaceAll(rsp, "```tool_call\n", "<tool_call>")
		rsp = strings.ReplaceAll(rsp, "\n```", "<tool_call>")
		rsp = strings.ReplaceAll(rsp, "```\n", "<tool_call>")

		rsp = strings.ReplaceAll(rsp, "```tool_call>", "<tool_call>")

		items := strings.Split(rsp, "<tool_call>")
		//case json only
		if len(items) > 3 {
			items = items[1 : len(items)-1]
		}
		for _, toolcallString := range items {
			if len(toolcallString) < 10 {
				continue
			}
			if i := strings.Index(toolcallString, "{"); i > 0 {
				toolcallString = toolcallString[i:]
			}
			if i := strings.LastIndex(toolcallString, "}"); i > 0 {
				toolcallString = toolcallString[:i+1]
			}

			if toolcall := parseOneToolcall(toolcallString); toolcall != nil {
				toolCalls = append(toolCalls, toolcall)
			}
		}
	}
	return toolCalls
}
func ToolcallParserFileSaver(resp openai.ChatCompletionResponse) (toolCalls []*FunctionCall) {
	var builder strings.Builder
	for _, choice := range resp.Choices {
		builder.WriteString(choice.Message.Content)
	}
	rsp := builder.String()
	// 检查 rsp 是否为空
	if rsp == "" {
		return nil
	}

	const (
		pathMarker    = "\n\nPath:"
		contentMarker = "\nContent:\n"
		eofMarker     = "\nEOF\n"
	)
	items := strings.Split(rsp, pathMarker)
	for _, item := range items[1:] { // 跳过第一个无效部分
		// 截断到 EOF
		eofIndex := strings.Index(item, eofMarker)
		if eofIndex == -1 {
			continue // 如果缺少 EOF，跳过处理
		}
		item = item[:eofIndex]

		// 分割 Path 和 Content
		subItems := strings.Split(item, contentMarker)
		if len(subItems) != 2 {
			continue // 如果格式不符合预期，跳过
		}

		// 去除多余的空白字符
		filename := strings.TrimSpace(subItems[0])
		content := strings.TrimSpace(subItems[1])

		// 构建结果
		toolCalls = append(toolCalls, &FunctionCall{
			Name:      "SaveToFile",
			Arguments: map[string]any{"filename": filename, "content": content},
		})
	}

	return toolCalls
}
func (a *Agent) WithToolcallParser(parse func(resp openai.ChatCompletionResponse) (toolCalls []*FunctionCall)) *Agent {
	if parse == nil {
		parse = ToolcallParserDefault
		a.functioncallParsers = append(a.functioncallParsers, ToolcallParserDefault)
	}
	return a
}

package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/doptime/eloevo/tools"
	"github.com/doptime/eloevo/utils"
	openai "github.com/sashabaranov/go-openai"
	"golang.design/x/clipboard"
)

func (a *Agent) GetResponse(Client *openai.Client, req openai.ChatCompletionRequest) (resp openai.ChatCompletionResponse, err error) {

	// Send the request to the OpenAI API
	if a.msgDeFile != "" {
		resp, err = utils.FileToResponse(a.msgDeFile)
		return resp, err
	}
	ctx := context.Background()
	if a.msgFromCliboard {
		textbytes := clipboard.Read(clipboard.FmtText)
		if len(textbytes) == 0 {
			return resp, fmt.Errorf("no data in clipboard")
		}
		msg := openai.ChatCompletionMessage{
			Role:    "assistant",
			Content: string(textbytes),
		}
		resp = openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: msg,
				},
			},
		}
		return resp, nil
	} else if len(req.Messages) > 0 {
		resp, err = Client.CreateChatCompletion(ctx, req)
	}

	if a.msgToFile != "" {
		if jsonbytes, err := json.Marshal(resp); err == nil {
			tools.SaveToFile(&tools.FileNameString{Filename: a.msgToFile, Content: string(jsonbytes)})
		}
	}
	if a.msgContentToFile != "" && len(resp.Choices) > 0 {
		tools.SaveToFile(&tools.FileNameString{Filename: a.msgContentToFile, Content: resp.Choices[0].Message.Content})
	}
	return resp, err
}

package agents

import (
	"context"
	"fmt"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
)

var AgentResponse = agent.NewAgent(template.Must(template.New("question").Parse(`
直接回复：
1+11=?
`))).WithModels(models.Qwen30BA3) //.WithModels(models.Qwen3B14)
func AgentResponseTest() {
	// 运行Agent
	err := AgentResponse.Call(context.Background(), map[string]any{})
	if err != nil {
		fmt.Println("Error calling AgentResponse:", err)
		return
	}
}

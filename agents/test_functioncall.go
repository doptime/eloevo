package agents

import (
	"fmt"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
)

type testFunctionCall struct {
	Title  string `description:"Title of output"`
	Answer string `description:"The content string to save"`
}

var toolTestFunctionCall = tool.NewTool("AnswerSaver", "Test Tool by calling this functioncall", func(param *testFunctionCall) {
	if param.Title == "" || param.Answer == "" {
		return
	}
	fmt.Println("tool test success! ", param.Title, "Content: ", param.Answer)

})

var AgentFunctioncallTest = agent.NewAgent(template.Must(template.New("question").Parse(`
请调用所提供的TolCall:AnswerSaver 回复：猫有几条腿。
`))).WithModels(models.Qwen30BA3).WithTools(toolTestFunctionCall) //.WithModels(models.Qwen3B14)

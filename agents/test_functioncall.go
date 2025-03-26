package agents

import (
	"fmt"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
)

type testFunctionCall struct {
	Title   string `json:"title,omitempty" description:"Title of output"`
	Content string `json:"content,omitempty" description:"The content string to save"`
}

var toolTestFunctionCall = tool.NewTool("ToolTester", "Test Tool by calling this functioncall", func(param *testFunctionCall) {
	if param.Title == "" || param.Content == "" {
		return
	}
	fmt.Println("tool test success! ", param.Title, "Content: ", param.Content)

})

var AgentFunctioncallTest = agent.NewAgent(template.Must(template.New("question").Parse(`
请调用所提供的FunctionTool,Title 为"Hello", Content 为“Hello World”。
`)), toolTestFunctionCall).WithModel(models.FuseO1)

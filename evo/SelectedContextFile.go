package evo

import (
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
)

type SelectLLMContextFiles struct {
	RelatedFileNames []string  `description:"array, the file names to reserve in the context"`
	Result           *[]string `description:"-"`
}

var PromptSelectContextFiles = template.Must(template.New("SelectLLMContextFiles").Parse(`
{{.ContextFiles}}

## 任务说明
以上是提交给另外一个LLM的context信息。其中的许多文件对于实现system evolution Goals不是必须的。请分析上面列出的所有文件，并筛选出与实现system evolution Goals 相关的文件。以简化context，给后续的LLM作为输入。 

避免过度思考，直接给出与实现目标相关的文件列表。

最后请调用: SelectRelatedFiles 来提交变更. 
`))
var ToolRelatedFileNames = tool.NewTool("SelectRelatedFiles", "从原始context中进一步对实现目标不必要的context，或使用注释来简化context，为后续更大的模型进行必要的context增强预处理。", func(commits *SelectLLMContextFiles) {
	commits.Result = &commits.RelatedFileNames
})

func SelectContextFiles(RawPrompt string, model *models.Model) (NewContextFiles []string) {
	var ReturnLineKept = &[]string{}
	agent.NewAgent(promptSelectContextlines).WithTools(toolSelectContextLines).Call(map[string]any{
		"ContextFiles": RawPrompt,
		"Result":       ReturnLineKept,
		agent.UseModel: model,
	})

	return *ReturnLineKept
}

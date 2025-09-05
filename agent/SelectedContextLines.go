package agent

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
)

type LineRangesToRemove struct {
	Start int `description:"int, 需要移除的范围的起始行号（包含此行）"`
	End   int `description:"int, 需要移除的范围的结束行号（包含此行）"`
}

type SimplifyLLMPrompt struct {
	LineRangeToRemove []LineRangesToRemove `description:"array, the line ranges to remove in the context"`
	Result            *string              `description:"-"`
}

var AgentSelectContextLines = NewAgent().WithTemplate(template.Must(template.New("SimplifyLLMPrompt").Parse(`
{{.LinedContext}}

## 任务说明
以上是提交给另外一个LLM的上下文信息。其中的许多文件对于system evolution Goals求解不是必须的。检查上面列出的所有文件，并筛选出对实现system evolution Goals 相关的文件。以简化上下文，给后续的LLM作为输入。 

最后请调用: SimplifyPromptByRemoveContextWithComment 来提交变更. 
`))).WithModels(models.Oss120b).WithTools(tool.NewTool("SimplifyPromptByRemoveContextWithComment", "从原始上下文中进一步对实现目标不必要的上下文，或使用注释来简化上下文，为后续更大的模型进行必要的上下文增强预处理。", func(commits *SimplifyLLMPrompt) {
	lines := strings.Split(*commits.Result, "\n")
	linemap := make(map[int]string)
	for i, line := range lines {
		linemap[i] = line
	}
	for _, r := range commits.LineRangeToRemove {
		for i := r.Start; i <= r.End && i < len(lines); i++ {
			delete(linemap, i)
		}
	}

	var output bytes.Buffer
	for i := 0; i < len(lines); i++ {
		if line, ok := linemap[i]; ok {
			output.WriteString(line + "\n")
		}
	}
	*commits.Result = output.String()

}))

func AbbreviateContext(llmContext string, model *models.Model) (NewContext string) {
	lines := strings.Split(llmContext, "\n")
	var contextLined strings.Builder
	for i := range lines {
		contextLined.WriteString(fmt.Sprintf("%4d: %s\n", i+1, lines[i]))
	}
	var ReturnLineKept = &llmContext
	AgentSelectContextLines.Call(context.Background(), map[string]any{
		"LinedContext": contextLined.String(),
		"Result":       ReturnLineKept,
		"Model":        model,
	})

	return *ReturnLineKept
}

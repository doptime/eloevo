package agent

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
)

type LineRange struct {
	StartLine int `description:"int, 需要保留的范围的起始行号（包含此行）"`
	EndLine   int `description:"int, 需要保留的范围的结束行号（包含此行）"`
}

type SelectedContextLines struct {
	LineRanges []LineRange  `description:"[]LineRange, the line ranges to keep in the context"`
	Result     *[]LineRange `description:"-"`
}

var AgentSelectContextLines = NewAgent(template.Must(template.New("AgentEvoLearningSolutionLearnByChoose").Parse(`
{{.LinedContext}}


请分析目标意图，理解需要保留的上下文行范围。
然后请调用: SelectedContextLines 删除原始的上下当中不必要的内容，进一步筛选出对于保存必要的上下文行信息。
`))).WithModels(models.Oss120b).WithTools(tool.NewTool("SelectedContextLines", "从原始上下文中选择并保留相关的文本行范围，用于后续处理。", func(commits *SelectedContextLines) {
	*commits.Result = commits.LineRanges
}))

func AbbreviateContext(llmContext string, model *models.Model) (NewContext string) {
	lines := strings.Split(llmContext, "\n")
	var contextLined strings.Builder
	for i := range lines {
		contextLined.WriteString(fmt.Sprintf("%4d: %s\n", i+1, lines[i]))
	}
	var ReturnLineKept *[]LineRange
	AgentSelectContextLines.Call(context.Background(), map[string]any{
		"LinedContext": contextLined.String(),
		"Result":       ReturnLineKept,
	})

	var NewContextSB strings.Builder
	for _, lineRange := range *ReturnLineKept {
		if lineRange.StartLine < 1 {
			continue
		}
		if lineRange.EndLine > len(lines) {
			continue
		}
		if lineRange.StartLine > lineRange.EndLine {
			continue
		}
		for i := lineRange.StartLine - 1; i < lineRange.EndLine; i++ {
			NewContextSB.WriteString(lines[i] + "\n")
		}
	}
	return NewContextSB.String()
}

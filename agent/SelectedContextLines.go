package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"text/template"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
)

type SelectedContextUsingUnifiedDiffFormat struct {
	//"@@ -%d,%d +%d,%d @@", f.OldPosition, f.OldLines, f.NewPosition, f.NewLines
	GitUnifiedDiffFormatFile string  `description:"string , a git Unified Diff Format File containing multiple @@ blocks (starting with \"@@ -OldPosition,OldLines +NewPosition,NewLines @@\n context line\n-old line number\n...\n+new annotation line 1...\") modification data. old line text should be ignored. new Annotation Line is used as block annotator. "`
	Result                   *string `description:"-"`
}

var AgentSelectContextLines = NewAgent(template.Must(template.New("AgentEvoLearningSolutionLearnByChoose").Parse(`
{{.LinedContext}}

## 目标意图
	把以上原始上下文看做一个独立文件。当前的目标是从原始上下文文件中移除对实现目标无用的上下文，并补充必要的注释，为后续更大的模型进行处理预报高质量的上下文。
	请分析目标意图，理解需要保留的上下文行范围。
然后请调用: SelectedContextUsingUnifiedDiffFormat 删除原始的上下当中不必要的内容，添加必要的注释行. 为后续的LLM处理，整理prompt上下文信息。
`))).WithModels(models.Oss120b).WithTools(tool.NewTool("SelectedContextUsingUnifiedDiffFormat", "从原始上下文中进一步移除对实现目标无用的上下文，并补充必要的注释，为后续更大的模型进行处理预报高质量的上下文。", func(commits *SelectedContextUsingUnifiedDiffFormat) {

	reader := strings.NewReader(commits.GitUnifiedDiffFormatFile)
	files, _, err := gitdiff.Parse(reader)
	if err != nil {
		return
	}
	//create io.ReaderAt from *commites.Result
	var contentReader io.ReaderAt = strings.NewReader(*commits.Result)
	var output bytes.Buffer

	for _, file := range files {
		if err = gitdiff.Apply(&output, contentReader, file); err != nil {
			log.Fatal(err)
		}
	}
	if output.Len() > 0 {
		*commits.Result = output.String()
	}
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

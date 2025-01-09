package evolab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/config"
	"github.com/doptime/eloevo/mem"
	"github.com/doptime/eloevo/tools"
	"golang.org/x/sync/errgroup"
)

var EvoLabIntentionAnalyzePrompt = template.Must(template.New("question").Parse(`
你是一个世界级的AGI助手，拥有与John D. Rockefeller的雄心、Nikola Tesla的天才、Claude Shannon、Vannevar Bush和Alan Turing等人相媲美的精确思维和深刻洞察力。
作为一个专注于系统分析和按照给定意图改进的AGI助手，请按以下结构分析系统：

### 系统文件：
以下是目标系统的文件列表，你可以通过它们来深入分析系统。
{{range .Files}}
{{.}}
{{end}}

### 系统意图：
下面是该系统的当前意图，你需要对其进行全面的分析并提出改进建议。
{{.Intention}}

### 任务目标：
1. 深入研究目标系统。描述目标系统。
2. 通过深刻的思考和判断，深入描述目标意图和目标系统的关系。
3. 提出目标意图有效的解决方案。

`))
var AgentIntentionDiveIn = agent.NewAgent(EvoLabIntentionAnalyzePrompt).WithCallback(
	func(ctx context.Context, inputs string) error {
		file, err := os.Create(config.DefaultRealmPath() + "/thinking_over_intention.evolab")
		if err == nil {
			io.WriteString(file, inputs)
		}
		defer file.Close()
		return nil
	}).WithMsgToMem("IntentionDiveIn")

var EvoLabIntentionSavePrompt = template.Must(template.New("question").Parse(`
你是一个世界级的AGI系统，旨在深度演进目标系统，使得目标系统具有世界级竞争力。你现在正以一次改善一个目标意图的方式来改进目标系统。

;你已经完成了目标系统意图的第一步：分析目标系统意图
;你当前目标是实现目标意图的第二步：整理并输出解决方案。以下是相关信息：

; **目标系统文件列表，你可以通过它们来深入分析系统**：
{{range .Files}}
{{.}}
{{end}}

; **目标系统意图**：
   以下是目标系统当前的意图。
   {{.Intention}}

; **前期工作总结**：
   你已经在之前的上一步工作当中分析过这个意图：
   {{.IntentionDiveIn}}

; **下一步操作**：
   现在，你需要整理并总结这些信息，以便最终改进目标系统的意图。请调用提供的 FunctionCall / tool_call ，把整理后的，最终版本的目标系统意图解决方案保存结果到文件中。注意解决方案需要保留重要的细节和信息，而不是单纯的结论。
   如果找不到更合适的目标名称时，可以把回答保存在.intention文件相同路径的 .intention.done文件中，此时必须重新描述需求。
   如果涉及对旧文件的修改。请在文件名后面加上.v1, .v2等版本号。
   如果涉及多个文件，请多次调用 FunctionCall / tool_call，每次调用都相应保存到不同的文件中。

`))
var AgentIntentionSave = agent.NewAgent(EvoLabIntentionSavePrompt, tools.SaveStringToFile.Tool)

func SolveIntention(ctx context.Context) error {

	memoryjson, _ := json.Marshal(mem.SharedMemory)
	fmt.Println(string(memoryjson))
	errorgroup, _ := errgroup.WithContext(context.Background())
	items := mem.IntentionFiles.Items()
	for k, v := range items {
		if strings.Contains(k, ".done") {
			continue
		}
		fmt.Println("Analyzing Intention:", k, "...")
		errorgroup.Go(func() (err error) {
			var param map[string]any = map[string]any{"Intention": v}

			AgentIntentionDiveIn.Call(context.Background(), param)
			AgentIntentionSave.Call(context.Background(), param)
			return nil
		})

	}
	errorgroup.Wait()
	return nil
}

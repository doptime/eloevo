package evolab

import (
	"context"
	"strings"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/agents"
	"github.com/doptime/eloevo/config"
	"github.com/doptime/eloevo/mem"
	"github.com/doptime/eloevo/tools"
	"github.com/doptime/eloevo/utils"
)

var AgentIntentionSolveWithMinimalFiles = agent.NewAgent(template.Must(template.New("question").Parse(`
你是一个专注于改进目系统的AGI助手，请分析系统并修改后的内容到文件：

### 系统意图：
系统意图定义在!system_goals.md文件当中，它包含许多条意图。你的目标是按照 !system_goals.md 文件中的描述 实现当前步骤需要的实现目标。
实现这个目标分为两个步骤：
步骤1. 请调用 名为"PickFileNames" 的 FunctionalCall / tool_call，把下一个目标涉及的上下文相关的文件挑选出来。
步骤2. 这一将在另一个AGI的对话上下文中进行，以便正式修改目标系统的意图。

### 系统文件：
以下是目标系统的文件列表，你可以通过它们来深入分析系统。
{{range .Files}}
{{.}}
{{end}}

注意 "PickFileNames" 函数的 Category 请设置为 ”ContextFiles“
`)), agents.StoreFilenamesToMemory.Tool).
	WithMsgDeFile("IntentionSolved.md").CopyPromptOnly()

func GenQWithMinimalFiles() {
	AgentIntentionSolveWithMinimalFiles.Call(context.Background(), map[string]any{})
	ctxFiles, ok := mem.SharedMemory["ContextFiles"].([]string)
	if !ok {
		return
	}
	var keptFiles string = strings.Join(ctxFiles, "\n")

	originalFiles, ok := mem.SharedMemory["Files"].([]*config.FileData)
	if !ok {
		return
	}
	var leftFiles []*config.FileData
	for _, file := range originalFiles {
		if strings.Contains(keptFiles, file.RealmName()) {
			leftFiles = append(leftFiles, file)
		}
	}
	mem.SharedMemory["Files"] = leftFiles

	var AgentIntentionSolve = agent.NewAgent(template.Must(template.New("question").Parse(`
	你是一个专注于改进目系统的AGI助手，请分析系统并修改后的内容到文件：
	
	### 系统意图：
	系统意图定义在!system_goals.md文件当中，它包含许多条意图。你的目标是按照 !system_goals.md 文件中的描述 依次实现下一个未被标定为已实现的目标。
	
	### 系统文件：
	以下是目标系统的文件列表，你可以通过它们来深入分析系统。
	{{range .Files}}
	{{.}}
	{{end}}
	
	`)), tools.SaveStringToFile.Tool).WithMsgDeFile("IntentionSolved.md").CopyPromptOnly()

	AgentIntentionSolve.Call(context.Background(), map[string]any{})

	utils.PlayBeep()

}

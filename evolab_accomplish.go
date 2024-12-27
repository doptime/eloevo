package evolab

import (
	"text/template"

	"github.com/doptime/evolab/agents"
)

var AgentIntentionAccomplish = agents.NewAgent(template.Must(template.New("question").Parse(`
# 你是一个专注于意图实现的AGI助手
	- 请仔细分析系统文件和系统的意图. 按照系统的意图创建或更新相关的目标文件.
	- 如果需要向目标系统发起必要的反馈。请更新!system_accomplish_feedbacks.md 

---

## 系统意图：
系统意图定义在!system_goals.md文件当中，它包含多条意图。你的目标是按照 !system_goals.md 文件中的描述 以良好的方式实现或改进目标。

---

## 系统文件：
以下是目标系统的文件列表，你可以通过它们来深入分析系统。
{{range .Files}}
{{.}}
{{end}}

---

## 创建或更新相关的目标文件：
    - 输出文件的格式采用系统文件输入的格式， 也就是\n\nPath:...\nContent:\n...\nEOF\n
	- 对修改文件情形的,文件名用.vx,比如v1,v2 ... 作为扩展名，避免不必要的覆盖。
	- 对删除文件情形的,文件名用.del 作为扩展名，避免不必要的覆盖。
	- 对修改意见仅涉及部分文件内容修改的. 请确保修改后的文件内容的完整性，需要完整保留除了修改处的其余部分.避免意外丢失内容.

`))).WithToolcallParser(agents.ToolcallParserFileSaver)

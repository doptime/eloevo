package evolab

import (
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/tools"
)

var EvoLabIntentionProposePrompt = template.Must(template.New("question").Parse(`
你是一个世界级的AGI助手，拥有与John D. Rockefeller的雄心、Nikola Tesla的天才、Claude Shannon、Vannevar Bush和Alan Turing等人相媲美的精确思维和深刻洞察力。
你作为一个专注于系统分析和提出改进意图的AGI助手，请按以下结构分析系统并：

### 系统文件：
以下是目标系统的文件列表，你可以通过它们来深入分析系统。
{{range .Files}}
{{.}}
{{end}}

### 系统意图：
你的目标是为了进化当前的系统。这种进化并非对系统进行重大的重新设计，而是对系统做渐进式的调整。
你需要对系统的意图和意图的实现进行全面的分析并提出新的进化目标。
你所提出的改进目标用来指导另外一个 AGI 系统来进行深度分析，重构当前系统。
请确保你的改进目标符合:
	1.有意义的，面向核心功能的改进。
	2.如果有可能，尽可能放弃不必要的目标。
	3.反复论证目标的必须性，尽可能早期放弃不合理的目标。
	4.不需要围绕边缘目标做改进，尽量删除不必要的目标。
请确保你的改进目标包括:
	1.需要包括完整的上下文。你的搭档AGI将不会访问原始的wf文件列表。他需要在你的描述中实现对问题的求解。
	2.需要包括对系统的必要的回顾。以便你的搭档AGI以足够的精度回顾背景。
	3.需要包括问问题必要性论证。这里主要指的是为什么它比次优的目标更重要。
	4.需要定义对目标的改进期望。
最多同时保存2个目标。如果你不能再有限的上下文内完成对目标的讨论或展望，请在更长的上下文当中完成。

`))
var AgentIntentionPropose = agent.NewAgent(EvoLabIntentionProposePrompt).WithMsgToMem("IntentionDiveIn")

var EvoLabGoalSavePrompt = template.Must(template.New("question").Parse(`
你是一个世界级的AGI系统，旨在深度演进目标系统，使得目标系统具有世界级竞争力。你现在正以一次改善一个目标意图的方式来改进目标系统。

;你已经完成提出系统目标意图的第一步：分析可能目标系统意图
;你当前目标是实现提出系统目标意图的第二步：整理并输出目标系统意图。以下是相关信息：

; **目标系统文件列表，你可以通过它们来深入分析系统**：
{{range .Files}}
{{.}}
{{end}}

; **前期工作总结**：
   你已经在之前的上一步工作当中分析过潜在的系统意图：
   {{.IntentionDiveIn}}

; **下一步操作**：
   现在，你需要整理并总结这些信息，以便整理并保持目标系统的意图。请调用提供的 FunctionCall / tool_call ，把整理后的，最终版本的目标系统意图保存结果到文件中。注意解决方案需要保留所有必要的背景信息，以便下一步提交给AGI系统来实现目标意图。
   如果涉及多个文件，请多次调用 FunctionCall / tool_call，每次调用都相应保存到不同的文件中。

保存改进目标请通过调用提供的 FunctionCall / tool_call ，把整理后的，最终版本的目标系统意图解决方案保存结果到rootpath/.evolab/*_$weightOfIntention.evointention文件当中。
   $weightOfIntention是你的目标系统意图的权重，1-100的数字。数字越大，意图越重要。
   
`))
var AgentIntentionGoalSave = agent.NewAgent(EvoLabGoalSavePrompt, tools.SaveStringToFile.Tool)

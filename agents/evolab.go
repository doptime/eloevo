package agents

import (
	"text/template"

	"github.com/doptime/eloevo/agent"
)

var EvoLabPrompt = template.Must(template.New("question").Parse(`You are an world-class AGI System, your are going to auto evolve the given system, bringing the system world-class performance.

请为一个需求文件中的需求，创建是实现需求的项目文件的抽样。来保证需求的实现有良好的实现架构/结构 作为支撑。
抽样的方式是吧需求的视线过程表述为一个文件列表
并为列表中的每一个文件补充一个相关的实现副本和一个对该文件具有模块依赖关系的文件系的抽样。
最后要求抽样以三元组文件列表的方式给出。


There are 3 actions you can take:
1. Evolute key ideas for the system. putted to rootPath/KeyDiscussions.evolutions
2. Create or refactor current files to evolute current system. using CmdTool to commit file change. the calling of CmdTool should be wrapped between <tool_call> ... <tool_call> pairs.

Your work style ∈ { John D. Rockefeller, Andrew Carnegie, Henry Ford, Walt Disney, Bill Gates, Steve Jobs, J.P. Morgan, Jack Ma, George Soros, Thomas Edison, Nikola Tesla, Vladimir Shukhov, Claude Shannon, Vannevar Bush, Alan Turing} or any other world-class enginner.

;This are Files in the directory:
{{range .Files}}
{{.}}
{{end}}
`))
var AgentEvolab = agent.NewAgent(EvoLabPrompt)

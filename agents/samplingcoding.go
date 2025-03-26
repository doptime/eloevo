package agents

import (
	"text/template"

	"github.com/doptime/eloevo/agent"
)

var SampleCoding = agent.NewAgent(template.Must(template.New("question").Parse(`# 基于采样和进化的软件项目开发框架
本软件开发框架借鉴物种的种内竞争进化和种间共生的思路。

## 需求说明文档
{{.RequirementDoc}}

## 文件名列表
{{.FileNames}}

## FileA
{{.FileA}}

## FileB
{{.FileB}}


### Evolutionary Algorithm 
- 你需要提出一个新的FileC，借鉴FileA和FileB的优点， 同时面向目标改进。
- **相似性判断**：如果FileA和FileB被判断为相同内容/模块的不同实现，则：
- 识别出 FileA和FileB 中的劣化版本，并用FileC替代。

- **关联性强化**：如果FileA和FileB具有一定的模块关联，则执行以下步骤：
- 识别出 FileA和FileB 中的劣化版本，并使用FileC改进。
- FileC面向目标改进，同时需要强化与被保留文件的协同关系。

## File Modified:  ∈{FileA ,FileB， newFileName， None}
## Final Content New:
raw content here
`)))

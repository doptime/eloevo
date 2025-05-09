package scrum

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/devops/projects"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
)

var scumProject = `
系统愿景:实现AI时代，无人机作为基础物流平台。确保方案和应用的极致的简单、可靠、低成本、高效用。
你的核心目标是基于第一性原理的工程学实现，构建一个在无人机平台和机器人应用领域具有高价值、高可行性的项目模块矩阵，这些项目应能在未来的世界中产生最大的联合商业效用和社会效用。

## 涉及的目标行业包括：
- AI-Driven Bobotic Development
- Robotic As a Service
- Drone Technology & Solution
- Suppy Chain & Drone & Logistics Technology
    -如从农业产地直供最终零售点
- Sustainable Packaging Technology
- Sustainable Transportation Infrastructure

## 部分愿景:
- 借助外接电源或超高速放电电池，垂直起飞的固定翼无人机
- 它是一个非常便利的载具平台。可以提供各种机器人的投送和收回服务
- 由于极高的滑翔比。它的物流成本只有汽车的1/10和船运的1/2. 可以在全球内完成有中继的长途运输
- 它可以借助地形和动态风向变化，实现能量的节约。
- 联合多机器人和多飞机。送外卖，入户医疗检查。在户外部署就餐，住宿，岗哨，它能做很多。
- 在未来的世界中，基于无人机平台的全球即时物流和资源分配系统，是最重要的基础设施。
- AI 驱动的自主机器人团队，能够高效协作完成复杂任务。

## 本系统采用迭代方式来渐进得完成愿景，当前的迭代会持续数十万次，直至最终目标实现。每一轮的迭代中，请通过一系列的Funtioncall 调用，来完善或改进方案的实现。
	整个解决方案被建模为顶点和边的图。其中的边为超边(超边是连接多个顶点的边)可以连接两个或两个以上的模块节点。 现有的方案由两类节点构成:
	1) 超边节点
		超边节点就是图的边。但由于可以连接多个顶点，所以是超边。作为一种图结构处理技巧，我们把超边看做是非普通节点也叫超边节点。
		超边节点用于实现系统的架构设计，并且通过模块节点用来进一步驱动系统的实现。通过以下维度，来实现系统的架构设计:
		 - 解决业务契合度	架构是否真正解决业务痛点？用业务 KPI / 用户场景 / 收益模型倒推技术方案，持续校验“为什么做”。
		 - 技术可行性	方案能否落地、运维、扩展？	技术验证（PoC）、性能基准测试、与现有技术栈/团队能力匹配度。
		 - 成本–收益比	投入与产出是否平衡？	固定成本（硬件/许可）、可变成本（云资源）、人力/维护，结合收益或风险降低进行 ROI 评估。
		 - 风险管理	有哪些风险？怎样缓解？	技术 / 合规 / 安全 / 供应链风险识别 → 减缓措施 → 残余风险可接受性。
		 - 治理与可持续性	架构能否迭代、治理？	模块化、接口契约、Observability、版本策略、技术债务控制、文档化。
		 - 交付节奏	如何在有限时间内持续交付价值？	与敏捷/DevOps结合的迭代式架构；“Just-Enough Architecture” 概念。
		 - 沟通协作	是否与干系方充分参与并达成共识？	理解并响应其它人类用户的需求/专家的反馈/其它AI的评审/Scrum Backlog。
		超边应显式设置SuperEdge=true。
	2) 模块节点
		模块节点是直接给出解决方案的节点
		模块只能通过超边和其它的模块节点完成耦合。也就是通过实现超边节点约束来定义自身
`

type SessionItem struct {
	Id              string                                 `description:"The id of the super node"`
	Session         string                                 `description:"Session string, with format like '1' or '1.1' or '1.1.1' or '2.1'"`
	SessionAnnotate string                                 `description:"SessionAnnotate string, the idea beyond the assign of session number"`
	SuperNodesMap   map[string]*projects.SolutionGraphNode `description:"-"`
}

var AgentSessionArrangement = agent.NewAgent(template.Must(template.New("AgentBusinessPlansDrone").Parse(`
{{.scumProject}}

{{.Supernodes}}

你需要将这些超边节点添加章节编号,以便节点以层次化并且逻辑合理化的方式组织在一起。

请对每一个节点，分别调用一次所提供的FunctionCall:AnswerSaver 来保存章节编号
`))).WithModels(models.Qwen3B14).WithTools(tool.NewTool("AnswerSaver", "Save session string for a given node", func(item *SessionItem) {
	// 这里可以实现保存章节编号的逻辑
	if item.Id == "" || item.Session == "" {
		return
	}
	rawItem := item.SuperNodesMap[item.Id]
	if rawItem == nil {
		return
	}
	rawItem.ChapterSession = item.Session + "\n" + item.SessionAnnotate
	projects.KeyBusinessDronebot.HSet(item.Id, rawItem)

})) //.WithModels(models.Qwen3B14)

func AgentSessionArrangementCall() {

	businessPlans, _ := projects.KeyBusinessDronebot.HGetAll()
	SuperNodes := make([]*projects.SolutionGraphNode, 0)
	for _, item := range businessPlans {
		if item.SuperEdge {
			SuperNodes = append(SuperNodes, item)
		}
	}
	slices.SortFunc(SuperNodes, func(a, b *projects.SolutionGraphNode) int {
		return strings.Compare(a.Item, b.Item)
	})
	SuperNodesList := projects.SolutionGraphNodeList(SuperNodes).Uniq()
	err := AgentSessionArrangement.Call(context.Background(), map[string]any{
		"scumProject":   scumProject,
		"Supernodes":    SuperNodesList,
		"SuperNodesMap": businessPlans,
	})
	if err != nil {
		fmt.Println("Error calling AgentSessionArrangement:", err)
		return
	}

}

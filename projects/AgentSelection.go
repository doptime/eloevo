package projects

import (
	"fmt"
	"slices"
	"text/template"

	// "github.com/yourbasic/graph"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/scrum"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

type ApplySelectedAgent struct {
	WhatTodoInFollowingIter string                                      `description:"What is key focus to do in following iter, string"`
	MemoToTheNextIter       string                                      `description:"memo should pass to next iteration, string"`
	BackgroundNodesToPass   []string                                    `description:"background solution nodes (Id) should pass to next iteration, array of string"`
	SuperEdges              string                                      `description:"-"`
	AllItems                map[string]*SolutionGraphNode               `description:"-"`
	Backlogs                []*scrum.Backlog                            `description:"-"`
	ProductGoal             string                                      `description:"-"`
	HashKey                 redisdb.HashKey[string, *SolutionGraphNode] `description:"-"`
	ThisAgent               *agent.Agent                                `description:"-"`
}

var AgentApplySelectedAgent = agent.NewAgent(template.Must(template.New("AgentAutoSelect").Parse(`

## 本系统采用迭代方式来渐进实现系统的自动化构建，当前的迭代会持续数千次，直至最终目标实现。每一轮的迭代中，请通过一系列的Funtioncall 调用，来完善或改进方案的实现。
	整个解决方案被建模为顶点和边的图。其中的边为超边(超边是连接多个顶点的边)可以连接两个或两个以上的模块节点。 现有的方案由两类节点构成:
	1) 超边节点
		超边节点就是图的边。但由于可以连接多个顶点，所以是超边。作为一种图结构处理技巧，我们把超边看做是非普通节点也叫超边节点。
		超边节点用于实现系统的架构设计，并且通过模块节点用来进一步驱动系统的实现。通过以下章节/维度，来实现系统的架构设计:
		 - 维度:解决业务契合度	核心问题:架构是否真正解决业务痛点？	要点:用业务 KPI / 用户场景 / 收益模型倒推技术方案，持续校验“为什么做”。
		 - 维度:技术可行性	核心问题:方案能否落地、运维、扩展？	要点:技术验证（PoC）、性能基准测试、与现有技术栈/团队能力匹配度。
		 - 维度:成本–收益比	核心问题:投入与产出是否平衡？	要点:固定成本（硬件/许可）、可变成本（云资源）、人力/维护，结合收益或风险降低进行 ROI 评估。
		 - 维度:风险管理	核心问题:有哪些风险？怎样缓解？	要点:技术 / 合规 / 安全 / 供应链风险识别 → 减缓措施 → 残余风险可接受性。
		 - 维度:治理与可持续性	核心问题:架构能否迭代、治理？	要点:模块化、接口契约、Observability、版本策略、技术债务控制、文档化。
		 - 维度:交付节奏	核心问题:如何在有限时间内持续交付价值？	要点:与敏捷/DevOps结合的迭代式架构；“Just-Enough Architecture” 概念。
		 - 维度:沟通协作	核心问题:是否与干系方充分参与并达成共识？	要点:理解并响应其它人类用户的需求/专家的反馈/其它AI的评审/Scrum Backlog。
		超边应显式设置SuperEdge=true。
	2) 解决方案节点/ 模块节点
		模块节点是直接给出解决方案的节点
		模块节点应显式设置SuperEdge=false。
		模块只能通过超边和其它的模块节点完成耦合。也就是模块的耦合应该被显式提出为超边节点。解决方案节点通过实现超边节点约束来完成定义。
	
	系统的改进架构: 敏捷开发/scrum 与 增量建构/修改
	本系统采用敏捷开发的思路驱动。每一次的敏捷开发是一次Scrum。开发内容通过 增量建构文档，也就是scrum.Backlog 来驱动。


## 当前的系统状态:

这是系统的product goal:
{{.ProductGoal}}

这是系统的当前的超边节点:
{{.SuperEdges}}

这是当前的Scrum.Backlog:
{{range $i, $backlog := .Backlogs}}
{{$backlog}} 
{{end}}


## 这是当前解决方案摘要(非超标节点):
{{.SolutionSummary}}

这是系统的全部可用的LLM Agents,这些Agent 用于修改，删除，新增超边节点或者是普通节点:

- AgentName: BusinessAndGoalClarification
业务与目标澄清:
• 模拟干系人访谈、业务流程图、痛点 / 成长目标 / 约束收集
• 产出内容: 业务目标清单、优先级、可量化 Key Performance Indicator

- AgentName: DemandSorting
需求梳理（功能 & 非功能）
• User Stories、用例、Domain Event
• non functional requirements：性能、伸缩性、可用性、合规、安全、可观察性
• 产出内容: 需求规格说明（functional requirements documents + Non functional requirements documents）


- AgentName: DeleteNodes
删除超边节点
• 在现有模糊不清的条目当中，有哪些条目的存在可能不是必须而应当消融删除的？删除节点，以便维持系统的简洁性和可维护性。
• 产出内容: 对超边节点和普通节点（模块节点）的删除操作

{{if false}}

- AgentName: DemandSorting
需求梳理（功能 & 非功能）
• User Stories、用例、Domain Event
• NFR：性能、伸缩性、可用性、合规、安全、可观察性
• 产出内容: 需求规格说明（FRD+NFRD）

- AgentName: CurrentStateBaseline 
当前状态基线（As-Is）
• 现有系统拓扑、依赖、痛点、成本
• 产出内容: 现状架构图、问题清单

- AgentName: KeyScenariosAndCapacityEstimation
关键场景与容量预估
• 选 3–5 个最关键的业务/技术场景做容量 & SLA 预估
• 产出内容: 容量模型、流量曲线、SLA & SLO


- AgentName: ArchitecturePrinciplesAndDecisionFramework
架构原则与决策框架
• 定义指导原则（如：云优先、事件驱动、松耦合、开放标准 …）
• 确认决策流程（ADR、原则打分、专家评审）


- AgentName: AdvancedSolutionDesign
高阶方案设计（To-Be View）
• 架构图（C4 Model、分层、组件、数据流、部署视图）
• Technology Radar：候选技术、优劣、约束
• 产出内容: 多视图架构草稿 & 备选方案


- AgentName: DeepVerification
深度验证（PoC / Spike）
• 对关键技术/性能/安全点做快速 PoC
• 产出内容: PoC 报告、基准数据、Go/No Go 决策


- AgentName: DetailedDesignAndTradeOff
详细设计 & Trade-off
• 接口契约、数据模型、API、时序图
• 容错、幂等、Observability、CI/CD 流水线
• 产出内容: 详细设计文档（DDD、数据库 ERD、API 定义 …）


- AgentName: RiskAssessmentAndComplianceAudit
风险评估 & 合规审计
• 威胁建模、隐私评估、软件许可、成本灵敏度分析
• 产出内容: 风险登记册、缓解计划


- AgentName: ReviewAndConsensus
评审 & 共识
• 架构评审会（内部 + 外部专家）
• 记录反馈、确认版本 & 里程碑

- AgentName: ContinuousImprovement
持续治理 & 迭代
• ADR/Changelog 持续更新、Observability 指标监控
• 技术债务看板、能力培训、定期架构回顾

- AgentName: MindStorming
头脑风暴
• 进行头脑风暴，提出新的想法和解决方案
• 产出内容: 新的想法列表
{{end}}

- AgentName:  AgentCreateNode
创建一个解决方案节点/模块节点 
• 依照需要解决的主题; 和相关约束条件创建缺失的模块节点
• 产出内容: 新的解决方案节点或者是节点的修改

- AgentName:  AgentRefineNode
精炼一个模块节点 
• 输入:需要解决的主题; 需要考虑的约束条件, 已有的模块节点等
• 产出内容: 新的解决方案节点或者是节点的修改

- AgentName:  PathnameRefine
重新整理模块节点的路径名
• 产出内容: 新的解决方案节点名称

## To Do:
	仅进行深度分析，以便从上面的UserAgents 中选择一个或者是多个Agent 来完成Backlog中的任务。
	不是所有的Agent 都适合当前的任务。不同的任务需要的实现精度差异很大。仅确保以Just-Enough 的方式来完成当前的任务。
	最后通过 FunctionCall:ApplySelectedAgent 来启用选定的Agent来进一步迭代系统。



`))).WithToolCallMutextRun().WithTools(tool.NewTool("ApplySelectedAgent", "choose a proper agent for the next step", func(param *ApplySelectedAgent) {
	solution := ""
	if len(param.BackgroundNodesToPass) > 0 {
		nodes := lo.Filter(lo.Values(param.AllItems), func(v *SolutionGraphNode, _ int) bool {
			return lo.Contains(param.BackgroundNodesToPass, v.Id)
		})
		solution = SolutionGraphNodeList(nodes).Solution().PathnameSorted().FullView()
	}

	err := param.ThisAgent.Call(map[string]interface{}{
		"WhatTodoInFollowingIter": param.WhatTodoInFollowingIter,
		"MemoToTheNextIter":       param.MemoToTheNextIter,
		"RelativeSolutionNodes":   solution,
		"ProductGoal":             param.ProductGoal,
		"AllItems":                param.AllItems,
		"Backlogs":                param.Backlogs,
		"SuperEdges":              param.SuperEdges,
		"HashKey":                 param.HashKey,
	})
	if err != nil {
		fmt.Printf("Error calling AgentGenSuperEdge: %v\n", err)
	}
})).WithModels(models.Qwen3B32Thinking)

var KeyBacklog = redisdb.NewListKey[*scrum.Backlog](redisdb.Opt.HttpVisit().Key("AntiAgingBacklog"))

var KeyAntiAging = redisdb.NewHashKey[string, *SolutionGraphNode](redisdb.Opt.HttpVisit().Key("AntiAgingNodes"))

func AgentSelectAndExecute() {
	const MaxThreads = 1
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)
	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		backlogs, _ := KeyBacklog.LRange(0, -1)
		allNodes, _ := KeyAntiAging.HGetAll()
		restartNeeded := false
		for _, item := range allNodes {
			if len(item.Id) > 4 {
				item.Id = utils.ID(item.BulletDescription, 4)
				KeyAntiAging.HSet(item.Id, item)
				restartNeeded = true
			}
			KeyAntiAging.HSet(item.Id, item)
		}
		if restartNeeded {
			continue
		}

		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		newNodes := lo.Filter(lo.Values(allNodes), func(v *SolutionGraphNode, _ int) bool { return !v.Locked })
		newNodesSorted := slices.Clone(newNodes)
		if len(newNodesSorted) > 0 {
			fmt.Printf("NewNodes Best: %v\n", newNodesSorted[len(newNodesSorted)-1].String())
		}

		var constraints SolutionGraphNodeList = SolutionGraphNodeList(lo.Values(allNodes)).SuerEdge().PathnameSorted()
		constraintsStr_ := SolutionGraphNodeList(constraints).Uniq().FullView()
		SolutionSummary := SolutionGraphNodeList(lo.Values(allNodes)).Solution().PathnameSorted().SummaryView()

		go func(backlogs []*scrum.Backlog, SuperEdges string, AllItems map[string]*SolutionGraphNode, newNodes SolutionGraphNodeList) {
			defer func() { <-MaxThreadsSemaphore }()
			err := AgentApplySelectedAgent.WithModels(models.Qwen3B32Thinking).Call(map[string]any{
				"Backlogs":        backlogs,
				"SuperEdges":      SuperEdges,
				"ProductGoal":     scrum.ProductGoalAntiAging,
				"HashKey":         KeyAntiAging,
				"SolutionSummary": SolutionSummary,
			})
			if err != nil {
				fmt.Printf("Agent call failed: %v\n", err)
			}
		}(backlogs, constraintsStr_, allNodes, newNodesSorted)
	}
	// Wait for all the goroutines to finish)
	for i := 0; i < MaxThreads; i++ {
		MaxThreadsSemaphore <- struct{}{}
	}

}

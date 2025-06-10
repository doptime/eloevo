package projects

import (
	"strings"
	"text/template"

	// "github.com/yourbasic/graph"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/samber/lo"
)

var ForbiddenWords = []string{"区块链", "量子", "氢燃料", "纠缠", "quantum", "blockchain", "hydrogen", "entanglement", "co2", "carbon sequestration"}

var AgentGenSuperEdge = agent.NewAgent(template.Must(template.New("AgentGenSuperEdge").Parse(`
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

这是本次迭代的目标:
{{.WhatTodoInFollowingIter}}

这是来自上一次迭代的反馈:
{{.MemoToTheNextIter}}

这是系统的当前的超边节点:
{{.SuperEdges}}

{{if lt (len .RelativeSolutionNodes) 0 }}
这是当前重点关注的部分解决方案节点:
{{.RelativeSolutionNodes}}
{{end}}

这是当前的Scrum.Backlog:
{{range $i, $backlog := .Backlogs}}
{{$backlog}} 
{{end}}

现在你的角色是通过提出、修改、删除超边节点来改进系统的架构设计。请思考并提出一个或多个超边节点来改进系统的架构设计。
{{if eq .Task "BusinessAndGoalClarification"}}
本次目标是 - 业务与目标澄清:
	• 进行模拟的干系人访谈、业务流程图、痛点 / 成长目标 / 约束收集
	• 产出内容: 业务目标清单、优先级、可量化 KPI
{{else if eq .Task "CurrentStateBaseline"}}
本次目标是 - 更新当前状态基线（As-Is）
• 现有系统拓扑、依赖、痛点、成本
• 产出内容: 现状架构图、问题清单
{{else if eq .Task "DemandSorting"}}
本次目标是 - 需求梳理（功能 & 非功能）
• User Stories、用例、Domain Event
• NFR：性能、伸缩性、可用性、合规、安全、可观察性
• 产出内容: 需求规格说明（FRD+NFRD）
{{else if eq .Task "KeyScenariosAndCapacityEstimation"}}
本次目标是 - 关键场景与容量预估
• 选 3–5 个最关键的业务/技术场景做容量 & SLA 预估
• 产出内容: 容量模型、流量曲线、SLA & SLO
{{else if eq .Task "ArchitecturePrinciplesAndDecisionFramework"}}
本次目标是 - 架构原则与决策框架
• 定义指导原则（如：云优先、事件驱动、松耦合、开放标准 …）
• 确认决策流程（ADR、原则打分、专家评审）
{{else if eq .Task "AdvancedSolutionDesign"}}
本次目标是 - 高阶方案设计（To-Be View）
• 架构图（C4 Model、分层、组件、数据流、部署视图）
• Technology Radar：候选技术、优劣、约束
• 产出内容: 多视图架构草稿 & 备选方案
{{else if eq .Task "DeepVerification"}}
本次目标是 - 深度验证（PoC / Spike）
• 对关键技术/性能/安全点做快速 PoC
• 产出内容: PoC 报告、基准数据、Go/No Go 决策
• 产出对象: 以超边节点增量的形式，包括修改，删除，新增
{{else if eq .Task "DetailedDesignAndTradeOff"}}
本次目标是 - 详细设计 & Trade-off
• 接口契约、数据模型、API、时序图
• 容错、幂等、Observability、CI/CD 流水线
• 产出内容: 详细设计文档（DDD、数据库 ERD、API 定义 …）
{{else if eq .Task "RiskAssessmentAndComplianceAudit"}}
本次目标是 - 风险评估 & 合规审计
• 威胁建模、隐私评估、软件许可、成本灵敏度分析
• 产出内容: 风险登记册、缓解计划
{{else if eq .Task "ReviewAndConsensus"}}
本次目标是 - 评审 & 共识
• 进行模拟的架构评审会（内部 + 外部专家）
• 记录反馈、确认版本 & 里程碑
{{else if eq .Task "ContinuousImprovement"}}
本次目标是 - 持续治理 & 迭代
• ADR/Changelog 持续更新、Observability 指标监控
• 技术债务看板、能力培训、定期架构回顾
• 产出对象: 以超边节点增量的形式，包括修改，删除，新增

{{else if eq .Task "DeleteNodes"}}
本次目标是 - 删除不必要超边节点
• 深度分析思考如何才能确保Just-Enough Architecture / Module
• 哪些是不符合第一性原理的超边节点
• 其它可以跳过或简化的架构约束
• 在现有模糊不清的条目当中，有哪些条目的存在可能不是必须而应当消融删除的
• 如何才能强化系统的简洁性和可维护性
• 产出对象: 节点增量，包括修改，删除，新增

{{else if eq .Task "MindStorming"}}
本次目标是 - 思维风暴
• 通过思维风暴，提出新的超边节点或模块节点
• 产出对象: 以超边节点增量的形式，包括修改，删除，新增

{{else if eq .Task "AgentCreateNode"}}
本次目标是 - 创建新的解决方案节点/模块节点
讨论并提出生成新的模块节点 框架
• 产出对象: 以超边节点增量的形式，包括修改，删除，新增

{{else if eq .Task "AgentRefineNode"}}
本次目标是 - 方案节点的改进

{{else if eq .Task "PathnameRefine"}}
本次目标是 - 路径名的改进
讨论并提出新的节点路径名。
路径名称应该以数字开头像是1-1, 1-2-1 这样像是章节编号的形式。以便于它们以正确的顺序被显示。

{{end}}

## To Do:
	请先进行深度思考，并提出合理的Refine 方案。
	最后通过一次或多次调用 FunctionCall:SolutionSuperEdgeRefine 来保存方案改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Id,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)通过指定Id,修改Importance = -1 来删除无效节点

`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionSuperEdgeRefine", "create/modify/remove supernodes", func(newItem *SolutionGraphNode) {
	newItem.BulletDescription = strings.TrimSpace(newItem.BulletDescription)
	newItem.Id = lo.Ternary(strings.Contains(strings.ToLower(newItem.Id), "new"), "", newItem.Id)
	var oItem *SolutionGraphNode = nil
	if newItem.Id != "" {
		oItem, _ = newItem.HashKey.HGet(newItem.Id)
		// any futher modification to the item should be neglected if the item is locked
	}
	//skip node that is locked
	if oItem != nil && oItem.Locked {
		return
	}
	if newItem.Id = utils.ID(newItem.BulletDescription, 4); oItem != nil {
		newItem.Id = oItem.Id
		oItem.Importance = newItem.Importance
		oItem.Priority = newItem.Priority
		oItem.Content, _ = lo.Coalesce(newItem.Content, oItem.Content)
		oItem.BulletDescription, _ = lo.Coalesce(newItem.BulletDescription, oItem.BulletDescription)

	}
	if isNewModel := oItem == nil; isNewModel {
		if len(newItem.BulletDescription) > 0 && !utils.HasForbiddenWords(newItem.BulletDescription, ForbiddenWords) {
			newItem.HashKey.HSet(newItem.Id, newItem)
		}
		return
	}
	newItem.HashKey.HSet(oItem.Id, oItem)

})).WithModels(models.Qwen3B32Thinking)

// var KeyBacklog = redisdb.NewListKey[*scrum.Backlog](redisdb.Opt.HttpVisit().Key("AntiAgingBacklog"))

// var KeyAntiAging = redisdb.NewHashKey[string, *SolutionGraphNode](redisdb.Opt.HttpVisit().Key("AntiAgingNodes"))

// func AgentSelect() {
// 	const MaxThreads = 1
// 	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)
// 	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
// 		backlogs, _ := KeyBacklog.LRange(0, -1)
// 		businessPlans, _ = KeyAntiAging.HGetAll()

// 		AgentBusinessPlans.ShareMemoryUpdate("AllItems", businessPlans)
// 		//commit remove where remove >= 5
// 		for _, key := range lo.Keys(businessPlans) {
// 			item := businessPlans[key]

// 			if utils.HasForbiddenWords(strings.ToLower(item.BulletDescription), ForbiddenWords) || item.Importance < 0 {
// 				KeyAntiAging.HDel(item.Id)
// 				delete(businessPlans, item.Id)
// 				if err := milvusCollection.Remove(item); err != nil {
// 					fmt.Println("Error removing item from Milvus:", err)
// 				}
// 			}

// 		}

// 		//Upsert to milvus
// 		var milvusInserts []*SolutionGraphNode
// 		for _, item := range lo.Values(businessPlans) {
// 			if item.BulletDescription!= "" && len(item.Embed()) != 1024 {
// 				embed, err := utils.GetEmbedding(item.BulletDescription)
// 				if err == nil {
// 					item.Embed(embed)
// 					milvusInserts = append(milvusInserts, item)
// 				}
// 			}
// 		}
// 		if len(milvusInserts) > 0 {
// 			milvusCollection.Upsert(lo.Uniq(milvusInserts)...)
// 		}

// 		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
// 		newNodes := lo.Filter(lo.Values(businessPlans), func(v *SolutionGraphNode, _ int) bool { return !v.Locked })
// 		newNodesSorted := slices.Clone(newNodes)
// 		slices.SortFunc(newNodesSorted, func(i, j *SolutionGraphNode) int {
// 			return -int(i.Elo - j.Elo)
// 		})
// 		if len(newNodesSorted) > 0 {
// 			fmt.Printf("NewNodes Best: %v\n", newNodesSorted[len(newNodesSorted)-1].String())
// 		}

// 		var constraints SolutionGraphNodeList = SolutionGraphNodeList(lo.Values(businessPlans)).SuerEdge().ChapterSorted()
// 		constraintsStr_ := SolutionGraphNodeList(constraints).Uniq().LockedOnly().String()

// 		go func(backlogs []*scrum.Backlog, SuperEdges string, AllItems map[string]*SolutionGraphNode, newNodes SolutionGraphNodeList) {
// 			defer func() { <-MaxThreadsSemaphore }()
// 			err := AgentApplySelectedAgent.WithTools(ToolDroneBatchEloResults).WithModels(models.Qwen3B32Thinking).Call(context.Background(), map[string]any{
// 				"Backlogs":    backlogs,
// 				"SuperEdges":  SuperEdges,
// 				"ProductGoal": scrum.ProductGoalAntiAging,
// 			})
// 			if err != nil {
// 				fmt.Printf("Agent call failed: %v\n", err)
// 			}
// 		}(backlogs, constraintsStr_, businessPlans, newNodesSorted)
// 	}
// 	// Wait for all the goroutines to finish)
// 	for i := 0; i < MaxThreads; i++ {
// 		MaxThreadsSemaphore <- struct{}{}
// 	}

// }

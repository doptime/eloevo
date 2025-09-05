package projects

import (
	"context"
	"fmt"
	"slices"
	"strings"
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

var AgentEvoLearningCallback = agent.NewAgent().WithTemplate(template.Must(template.New("AgentEvoLearningCallback").Parse(`
## 本系统采用迭代方式来渐进实现系统的自动化构建，当前的迭代会持续数千次，直至最终目标实现。每一轮的迭代中，请通过一系列的Funtioncall 调用，来完善或改进方案的实现。
	整个解决方案被建模为顶点和边的图。其中的边为超边(超边是连接多个顶点的边)可以连接两个或两个以上的模块节点。 现有的方案由两类节点构成:
	1) 超边节点
		超边节点是实现解决方案的辅助。
		超边应显式设置SuperEdge=true。
	2) 解决方案节点/ 模块节点
		模块节点是解决方案的模块（文件）。
		模块节点应显式设置SuperEdge=false。
		模块只能通过超边和其它的模块节点完成耦合。也就是模块的耦合应该被显式提出为超边节点。解决方案节点通过实现超边节点约束来完成定义。
	
	系统的改进架构:
	本系统采用敏捷开发的思路驱动。每一次的迭代是一次敏捷开发的Scrum。开发内容优先使用scrum.Backlog 来驱动。


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


## To Do:
	请按照本次迭代的目标，行深度思考，并提出有效的Refine 方案。
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

var AgentEvoLearning = agent.NewAgent().WithTemplate(template.Must(template.New("AgentEvoLearning").Parse(`
## 本系统采用迭代方式来渐进实现系统的自动化构建，当前的迭代会持续数千次，直至最终目标实现。每一轮的迭代中，请通过一系列的Funtioncall 调用，来完善或改进方案的实现。
	整个解决方案被建模为顶点和边的图。其中的边为超边(超边是连接多个顶点的边)可以连接两个或两个以上的模块节点。 现有的方案由两类节点构成:
	1) 超边节点
		超边节点是实现解决方案的辅助。
		超边应显式设置SuperEdge=true。
	2) 解决方案节点/ 模块节点
		模块节点是解决方案的模块（文件）。
		模块节点应显式设置SuperEdge=false。
		模块只能通过超边和其它的模块节点完成耦合。也就是模块的耦合应该被显式提出为超边节点。解决方案节点通过实现超边节点约束来完成定义。
	
	系统的改进架构:
	本系统采用敏捷开发的思路驱动。每一次的迭代是一次敏捷开发的Scrum。开发内容优先使用scrum.Backlog 来驱动。


## 当前的系统状态:

这是系统的product goal:
{{.ProductGoal}}

这是当前的Scrum.Backlog:
{{range $i, $backlog := .Backlogs}}
{{$backlog}} 
{{end}}


## 这是当前解决方案摘要(非超标节点):
{{.SolutionSummary}}

这是系统的全部可用的LLM Agents,这些Agent 用于修改，删除，新增超边节点或者是普通节点:

- AgentName: SceneryDesign
场景设计
	• 设计并使得一个场景，能够用直观的方式给出对目标知识的应用
	• 设计场景中的要是和互动方式。
	• 产出内容: 场景定义。要素定义。互动方式定义。

- AgentName: NodesAblation
删除超边节点
本次目标是 - 消融/删除现有的不重要的节点
• 产出对象: 节点增量，包括修改，删除，新增

- AgentName: MindStorming
头脑风暴
• 进行头脑风暴，提出新的想法和解决方案
• 产出内容: 新的想法列表

- AgentName:  ProductExperienceImprove
产品体验经理
• 依照需要尝试改进产品体验，使之更好
• 产出内容: 新的解决方案节点或者是节点的修改

- AgentName:  AgentRefineNode
精炼一个模块节点 
• 输入:需要解决的主题; 需要考虑的约束条件, 已有的模块节点等
• 产出内容: 新的解决方案节点或者是节点的修改

- AgentName:  PathnameRefactor
重新整理模块节点的路径名
• 产出内容: 新的解决方案节点名称

- AgentName:  StepByStepRefactor
分步求解器
设定目标 - 提出粒度合适的目标，确保目标有可能被高质量完成。
详细讨论并提出多个步骤的解决框架，以使得目标最有可能被完成。并交给后续的Agent 来进行详细实现
• 产出内容: 合适的目标，和对应的步骤框架

AgentName: EndGoalDrivenPlanner
中文名: 逆向目标规划器

描述:
以产品最终目标 (ProductGoal) 和当前系统状态为出发点，进行逆向分析和规划。
识别并定义出为达成最终目标，在当前阶段最亟需完成或最为关键的中间目标。
将此关键中间目标进一步细化为一个或多个具体的、可独立执行的子任务/小目标。
为每个子任务/小目标选择最合适的现有Agent去执行，并准备好调用指令。
产出内容: 包含具体任务描述、任务优先级以及为每个任务推荐的执行Agent的列表。可能会准备调用 TaskExecutionDelegator 来调度这些任务的执行。

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

	err := AgentEvoLearningCallback.Call(context.Background(), map[string]interface{}{
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

var KeyBacklogLearning = redisdb.NewListKey[*scrum.Backlog](redisdb.Opt.HttpVisit()).ConcatKey("Learning")
var topic = "直观感受1-20数字的大小"
var SolutionLearning = redisdb.NewHashKey[string, *SolutionGraphNode](redisdb.Opt.HttpVisit()).ConcatKey("Learning")

func EvoLearning() {
	var KeyBacklog = KeyBacklogLearning.ConcatKey(topic)
	var keySolution = SolutionLearning.ConcatKey(topic)

	const MaxThreads = 1
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)
	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		backlogs, _ := KeyBacklog.LRange(0, -1)
		allNodes, _ := keySolution.HGetAll()
		restartNeeded := false
		for _, item := range allNodes {
			if len(item.Id) > 4 {
				item.Id = utils.ID(item.BulletDescription, 4)
				keySolution.HSet(item.Id, item)
				restartNeeded = true
			}
			keySolution.HSet(item.Id, item)
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
			err := AgentEvoLearning.WithModels(models.Qwen3B32Thinking).Call(context.Background(), map[string]any{
				"Backlogs":        backlogs,
				"SuperEdges":      SuperEdges,
				"ProductGoal":     scrum.ProductGoalUniLearning,
				"HashKey":         keySolution,
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

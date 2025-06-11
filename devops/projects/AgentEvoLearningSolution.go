package projects

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	// "github.com/yourbasic/graph"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/scrum"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

type AgentSelector struct {
	WhatTodoInFollowingIter string                                 `description:"What is key focus to do in following iter, string"`
	MemoToTheNextIter       string                                 `description:"memo should pass to next iteration, string"`
	BackgroundNodesToPass   []string                               `description:"background solution nodes (Pathname) should pass to next iteration, array of string"`
	AllItems                map[string]*SolutionNode               `description:"-"`
	Backlogs                []*scrum.Backlog                       `description:"-"`
	ProductGoal             string                                 `description:"-"`
	HashKey                 redisdb.HashKey[string, *SolutionNode] `description:"-"`
	ThisAgent               *agent.Agent                           `description:"-"`
}

var AgentEvoLearningSolution = agent.NewAgent(template.Must(template.New("AgentEvoLearningCallback").Parse(`
## 本系统采用迭代方式来渐进实现系统的自动化构建。每一轮的迭代中，请通过一系列的Funtioncall 调用，来完善或改进方案的实现。
本系统采用敏捷开发的思路驱动。每一次的迭代是一次敏捷开发的Scrum。优先使用scrum.Backlog 来驱动开发。


## 当前的系统状态:

这是系统的product goal:
{{.ProductGoal}}

这是本次迭代的目标:
{{.WhatTodoInFollowingIter}}

这是来自上一次迭代的反馈:
{{.MemoToTheNextIter}}

{{if lt (len .RelativeSolutionNodes) 0 }}
这是当前重点关注的部分解决方案节点:
{{.RelativeSolutionNodes}}
{{end}}

这是当前的Scrum.Backlog:
{{range $i, $backlog := .Backlogs}}
{{$backlog}} 
{{end}}



- AgentName:  StepByStepRefactor
分步求解器
设定目标 - 提出粒度合适的目标，确保目标有可能被高质量完成。
详细讨论并提出多个步骤的解决框架，以使得目标最有可能被完成。并交给后续的Agent 来进行详细实现
• 产出内容: 合适的目标，和对应的步骤框架


## To Do (可选列表):

	1. AgentName: EndGoalDrivenPlanner
	中文名: 逆向目标规划器

	描述:
	以产品最终目标 (ProductGoal) 和当前系统状态为出发点，进行逆向分析和规划。
	识别并定义出为达成最终目标，在当前阶段最亟需完成或最为关键的中间目标。
	将此关键中间目标进一步细化为一个或多个具体的、可独立执行的子任务/小目标。
	为每个子任务/小目标选择最合适的现有Agent去执行，并准备好调用指令。
	产出内容: 包含具体任务描述、任务优先级以及为每个任务推荐的执行Agent的列表。可能会准备调用 TaskExecutionDelegator 来调度这些任务的执行。

	仅进行深度分析，以便从上面的UserAgents 中选择一个或者是多个Agent 来完成Backlog中的任务。
	不是所有的Agent 都适合当前的任务。不同的任务需要的实现精度差异很大。仅确保以Just-Enough 的方式来完成当前的任务。
	最后通过 FunctionCall:ApplySelectedAgent 来启用选定的Agent来进一步迭代系统。


	2. AgentName: SolutionRefine
	中文名: 方案改进器
	请按给定的本次迭代的目标，行深度思考，并提出有效的Refine 方案。
	最后通过一次或多次调用 FunctionCall:SolutionSuperEdgeRefine 来保存方案改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Id,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)删除无效节点


`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionSuperEdgeRefine", "create/modify/remove supernodes", func(newItem *SolutionNode) {
	newItem.BulletDescription = strings.TrimSpace(newItem.BulletDescription)
	var oItem *SolutionNode = nil
	if newItem.Pathname == "" {
		return
	}
	oItem, _ = newItem.HashKey.HGet(newItem.Pathname)
	if oItem != nil && oItem.Locked {
		return
	} else if newItem.Delete {
		pathname := filepath.Join(RootPath, newItem.Pathname)
		os.Remove(pathname)
		return
	} else if oItem != nil {
		newItem.BulletDescription, _ = lo.Coalesce(newItem.BulletDescription, oItem.BulletDescription)
		newItem.FileContent, _ = lo.Coalesce(newItem.FileContent, oItem.FileContent)
	}

	os.WriteFile(filepath.Join(RootPath, newItem.Pathname), []byte(newItem.FileContent), 0644)
	newItem.HashKey.HSet(newItem.Pathname, newItem)

})).WithToolCallMutextRun().WithTools(tool.NewTool("ApplySelectedAgent", "choose a proper agent for the next step", func(param *AgentSelector) {
	solution := ""
	if len(param.BackgroundNodesToPass) > 0 {
		solution = SolutionNodeList(lo.Values(param.AllItems)).PathnameSorted().View(param.BackgroundNodesToPass)
	}

	err := param.ThisAgent.Call(context.Background(), map[string]interface{}{
		"WhatTodoInFollowingIter": param.WhatTodoInFollowingIter,
		"MemoToTheNextIter":       param.MemoToTheNextIter,
		"RelativeSolutionNodes":   solution,
		"ProductGoal":             param.ProductGoal,
		"AllItems":                param.AllItems,
		"Backlogs":                param.Backlogs,
		"HashKey":                 param.HashKey,
	})
	if err != nil {
		fmt.Printf("Error calling AgentGenSuperEdge: %v\n", err)
	}
})).WithModels(models.Qwen3B32Thinking)

var KeyBacklogBase = redisdb.NewListKey[*scrum.Backlog](redisdb.Opt.HttpVisit())
var ThisTopic = "直观感受1-20数字的大小"
var SolutionBase = redisdb.NewHashKey[string, *SolutionNode](redisdb.Opt.HttpVisit())
var RootPath = "/Users/yang/doptime/evolab/web/app/perceptual-understanding-numbers-1-to-20"

func EvoLearningSolution() {
	var KeyBacklog = KeyBacklogBase.ConcatKey(ThisTopic)
	var keySolution = SolutionBase.ConcatKey(ThisTopic)

	const MaxThreads = 1
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)
	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		backlogs, _ := KeyBacklog.LRange(0, -1)
		allNodes, _ := keySolution.HGetAll()
		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore

		SolutionSummary := SolutionNodeList(lo.Values(allNodes)).PathnameSorted().SummaryView()

		go func(backlogs []*scrum.Backlog, SolutionSummary string, AllItems map[string]*SolutionNode) {
			defer func() { <-MaxThreadsSemaphore }()
			err := AgentEvoLearningSolution.WithModels(models.Qwen3B32Thinking).Call(context.Background(), map[string]any{
				"Backlogs":        backlogs,
				"ProductGoal":     string(scrum.ProductGoalUniLearning) + "\n\n" + PlanForTopic + "\n\n" + FileItems,
				"HashKey":         keySolution,
				"SolutionSummary": SolutionSummary,
			})
			if err != nil {
				fmt.Printf("Agent call failed: %v\n", err)
			}
		}(backlogs, SolutionSummary, allNodes)
	}
	// Wait for all the goroutines to finish)
	for i := 0; i < MaxThreads; i++ {
		MaxThreadsSemaphore <- struct{}{}
	}

}

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


这是当前的解决方案:
{{.Solution}}

这是当前的Scrum.Backlog:
{{range $i, $backlog := .Backlogs}}
{{$backlog}} 
{{end}}



## To Do : SolutionRefine
	请按给定的本次迭代的目标，行深度思考，并提出有效的Refine 方案。
	最后通过一次或多次调用 FunctionCall:SolutionSuperEdgeRefine 来保存方案改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Id,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)删除无效节点


`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionSuperEdgeRefine", "create/modify/remove supernodes", func(newItem *SolutionFileNode) {
	newItem.BulletDescription = strings.TrimSpace(newItem.BulletDescription)
	var oItem *SolutionFileNode = nil
	if newItem.Filename == "" {
		return
	}
	oItem, _ = newItem.HashKey.HGet(newItem.Filename)
	if oItem != nil && oItem.Locked {
		return
	} else if newItem.Delete {
		pathname := filepath.Join(RootPath, newItem.Filename)
		os.Remove(pathname)
		return
	} else if oItem != nil {
		newItem.BulletDescription, _ = lo.Coalesce(newItem.BulletDescription, oItem.BulletDescription)
		newItem.FileContent, _ = lo.Coalesce(newItem.FileContent, oItem.FileContent)
	}

	os.WriteFile(filepath.Join(RootPath, newItem.Filename), []byte(newItem.FileContent), 0644)
	newItem.HashKey.HSet(newItem.Filename, newItem)

}))

var KeyBacklogBase = redisdb.NewListKey[*scrum.Backlog](redisdb.Opt.HttpVisit())
var ThisTopic = "直观感受1-20数字的大小"
var SolutionBase = redisdb.NewHashKey[string, *SolutionFileNode](redisdb.Opt.HttpVisit())
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

		SolutionSummary := SolutionFileNodeList(lo.Values(allNodes)).PathnameSorted().View()

		go func(backlogs []*scrum.Backlog, SolutionSummary string, AllItems map[string]*SolutionFileNode) {
			defer func() { <-MaxThreadsSemaphore }()
			err := AgentEvoLearningSolution.WithModels(models.Qwen3B32Thinking).Call(context.Background(), map[string]any{
				"WhatTodoInFollowingIter": "End Goal Driven Planner",
				"Backlogs":                backlogs,
				"ProductGoal":             string(scrum.ProductGoalUniLearning) + "\n\n这是当前规划的游戏场景:\n" + PlanForTopic + "\n\n" + FileItems,
				"HashKey":                 keySolution,
				"AllItems":                allNodes,
				"Solution":                SolutionSummary,
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

package projects

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	// "github.com/yourbasic/graph"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/scrum"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

var fileUpdating = map[string]time.Time{}
var CanUpdateFile = func(filename string) bool {
	if t, ok := fileUpdating[filename]; ok {
		if time.Since(t) < 240*time.Second {
			time.Sleep(100 * time.Millisecond)
			return false
		}
	}
	fileUpdating[filename] = time.Now()
	return true
}

func nodeRefine(newItem *SolutionFileNode) {
	newItem.BulletDescription = strings.TrimSpace(newItem.BulletDescription)
	var oItem *SolutionFileNode = nil
	if newItem.Filename == "" {
		return
	}
	oItem, _ = newItem.HashKey.HGet(newItem.Filename)
	if newItem.Delete {
		pathname := filepath.Join(RootPath, newItem.Filename)
		newItem.HashKey.HDel(newItem.Filename)
		os.Remove(pathname)
		return
	} else if oItem != nil {
		newItem.BulletDescription, _ = lo.Coalesce(newItem.BulletDescription, oItem.BulletDescription)
		newItem.FileContent, _ = lo.Coalesce(newItem.FileContent, oItem.FileContent)
	}
	filename := filepath.Join(RootPath, newItem.Filename)
	err := os.WriteFile(filename, []byte(newItem.FileContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filename, err)
		return
	}
	newItem.HashKey.HSet(newItem.Filename, newItem)

}

var AgentEvoLearningSolution = agent.NewAgent(template.Must(template.New("AgentEvoLearningCallback").Parse(`
## 本系统采用迭代方式完成开发工作,以实现或改进方案的实现。
请确保采用简单且Just-Enough的方式来实现或改进方案的实现。

## 当前的系统状态:
这是系统的product goal:
{{.ProductGoal}}

这是本次迭代的目标: 
迭代一个或若干个文件，以使得项目场景得以简洁、尽可能完整地建构
如果必要
	- 对粒度过大，难以实施的文件，优先对粒度进行拆分，也就是创建多个更细粒度的文件实现，并且删除原有的文件。
	- 如果有必要，提出或删除相关模块文件。维护解决方案的完整性和简洁性。
	- 优先删除无效的引用、消除错误。尽可能使得项目处于可编译、可运行的状态。


这是当前的解决方案:
{{.Solution}}

这是当前的编译信息:
{{.runtimeError}}


## To Do : SolutionFileRefine
	请按给定的本次迭代的目标，行深度思考，并提出有效的Refine 方案，确保目标文件内容正确，并且确保在下一轮的迭代中可以被进一步改善。
	最后通过（多次）调用 FunctionCall:SolutionFileRefine 来提交不同文件的改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Filename,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)删除无效节点


`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionFileRefine", "create/modify/remove solution file", nodeRefine))

var KeyBacklogBase = redisdb.NewListKey[*scrum.Backlog](redisdb.Opt.HttpVisit())
var ThisTopic = "直观感受1-20数字的大小"
var SolutionBase = redisdb.NewHashKey[string, *SolutionFileNode](redisdb.Opt.HttpVisit())
var RootPath = "/Users/yang/doptime/evolab/web/app/perceptual-understanding-numbers-1-to-20"

func EvoLearningSolution() {
	//var KeyBacklog = KeyBacklogBase.ConcatKey(ThisTopic)
	var keySolution = SolutionBase.ConcatKey(ThisTopic)

	const MaxThreads = 1
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	// allNodes, _ := keySolution.HGetAll()
	// for _, fileNode := range allNodes {
	// 	if strings.HasPrefix(fileNode.Filename, "./") {
	// 		keySolution.HDel(fileNode.Filename)
	// 		fileNode.Filename = strings.TrimPrefix(fileNode.Filename, "./")
	// 		keySolution.HSet(fileNode.Filename, fileNode)

	// 		filename := filepath.Join(RootPath, fileNode.Filename)
	// 		os.WriteFile(filename, []byte(fileNode.FileContent), 0644)
	// 	}
	// }

	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		//backlogs, _ := KeyBacklog.LRange(0, -1)
		allNodes, _ := keySolution.HGetAll()

		//load nodes from file
		for _, node := range allNodes {
			utils.TextFromFile(filepath.Join(RootPath, node.Filename), &node.FileContent)
		}

		SolutionSummary := SolutionFileNodeList(lo.Values(allNodes))
		time.Sleep(300 * time.Millisecond)

		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		go func(SolutionSummary string, AllItems map[string]*SolutionFileNode) {
			defer func() { <-MaxThreadsSemaphore }()

			// 当前项目已经通过编译。摄像头也可以打开并且显示视频。但前端页面只有空白的页面和
			// 等待挑战开始 文字
			// 没有办法开始游戏。接下来应该怎么让游戏正常运行起来呢？请给出方案。
			runtimeError, _ := utils.ExtractNextJSError("http://localhost:3000/perceptual-understanding-numbers-1-to-20/error.js")
			runtimeError = `
命题端 的 视觉化的方式显示数量 异常，应该能够显示相应数量的AnerygyBall. 

ChallengeWalls 工作异常
如果注释ChallengeWalls ，则不会报错。

如果不注释ChallengeWalls，则会报错，提示
Something went wrong:
Cannot convert undefined or null to object
Try again
`
			err := AgentEvoLearningSolution.WithModels(models.Gemini25FlashNonthinking).Call(context.Background(), map[string]any{
				"runtimeError": string(runtimeError),
				"ProductGoal":  string(scrum.ProductGoalUniLearning) + "\n\n这是当前规划的游戏场景:\n" + PlanForTopic + "\n\n",
				"HashKey":      keySolution,
				"AllItems":     allNodes,
				"Solution":     SolutionSummary,
			})
			if err != nil {
				fmt.Printf("Agent call failed: %v\n", err)
			}
		}(SolutionSummary.FullView(), allNodes)
	}
	// Wait for all the goroutines to finish)
	for i := 0; i < MaxThreads; i++ {
		MaxThreadsSemaphore <- struct{}{}
	}

}

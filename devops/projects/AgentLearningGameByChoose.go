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
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

// var ThisTopic = "直观感受1-20数字的大小"
var SolutionBaseLearnByChoose = redisdb.NewHashKey[string, *FileRefine](redisdb.Opt.HttpVisit(), redisdb.Opt.Key("SolutionBaseLearnByChoose"))

// var RootPathLearnByChoose = "/Users/yang/doptime/evolab/web/app/perceptual-understanding-numbers-1-to-20"
var RootPathLearnByChoose = "/Users/yang/doptime/evolab/web/app/learnning-game-by-choose"
var ExtraPathLearnByChoose = "../../components/guesture"

func LoadExtraPathToMap1(solution map[string]*FileRefine) {
	extraPath := filepath.Join(RootPathLearnByChoose, ExtraPathLearnByChoose)
	ExtraPathFiles, _ := os.ReadDir(extraPath)
	for _, file := range ExtraPathFiles {
		fn := filepath.Join(extraPath, file.Name())
		filename := ExtraPathLearnByChoose + "/" + file.Name()
		solution[filename] = &FileRefine{
			Filename:    filename,
			FileContent: utils.TextFromFile(fn),
		}
	}
}
func SaveContentToPath1(node *FileRefine) {
	//save to root path
	filename := filepath.Join(RootPathLearnByChoose, node.Filename)
	err := os.WriteFile(filename, []byte(node.FileContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filename, err)
	}
}

var AgentEvoLearningSolutionLearnByChoose = agent.NewAgent(template.Must(template.New("AgentEvoLearningSolutionLearnByChoose").Parse(`
## 本系统采用迭代方式完成开发工作,以实现或改进方案的实现。
- 简洁可靠: 优先删除无效的引用、消除错误。确保实现方案和代码简洁可靠。
- 以终为始: 努力实现系统的目标意图，并使得项目处于可编译、可运行的状态。
- 粒度适中:，对粒度过大，难以实施的文件，优先对粒度进行拆分，也就是创建多个更细粒度的文件实现，并且删除原有的文件; 反之需要用高内聚低耦合的方式，重构文件内容。

## To Do : SolutionFileRefine
	请按给定的本次迭代的目标，行深度思考，并提出有效的Refine 方案，确保目标文件内容正确，并且确保在下一轮的迭代中可以被进一步改善。
	最后通过（多次）调用 FunctionCall:SolutionFileRefine 来提交不同文件的改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Filename,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)删除无效节点


## 当前的系统状态:
这是系统的product goal:
{{.ProductGoal}}


这是当前的解决方案:
{{.Solution}}


这是当前的编译信息:
{{.runtimeError}}


`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionFileRefine", "create/modify/remove solution file", func(newItem *FileRefine) {
	newItem.BulletDescription = strings.TrimSpace(newItem.BulletDescription)
	var oItem *FileRefine = nil
	if newItem.Filename == "" {
		return
	}
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "Pathname")
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "Path")
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
	newItem.SaveContentToPath(RootPathLearnByChoose)
	newItem.HashKey.HSet(newItem.Filename, newItem)

}))

func EvoLearnByChooseSolution() {
	//var KeyBacklog = KeyBacklogBase.ConcatKey(ThisTopic)
	var keySolution = SolutionBaseLearnByChoose.ConcatKey(ThisTopic)

	const MaxThreads = 1
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		//backlogs, _ := KeyBacklog.LRange(0, -1)
		allNodes, _ := keySolution.HGetAll()

		//load nodes from file
		for _, node := range allNodes {
			utils.TextFromFile(filepath.Join(RootPathLearnByChoose, node.Filename), &node.FileContent)
		}
		LoadExtraPathToMapFileRefineMap(RootPathLearnByChoose, ".", allNodes)
		LoadExtraPathToMapFileRefineMap(RootPathLearnByChoose, ExtraPathLearnByChoose, allNodes)

		SolutionSummary := FileRefineList(lo.Values(allNodes))
		time.Sleep(300 * time.Millisecond)

		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		go func(SolutionSummary string, AllItems map[string]*FileRefine) {
			defer func() { <-MaxThreadsSemaphore }()

			//- 蓝色的手势位置x轴方向（左右移动方向）和手的移动方向相反。

			//runtimeError, _ := utils.ExtractNextJSError("http://localhost:3000/perceptual-understanding-numbers-1-to-20/error.js")
			runtimeError := `
现现有的代码是另一个相近项目的源码。和现在的项目差异非常巨大。要求参考旧项目进行完全的重构。
数据也需要重新设计。	

这是现在得到的改进意见，请将它提交到SolutionFileRefine中。

` + utils.TextFromClipboard()
			ProductGoalUniLearning := utils.TextFromFile("/Users/yang/eloevo/devops/projects/learninggame.md")
			//Gemini25Flashlight Gemini25ProAigpt
			err := AgentEvoLearningSolutionLearnByChoose.WithModels(models.Qwen3B14). //CopyPromptOnly(). //UseClipboardMsg().
													Call(context.Background(), map[string]any{
					"runtimeError": string(runtimeError),
					"ProductGoal":  string(ProductGoalUniLearning) + "\n\n",
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

package learnbychoose

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
var SolutionBaseLearnByChoose = redisdb.NewHashKey[string, *FileRefine](redisdb.Opt.HttpVisit(), redisdb.Opt.Key("ProjectRefine:LearnByChoose"))

// var RootPathLearnByChoose = "/Users/yang/doptime/evolab/web/app/perceptual-understanding-numbers-1-to-20"
// /Users/yang/doptime/evolab/web/app/learn-by-choose/data-mock.ts
var RootPathLearnByChoose = "/Users/yang/doptime/evolab/web/app/learn-by-choose"
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

var AgentEvoLearningSolutionLearnByChoose = agent.NewAgent().WithTemplate(template.Must(template.New("AgentEvoLearningSolutionLearnByChoose").Parse(`
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



{{.runtimeError}}

{{.Revision}}

`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionFileRefine", "create/modify/remove solution file", func(newItem *FileRefine) {
	newItem.BulletDescription = strings.TrimSpace(newItem.BulletDescription)
	var oItem *FileRefine = nil
	if newItem.Filename == "" {
		return
	}
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "Pathname")
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "Path")
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "./")
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "src/")
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "app/")
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
	var keySolution = SolutionBaseLearnByChoose

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
		//LoadExtraPathToMapFileRefineMap(RootPathLearnByChoose, ExtraPathLearnByChoose, allNodes)

		SolutionSummary := FileRefineList(lo.Values(allNodes))
		time.Sleep(300 * time.Millisecond)

		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		go func(SolutionSummary string, AllItems map[string]*FileRefine) {
			defer func() { <-MaxThreadsSemaphore }()

			//- 蓝色的手势位置x轴方向（左右移动方向）和手的移动方向相反。

			//runtimeError, _ := utils.ExtractNextJSError("http://localhost:3000/perceptual-understanding-numbers-1-to-20/error.js")
			runtimeError := "这是当前的编译信息或迭代需求:\n" + `
fix: 
修复选项卡条目的内容高度。现在的选项卡条目高度固定的。但可以容纳的文本相当有限。请允许选项卡条目根据内容自适应高度。
现在会自动调整宽度，但是不会自动调整高度。现有的内容在高度上被遮挡。无法完全显示。需要修复


请思考，并且有重点地改进现有系统. 

`
			Revision := utils.TextFromClipboard()
			if len(Revision) > 200 && strings.LastIndex(Revision, ".ts") > 60 {
				Revision = `对上面的需求，这是我的修改意见，请接受合理的内容变动，如果不合理则重新思考实现。并提交完整的代码到本地文件:
1. 不同的文件内容，你需要多次调用 SolutionFileRefine , 每次修改一个本地文件当中。要遵守正确的函数调用规定。不要受被用户数据中错误的格式影响。
2. 文件内容必须完整而没有遗漏和错误，不能只提交部分内容。如果是增量修改，请将其转化为全量的文件内容。以避免编译失败。
3. 对每个文件的修改，应该先进行思考或者是讨论，以明确修改意图和修改内容。
4. 可能存在部分文件已经完成修改的情形。对于已经完成的修改，可以忽略或者是跳过。
给定的修改建议内容如下：` + Revision
			} else {
				Revision = ""
			}
			ProductGoalUniLearning := utils.TextFromFile("/Users/yang/learn-by-choose-goserver/learninggame.md")
			//Gemini25Flashlight Gemini25ProAigpt
			err := AgentEvoLearningSolutionLearnByChoose.WithModels(models.Glm45AirLocal). //CopyPromptOnly(). //Qwen3B32Thinking
													Call(context.Background(), map[string]any{
					"runtimeError": string(runtimeError),
					"ProductGoal":  string(ProductGoalUniLearning) + "\n\n",
					"HashKey":      keySolution,
					"AllItems":     allNodes,
					"Solution":     SolutionSummary,
					"Revision":     Revision,
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

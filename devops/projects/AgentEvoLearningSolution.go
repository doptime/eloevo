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
	SaveContentToPath(newItem)
	newItem.HashKey.HSet(newItem.Filename, newItem)

}

var AgentEvoLearningSolution = agent.NewAgent(template.Must(template.New("AgentEvoLearningCallback").Parse(`
小孩今年高考，文科，342，福建省排名是49735/60252人（文科生是60252人报考）

从职业和未来前景规划看，有什么建议。请帮设计一条国内学习，去日本或国外工作的路径


`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionFileRefine", "create/modify/remove solution file", nodeRefine))

var KeyBacklogBase = redisdb.NewListKey[*scrum.Backlog](redisdb.Opt.HttpVisit())

// var ThisTopic = "直观感受1-20数字的大小"
var ThisTopic = "中文认词"
var SolutionBase = redisdb.NewHashKey[string, *SolutionFileNode](redisdb.Opt.HttpVisit())

// var RootPath = "/Users/yang/doptime/evolab/web/app/perceptual-understanding-numbers-1-to-20"
var RootPath = "/Users/yang/doptime/evolab/web/app/chinese-word-recognition"
var ExtraPath = "../../components/guesture"

func LoadExtraPathToMap(solution map[string]*SolutionFileNode) {
	extraPath := filepath.Join(RootPath, ExtraPath)
	ExtraPathFiles, _ := os.ReadDir(extraPath)
	for _, file := range ExtraPathFiles {
		fn := filepath.Join(extraPath, file.Name())
		filename := ExtraPath + "/" + file.Name()
		solution[filename] = &SolutionFileNode{
			Filename:    filename,
			FileContent: utils.TextFromFile(fn),
		}
	}
}
func SaveContentToPath(node *SolutionFileNode) {
	//save to root path
	filename := filepath.Join(RootPath, node.Filename)
	err := os.WriteFile(filename, []byte(node.FileContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filename, err)
	}
}

func EvoLearningSolution() {
	//var KeyBacklog = KeyBacklogBase.ConcatKey(ThisTopic)
	var keySolution = SolutionBase.ConcatKey(ThisTopic)

	const MaxThreads = 1
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		//backlogs, _ := KeyBacklog.LRange(0, -1)
		allNodes, _ := keySolution.HGetAll()

		//load nodes from file
		for _, node := range allNodes {
			utils.TextFromFile(filepath.Join(RootPath, node.Filename), &node.FileContent)
		}
		LoadExtraPathToMap(allNodes)

		SolutionSummary := SolutionFileNodeList(lo.Values(allNodes))
		time.Sleep(300 * time.Millisecond)

		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		go func(SolutionSummary string, AllItems map[string]*SolutionFileNode) {
			defer func() { <-MaxThreadsSemaphore }()

			//- 蓝色的手势位置x轴方向（左右移动方向）和手的移动方向相反。

			//runtimeError, _ := utils.ExtractNextJSError("http://localhost:3000/perceptual-understanding-numbers-1-to-20/error.js")
			runtimeError := `
			
完成数据结构的调整：
现在的数据结构是这样：
    {
        id: 'three',
        word: '3',
        isNumeric: true,
        hints: {
            en: 'three',
            jp: '三',
            es: 'tres',
            emoji: '3️⃣',
            root: '词根: tri-, 来自拉丁语 tres，意为“三”。',
            association: '联想: 第三，三角形，三叶草，三原色。',
            svg: null,
        }
    }

新的数据结构是：

export interface WordLearningData {
    Word: string;
    AssociativeImaginationWords: string;
}

前端的页面也需要调整，使用新的数据结构。AssociativeImaginationWords包含换行符\n。需要能够正常显式。

// call using apiWordLearningData(WordSensationTask{Words: ["苹果", "窗户", "小说"]})
export const apiWordLearningData = createApi<WordSensationTask, WordLearningData[]>("WordLearningData");


`
			//Gemini25Flashlight
			err := AgentEvoLearningSolution.WithModels(models.Gemini20FlashImageAigpt). //UseClipboardMsg().
													Call(context.Background(), map[string]any{
					"runtimeError": string(runtimeError),
					"ProductGoal":  string(scrum.ProductGoalUniLearning) + "\n\n这是当前规划的游戏场景:\n" + PlanForWordSensation + "\n\n",
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

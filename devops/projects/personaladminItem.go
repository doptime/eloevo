package projects

import (
	"context"
	"fmt"
	"slices"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/prototype"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

type PersonalAdmin struct {
	Id       string         `description:"Id, string, unique"`
	Question string         `description:"Question,string"`
	Score    int            `msgpack:"Elo"`
	Extra    map[string]any `msgpack:"-" description:"Extra, call parameter of Agent"`
}

func (u *PersonalAdmin) GetId() string {
	return u.Id
}
func (u *PersonalAdmin) ScoreAccessor(delta ...int) int {
	var eloDelta int = append(delta, 0)[0]
	if eloDelta != 0 {
		u.Score += eloDelta
		keyPersonalAdmin.HSet(u.Id, u)
	}
	return u.Score
}

var keyPersonalAdmin = redisdb.NewHashKey[string, *PersonalAdmin](redisdb.Opt.Rds("projects"), redisdb.Opt.Key("PersonalAdminByItem"))

// 为什么Qwen能自我改进推理，Llama却不行 https://mp.weixin.qq.com/s/OvS61OrDp6rB-R5ELg48Aw
// 并且每次就一个确切的改进方向进行深度分析，反代的深度分析第一性原理之上的需求，深度创新以做出实质的改进。要痛恨泛泛而谈的内容，重复空洞的内容，因为现在是在开发世界级的工具。
var AgentPersonalAdmin = agent.NewAgent(template.Must(template.New("utilifyFunction").Parse(`
我想要用填写开放问题的方式来最大化一天的时间效用。
有哪些问题应当保持在清单上面？
避免超过15个问题。否则我将浪费太多时间在填表上。

这是现有的方案 ：
{{range  $index, $item := .ItemList}}
Id:{{$item.Id}} Question:{{$item.Question}}
{{end}}

ToDo:
步骤1. 对现有的选项进行思考和评估：
	1、对回溯或在检测到错误进行显式修改（例如，「这种方法行不通，因为...」）；
	2、验证或系统地检查中间结果（例如，「让我们通过... 来验证这个结果」）；
	3、子目标设定，即将复杂问题分解为可管理的步骤（例如，「要解决这个问题，我们首先需要...」）；
	4、逆向思考，即在目标导向的推理问题中，从期望的结果出发，逐步向后推导，找到解决问题的路径。（例如，「要达到 75 的目标，我们需要一个能被... 整除的数字」）。
	
{{if gt .TotoalNodes 13}}
步骤2. 根据上面的讨论. 重新整理方案。
首先，选择出需要保留到方案中的项目，并且把这些方案中的项目按重要性排列，整理成为ItemsInSolutionSortedById。
其次，选择出方案中需要移除的项目。并且把这些方案中的项目按重要性排列，整理成为ItemsExpelFromSolutionSorted。
将现有方案选项从最好到最坏排序(用Id列表表示)。调用FunctionCall:SaveItemsSorted 保存ItemsInSolutionSortedById 和 ItemsExpelFromSolutionSorted结果。
{{else}}
步骤2. 在讨论的基础上，进一步给方案补充一个新的方案中的条目。对条目同样按步骤1改进后，调用FunctionCall:NewPersonalAdmin 保存新的方案。
{{end}}

`))).WithToolCallLocked().WithTools(tool.NewTool("SaveItemsSorted", "Save sorted Items, Items represented as Id list.", func(model *prototype.SolutionRefine) {
	if model == nil {
		return
	}
	// all, _ := keyPersonalAdmin.HGetAll()
	// sortedElos := elo.ToSlice(all).Sort().Reverse()
	// ids := append(model.ItemsShouldKeptInSolutionSorted, model.ItemsShouldRemoveFromSolutionSorted...)
	// elo.BatchUpdateRanking(sortedElos.TakeByIds(ids...)...)

})).WithTools(tool.NewTool("NewPersonalAdmin", "create a new Solution item", func(model *PersonalAdmin) {
	model.Id = redisdb.NanoId(8)
	model.Score = 1000
	keyPersonalAdmin.HSet(model.Id, model)
}))

func PersonalAdminExploration() {
	const MaxThreads = 4
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	// keyPersonalAdmin.HSet("default", &PersonalAdmin{Id: "default0", Question: "今天我最期待的事情是什么？（关注积极面，激发动力）", Score: 1000})
	// keyPersonalAdmin.HSet("default1", &PersonalAdmin{Id: "default1", Question: "如果今天只能完成一件事，那会是什么？（明确优先级，避免分散精力）", Score: 1000})
	// keyPersonalAdmin.HSet("default2", &PersonalAdmin{Id: "default2", Question: "为了实现今天的目标，我需要什么样的能量/心态？（关注内在状态，为行动做准备）", Score: 1000})
	// keyPersonalAdmin.HSet("default3", &PersonalAdmin{Id: "default3", Question: "今天有什么潜在的挑战，我如何提前准备？（预见问题，减少应对压力）", Score: 1000})
	// keyPersonalAdmin.HSet("default4", &PersonalAdmin{Id: "default4", Question: "我今天可以为自己做些什么，让自己感觉更好？（关注自我关怀，提升幸福感）", Score: 1000})
	// keyPersonalAdmin.HSet("default5", &PersonalAdmin{Id: "default5", Question: "为了实现我最重要的目标，今天需要完成哪些具体的步骤？（将大目标分解为小任务）", Score: 1000})
	// keyPersonalAdmin.HSet("default6", &PersonalAdmin{Id: "default6", Question: "这些任务分别需要多少时间？（预估时间，提高计划的准确性）", Score: 1000})
	// keyPersonalAdmin.HSet("default7", &PersonalAdmin{Id: "default7", Question: "我什么时候精力最充沛，适合处理哪些任务？（根据精力曲线安排任务，提高效率）", Score: 1000})
	// keyPersonalAdmin.HSet("default8", &PersonalAdmin{Id: "default8", Question: "我今天有哪些固定的会议/约会？（明确时间限制，合理安排其他任务）", Score: 1000})
	// keyPersonalAdmin.HSet("default9", &PersonalAdmin{Id: "default9", Question: "我今天有哪些可以授权/外包的任务？（释放精力，专注于核心工作）", Score: 1000})
	// keyPersonalAdmin.HSet("default10", &PersonalAdmin{Id: "default10", Question: "我今天需要预留多少弹性时间，应对突发情况？（留有余地，避免计划被打乱）", Score: 1000})
	// keyPersonalAdmin.HSet("default11", &PersonalAdmin{Id: "default11", Question: "我今天完成了哪些任务？（记录进展，增强成就感）", Score: 1000})
	// keyPersonalAdmin.HSet("default12", &PersonalAdmin{Id: "default12", Question: "今天我学到了什么？（总结经验，促进成长）", Score: 1000})
	// keyPersonalAdmin.HSet("default13", &PersonalAdmin{Id: "default13", Question: "今天我最大的挑战是什么？我如何应对的？（分析问题，提升解决问题的能力）", Score: 1000})
	// keyPersonalAdmin.HSet("default14", &PersonalAdmin{Id: "default14", Question: "今天我最感激的事情是什么？（培养感恩之心，提升幸福感）", Score: 1000})

	for i, TotalTasks := 0, 1000*1000; i < TotalTasks; i++ {
		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		go func() {
			defer func() { <-MaxThreadsSemaphore }()
			best, _ := keyPersonalAdmin.HGetAll()
			listSorted := lo.Values(best)
			slices.SortFunc(listSorted, func(a, b *PersonalAdmin) int {
				return -(a.Score - b.Score)
			})
			//remove the worst 2
			for i := len(listSorted) - 1; i >= 0 && listSorted[i].Score < 996; i-- {
				keyPersonalAdmin.ConcatKey("Expired").HSet(listSorted[i].GetId(), listSorted[i])
				keyPersonalAdmin.HDel(listSorted[i].GetId())
				listSorted = listSorted[:i]
			}
			//print the lefts
			for i, v := range listSorted {
				fmt.Println("solution top ", i+1, v.GetId(), "Elo", v.ScoreAccessor(), v.Question)
			}
			//models.Qwq32B, models.Gemma3, models.DeepSeekV3
			err := AgentPersonalAdmin.WithModels(models.Qwq32B).Call(context.Background(),
				map[string]any{"ItemList": listSorted, "TotoalNodes": len(best)},
			)
			if err != nil {
				fmt.Printf("Agent call failed: %v\n", err)
			}
		}()
	}
	// Wait for all the goroutines to finish)
	for i := 0; i < MaxThreads; i++ {
		MaxThreadsSemaphore <- struct{}{}
	}

}

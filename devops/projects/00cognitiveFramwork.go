package projects

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"text/template"
	"time"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/elo"
	"github.com/doptime/eloevo/mixincached"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
)

// UtilityFunction = exp(WeightMarketSizeln(MarketSize) + 0.18ln(MarketGrowthRate) + 0.22ln(ExpectedReturn) + 0.10ln(TechnicalFeasibility) + 0.15ln(InnovationPotential) + 0.080ln(ResourceAllocation) - 0.12ln(ProjectRisk + 1) - 0.080ln(CompetitionIntensity) - 0.10ln(ImplementationDifficulty) + 0.060ln(TimeToMarket) + 0.040ln(TeamExperience) + 0.050ln(PolicySupport))
type RationalCognitiveFramework struct {
	Id    string
	Items map[string]any
}

func (u *RationalCognitiveFramework) GetId() string {
	return u.Id
}
func (u *RationalCognitiveFramework) ScoreAccessor(delta ...int) int {
	var eloDelta int = append(delta, 0)[0]
	return mixincached.WithElo("projects", "RationalCognitiveFramework", 1000).ScoreAccessor(u.Id, float64(eloDelta))
}

var RationalCognitionFrameworkDefault = &RationalCognitiveFramework{
	Id: "default",
	Items: map[string]any{
		"效用函数": "用于评估项目的效用函数，完成理论认知范式",
		"相关要素": "考虑问题时候，需要引入的用来定义问题的要素",
		"实际局限": "在执行的时候会遇到的局限，这是从前往后的，具有行动导向的认知范式",
		"理论局限": "在执行的时候会遇到的局限，这是从后往前的，第一性原理认知范式",
		"改进方案": "提出一个新的认知范式，用于替代现有的认知范式",
	},
}

type ItemsSorted struct {
	ItemsRefById []string `description:"Items sorted. Referenced by Items Id"`
}

var keyRationalCognitionFramework = redisdb.NewHashKey[string, *RationalCognitiveFramework](redisdb.Opt.Rds("projects"))

// 为什么Qwen能自我改进推理，Llama却不行 https://mp.weixin.qq.com/s/OvS61OrDp6rB-R5ELg48Aw
// 并且每次就一个确切的改进方向进行深度分析，反代的深度分析第一性原理之上的需求，深度创新以做出实质的改进。要痛恨泛泛而谈的内容，重复空洞的内容，因为现在是在开发世界级的工具。
var AgentUtilityFrameGen = agent.NewAgent(template.Must(template.New("utilifyFunction").Parse(`
现在我们要演进有限理性认知的一般范式。目标是生成一个数据模板，或者说是元认知模板。以便通过数百遍地迭代有限理性认知，从而最终完成对目标任务的认知。

这里我们有一些期望：
1. 如果一个认知范式需要进一步深入讨论一些课题。那么这个范式应该允许显性的引入这些课题，并使得这些课题在后续的迭代当中，被逐渐深入地讨论。
2. 有时候，有些主题包含一些特殊的要素，比如ToDoList,或Agenda . 那么在目标范式结构中，应当允许当主题更加明确之后，这些要素已经被当前认知范式的所允许，或者说动态补充，尽管这些要素在更一般范式当中并被需要。

这是现有的若干理性认知的一般范式 ：
{{range  $index, $item := .RationalCognitionFrameworks}}
第{{$index}}个认知范式：
	Id:{{$item.Id}}
	Items:
	{{range $key, $value := $item.BulletNames}}
		{{$key}}:{{$value}}
	{{end}}

{{end}}



ToDoList:
1. 把现有的多个方案视同一个整体。并且通过下面的4个办法，讨论以便找出改善后的新的方案：
	1、回溯或在检测到错误时显式修改方法（例如，「这种方法行不通，因为...」）；
	2、验证或系统地检查中间结果（例如，「让我们通过... 来验证这个结果」）；
	3、子目标设定，即将复杂问题分解为可管理的步骤（例如，「要解决这个问题，我们首先需要...」）；
	4、逆向思考，即在目标导向的推理问题中，从期望的结果出发，逐步向后推导，找到解决问题的路径。（例如，「要达到 75 的目标，我们需要一个能被... 整除的数字」）。 
在讨论的基础上，尝试提出一个更好的，新的方案。注意，属性Item 是一个Map,可以容纳无限的Key value 描述和无限层级的描述, 充分利用这个特点使得方案模块化。
验证（系统错误检查）、回溯（放弃失败的方法）、子目标设定（将问题分解为可管理的步骤）和逆向思考（从期望结果推理到初始输入）。

2. 根据上面的讨论，识别出现有的几个方案的优势和缺陷. 将他们从最好到最坏排序(用Id列表表示)。
3. 调用FunctionCall:SaveItemsSorted 保存排序结果。
4. 调用FunctionCall:RationalCognitiveFramework 保存新的方案。
`))).WithToolCallMutextRun().WithTools(tool.NewTool("SaveItemsSorted", "Save sorted Items, Items represented as Id list.", func(model *ItemsSorted) {
	if model == nil || len(model.ItemsRefById) == 0 {
		return
	}
	all, _ := keyRationalCognitionFramework.HGetAll()
	var values []elo.Elo
	for _, v := range all {
		if slices.Contains(model.ItemsRefById, v.Id) {
			values = append(values, v)
		}
	}
	elo.BatchUpdateRanking(values...)
	playersOrderByElo := lo.Keys(all)
	//sort the players by elo
	slices.SortFunc(playersOrderByElo, func(i, j string) int {
		return all[i].ScoreAccessor() - all[j].ScoreAccessor()
	})
	if len(all) > 16 {

		//remove the worst 2
		mapToRemove := lo.SliceToMap(playersOrderByElo[:2], func(v string) (string, *RationalCognitiveFramework) { return v, all[v] })
		for _, v := range lo.Values(mapToRemove) {
			if v.ScoreAccessor() > 1000 {
				delete(mapToRemove, v.Id)
			}
		}
		keyRationalCognitionFramework.ConcatKey("Expired").HMSet(mapToRemove)
		keyRationalCognitionFramework.HDel(playersOrderByElo[:2]...)

		//remove the best 1
		if rand.Float64() < 0 {
			bestNode := playersOrderByElo[len(playersOrderByElo)-1]
			NowString := time.Now().Format("2006-01-02 15:04:05")
			keyRationalCognitionFramework.ConcatKey("BestExpired").HSet(NowString, all[bestNode])
			keyRationalCognitionFramework.HDel(bestNode)
		}
	}
	slices.Reverse(playersOrderByElo)
	for i, v := range playersOrderByElo[0:min(10, len(playersOrderByElo))] {
		fmt.Println("Best Model,top ", i+1, v, "Elo", all[v].ScoreAccessor())
	}

})).WithTools(tool.NewTool("RationalCognitiveFramework", "Extract RationalCognitiveFramework struct from previous response", func(model *RationalCognitiveFramework) {
	if len(model.Items) == 0 {
		return
	}
	model.Id = redisdb.NanoId(8)
	keyRationalCognitionFramework.HSet(model.Id, model)

}))

func RationalCognitionFrameworkExploration() {
	const numThreads = 12
	const numCallsPerThread = 1000 * 1000 / numThreads

	keyRationalCognitionFramework.HSet(RationalCognitionFrameworkDefault.Id, RationalCognitionFrameworkDefault)
	var wg sync.WaitGroup
	wg.Add(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				best, _ := keyRationalCognitionFramework.HGetAll()
				//未尝试的优先尝试
				NodesWithPriority := []*RationalCognitiveFramework{}
				for k, v := range best {
					if elo := v.ScoreAccessor(); elo > 996 && elo < 1004 {
						NodesWithPriority = append(NodesWithPriority, v)
						delete(best, k)
					}
				}
				selected := lo.Values(best)
				mutable.Shuffle(selected)
				selected = append(NodesWithPriority, selected...)
				selected = selected[:min(6, len(best))]
				param := map[string]any{"RationalCognitionFrameworks": selected}
				err := AgentUtilityFrameGen.WithModels(models.Qwq32B).Call(context.Background(), param)
				if err != nil {
					fmt.Printf("Agent call failed: %v\n", err)
				}
			}
		}()
	}
	wg.Wait()
}

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
	"github.com/doptime/eloevo/prototype"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

// UtilityFunction = exp(WeightMarketSizeln(MarketSize) + 0.18ln(MarketGrowthRate) + 0.22ln(ExpectedReturn) + 0.10ln(TechnicalFeasibility) + 0.15ln(InnovationPotential) + 0.080ln(ResourceAllocation) - 0.12ln(ProjectRisk + 1) - 0.080ln(CompetitionIntensity) - 0.10ln(ImplementationDifficulty) + 0.060ln(TimeToMarket) + 0.040ln(TeamExperience) + 0.050ln(PolicySupport))
type AntiAging struct {
	//prototype.ExporationFramwork
	Id string `description:"required, ID"`
	//RelatedElements        []string `description:"required, 相关要素: 问题背景、要素分析、理论原则、效用函数、实际限制、改进建议、动态反馈机制。"`
	Background                        string `description:"required, Basic background of the problem."`
	ElementsAnalysis                  string `description:"required, Elements analysis of the problem."`
	TheoraticalFirstPrincipleAnalysis string `description:"required, First principles analysis of the problem."`
	//ImprovementInnovations            []string `description:"required, 改进方案: 提出动态分层模块化认知框架，整合第一性原理与实践反馈，建立实时动态调整机制。"`
	UtilityFunction string `description:"required, UtilityFunction A quantitative evaluation function"`

	AgendaForSolutions []string `description:"required, Agenda for the solutions, should be very specific and actionable, such as drink 1 cup of green tea 9:00 am"`
}

func (u *AntiAging) GetId() string { return u.Id }
func (u *AntiAging) ScoreAccessor(delta ...int) int {
	return mixincached.WithElo("projects", "AntiAging", 1000).ScoreAccessor(u.Id, float64(append(delta, 0)[0]))
}

//	.WithTools(tool.NewTool("SaveImprovements", "Save Improvements", func(model *prototype.ExporationImprovement) {
//		// var Improvements = mixincached.HashKeyStrs("projects", keyAntiAging.Key+":ExporationImprovement").WithNItemOnly(5)
//		// func (u *AntiAging) Feedbacks() string {
//		// 	return strings.Join(Improvements.GetInsert(u.Id), "\n")
//		// }
//		// Improvements.GetInsert(model.Id, model.ImprovementSuggestions...)
//	}))
var keyAntiAging = redisdb.NewHashKey[string, *AntiAging](redisdb.Opt.Rds("projects"))

var AgentUtilityFrameGen = agent.NewAgent().WithTemplate(template.Must(template.New("utilifyFunction").Parse(`
现在我们演进面向40岁男性的抗衰老方案，目标是最大化延长寿命，对衰老相关的器官进行有效抗衰老。
所有的解决方案必须来自市场中的可用材料。比如破壁机，益生菌，麦片，绿茶，发酵箱等。

这是现有的若干解决方案范式：
{{range  $index, $item := .AllItems}}
第{{$index}}个解决方案范式：
	Id:{{$item.Id}}
	Background:{{$item.Background}}
	ElementsAnalysis:{{$item.ElementsAnalysis}}
	TheoraticalFirstPrincipleAnalysis:{{$item.TheoraticalFirstPrincipleAnalysis}}
	UtilityFunction:{{$item.UtilityFunction}}
	AgendaForSolutions:
	{{range $item.AgendaForSolutions}}
		{{$item}}
	{{end}}
{{end}}

TaskToDo1:
现在请你仔细讨论这些方案的好坏。将他们从长期可行，能长期使得良好的效抗衰老成为可能的角度。最好到最坏排序(用Id列表表示)。
注意这个prompt会被调用上千次，以便采用Elo淘汰进化出最优方案。方案的深度探索优先；方案不应当广度优先，浅尝辄止。AgendaForSolutions要具体，要易于操作。
所有的字段 Background, ElementsAnalysis, TheoraticalFirstPrincipleAnalysis, UtilityFunction, AgendaForSolutions 都是必填的。
讨论现有方案的优缺点，同样必须涉及对这些内容的讨论。
调用FunctionCall:SaveItemsSorted 保存排序结果。

TaskToDo2:
在详细讨论现有方案的优缺点，在此基础上，尝试提出一个更好的，新的方案。新提出的方案应该继承旧方案的关键优势；并且每次应该就一个确切的主要问题进行深度分析，深度创新以做出实质的改进。
新的方案同样不能遗漏Background, ElementsAnalysis, TheoraticalFirstPrincipleAnalysis, UtilityFunction, AgendaForSolutions等参数。
最后调用FunctionCall:SaveAntiAgingExploration 保存新的方案。

`))).WithToolCallMutextRun().WithTools(tool.NewTool("SaveItemsSorted", "Save sorted Items, Items represented as Id list.", func(model *prototype.ItemsSorted) {
	if model == nil || len(model.ItemsRefById) == 0 {
		return
	}
	all, _ := keyAntiAging.HGetAll()
	playersRanked := []elo.Elo{}
	for _, v := range model.ItemsRefById {
		if _, ok := all[v]; ok {
			playersRanked = append(playersRanked, all[v])
		}
	}
	elo.BatchUpdateRanking(playersRanked...)
	if len(all) > 26 {
		//remove the worst 10% of the projects
		playersOrderByElo := lo.Keys(all)
		//sort the players by elo
		slices.SortFunc(playersOrderByElo, func(i, j string) int {
			return all[i].ScoreAccessor() - all[j].ScoreAccessor()
		})

		//remove the worst 2
		mapToRemove := lo.SliceToMap(playersOrderByElo[:2], func(Id string) (string, *AntiAging) { return Id, all[Id] })
		keyAntiAging.ConcatKey("Expired").HMSet(mapToRemove)
		keyAntiAging.HDel(playersOrderByElo[:2]...)

		//remove the best 1
		if rand.Float64() < 0.05 {
			bestNode := playersOrderByElo[len(playersOrderByElo)-1]
			NowString := time.Now().Format("2006-01-02 15:04:05")
			keyAntiAging.ConcatKey("BestExpired").HSet(NowString, all[bestNode])
			keyAntiAging.HDel(bestNode)
		}

		slices.Reverse(playersOrderByElo)
		for i, v := range playersOrderByElo[:min(5, len(playersOrderByElo))] {
			fmt.Println("Best Model,top ", i+1, v, "Elo", all[v].ScoreAccessor())
		}
	}

})).WithTools(tool.NewTool("SaveAntiAgingExploration", "Extract AntiAging struct from previous response", func(model *AntiAging) {
	model.Id = redisdb.NanoId(8)
	keyAntiAging.HSet(model.Id, model)

}))

func AntiAgingExploration() {
	const numThreads = 6
	const numCallsPerThread = 1000 * 1000 / numThreads
	//projectFunctions, _ = keyProjectsUtilityFunctionAll.HGet("GDvv1dsZ")
	var wg sync.WaitGroup
	wg.Add(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				best, _ := keyAntiAging.HGetAll()
				elos := lo.MapEntries(best, func(Id string, v *AntiAging) (string, elo.Elo) { return Id, v })
				values := elo.SamplingMap(elos, 5)
				param := map[string]any{"AllItems": values, "Task": j % 2}
				err := AgentUtilityFrameGen.WithModels(models.DeepSeekR132B).Call(context.Background(), param)
				if err != nil {
					fmt.Printf("Agent call failed: %v\n", err)
				}
			}
		}()
	}
	wg.Wait()
}

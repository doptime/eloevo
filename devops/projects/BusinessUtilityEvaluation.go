package projects

import (
	"context"
	"fmt"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/elo"
	"github.com/doptime/eloevo/mixincached"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/prototype"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
)

type BusinessUtilityEvaluation struct {
	Id                        string `description:"Id, string, unique"`
	BusinessUtilityEvaluation any    `description:"BusinessUtilityEvaluation, string, with business factors to build a business utility function"`
	Annotation                string `description:"Annotation, string, key consideration for this business utility function"`
}

// "BusinessUtility": "exp(0.35*ln(MarketSize) + 0.30*ln(MarketGrowthRate) + 0.10*ln(ExpectedReturn) + 0.25*ln(TechnicalFeasibility) + 0.20*ln(InnovationPotential) + 0.12*ln(ResourceAllocation) + 0.19*ln(AIAdoptionSynergy) + 0.09*ln(AIAdaptability) + 0.15*ln(DataQuality) + 0.30*ln(EthicalAlignment) + 0.20*ln(SustainabilityImpact) -0.50*ln(ProjectRisk+1) -0.40*ln(CompetitionIntensity) -0.25*ln(ImplementationDifficulty) -0.15*ln(TimeToMarket) +0.25*ln(TeamExperience) +0.23*ln(PolicySupport))",
// "Annotation": "Optimized AGI-era v6.0 holistic model: 1) Prioritizes market dynamics (total 0.65) with emphasis on growth (0.30) and size (0.35). 2) Reinforces ethical alignment (0.30) and sustainability (0.20) as core societal pillars. 3) Introduces adaptive AI readiness (0.28) combining adoption synergy (0.19) and technical adaptability (0.09). 4) Strengthens risk mitigation (-0.50) and competition penalty (-0.40) with logarithmic scaling. 5) Balances technical feasibility (0.25) with innovation potential (0.20). 6) Explicitly weights data quality (0.15) and policy alignment (0.23) for regulatory readiness. 7) Maintains team expertise (0.25) as critical human-AI interface. 8) De-emphasizes short-term returns (0.10) to prioritize long-term value creation through systematic weight redistribution."
func (u *BusinessUtilityEvaluation) GetId() string {
	return u.Id
}
func (u *BusinessUtilityEvaluation) ScoreAccessor(delta ...int) int {
	var eloDelta int = append(delta, 0)[0]
	return mixincached.WithElo("projects", "BusinessUtilityEvaluation", 1000).ScoreAccessor(u.Id, float64(eloDelta))
}

var keyBusinessUtilityEvaluation = redisdb.NewHashKey[string, *BusinessUtilityEvaluation](redisdb.WithRds("projects"))

// 为什么Qwen能自我改进推理，Llama却不行 https://mp.weixin.qq.com/s/OvS61OrDp6rB-R5ELg48Aw
// 并且每次就一个确切的改进方向进行深度分析，反代的深度分析第一性原理之上的需求，深度创新以做出实质的改进。要痛恨泛泛而谈的内容，重复空洞的内容，因为现在是在开发世界级的工具。
var AgentBusinessUtilityEvaluation = agent.NewAgent(template.Must(template.New("utilifyFunction").Parse(`
现在我们要设计AGI时代商业项目的评估函数。目标是在一个以无人化的方式运营几乎一切的业务的世界中，确定值得运营的业务的优先序。

这是现在确定的商业效用函数"BusinessUtility": "exp(0.35*ln(MarketSize) + 0.30*ln(MarketGrowthRate) + 0.10*ln(ExpectedReturn) + 0.25*ln(TechnicalFeasibility) + 0.20*ln(InnovationPotential) + 0.12*ln(ResourceAllocation) + 0.19*ln(AIAdoptionSynergy) + 0.09*ln(AIAdaptability) + 0.15*ln(DataQuality) + 0.30*ln(EthicalAlignment) + 0.20*ln(SustainabilityImpact) -0.50*ln(ProjectRisk+1) -0.40*ln(CompetitionIntensity) -0.25*ln(ImplementationDifficulty) -0.15*ln(TimeToMarket) +0.25*ln(TeamExperience) +0.23*ln(PolicySupport))"
这是注释 "Annotation": "Optimized AGI-era v6.0 holistic model: 1) Prioritizes market dynamics (total 0.65) with emphasis on growth (0.30) and size (0.35). 2) Reinforces ethical alignment (0.30) and sustainability (0.20) as core societal pillars. 3) Introduces adaptive AI readiness (0.28) combining adoption synergy (0.19) and technical adaptability (0.09). 4) Strengthens risk mitigation (-0.50) and competition penalty (-0.40) with logarithmic scaling. 5) Balances technical feasibility (0.25) with innovation potential (0.20). 6) Explicitly weights data quality (0.15) and policy alignment (0.23) for regulatory readiness. 7) Maintains team expertise (0.25) as critical human-AI interface. 8) De-emphasizes short-term returns (0.10) to prioritize long-term value creation through systematic weight redistribution."
你需要讨论并且提出这个评估函数的数据结构，和评估函数的具体表达式。

效用函数带有权重参数的，那你需要提供合理的权重参数值。如果已经有权重参数值，那你需要进一步改进这些参数值。

这是现有的分类 ：
{{range  $index, $item := .ItemList}}
第{{$index}}个认知范式：
	Id:{{$item.Id}}
	BusinessUtilityEvaluation:{{$item.BusinessUtilityEvaluation}}
	Annotation:{{$item.Annotation}}
{{end}}


ToDo:
1. 对现有的工作进行思考和评估：
	1、对回溯或在检测到错误进行显式修改（例如，「这种方法行不通，因为...」）；
	2、验证或系统地检查中间结果（例如，「让我们通过... 来验证这个结果」）；
	3、子目标设定，即将复杂问题分解为可管理的步骤（例如，「要解决这个问题，我们首先需要...」）；
	4、逆向思考，即在目标导向的推理问题中，从期望的结果出发，逐步向后推导，找到解决问题的路径。（例如，「要达到 75 的目标，我们需要一个能被... 整除的数字」）。 
2. 根据上面的讨论. 将现有方案从最好到最坏排序(用Id列表表示)。调用FunctionCall:SaveItemsSorted 保存排序结果。
{{if lt .TotoalNodes 10}}
步骤3. 在讨论的基础上，进一步提出一个更好的，新的方案草稿。对草稿同样按步骤1改进后，调用FunctionCall:BusinessUtilityEvaluation 保存新的方案。
{{end}}
`))).WithToolCallLocked().WithTools(tool.NewTool("SaveItemsSorted", "Save sorted Items, Items represented as Id list.", func(model *prototype.ItemsSorted) {
	if model == nil || len(model.ItemsRefById) == 0 {
		return
	}
	all, _ := keyBusinessUtilityEvaluation.HGetAll()
	sortedElos := elo.ToSlice(all).Sort().Reverse()
	elo.BatchUpdateRanking(sortedElos.TakeByIds(model.ItemsRefById...)...)

	for _, v := range lo.Filter(sortedElos, func(v elo.Elo, i int) bool { return v.ScoreAccessor() < 996 }) {
		keyBusinessUtilityEvaluation.ConcatKey("Expired").HSet(v.GetId(), v)
		keyBusinessUtilityEvaluation.HDel(v.GetId())
	}
	for i, v := range sortedElos[0:min(10, len(sortedElos))] {
		fmt.Println("Best Model,top ", i+1, v.GetId(), "Elo", v.ScoreAccessor())
	}

})).WithTools(tool.NewTool("NewBusinessUtilityEvaluation", "create a NewBusinessUtilityEvaluation plan", func(model *BusinessUtilityEvaluation) {
	model.Id = redisdb.NanoId(8)
	keyBusinessUtilityEvaluation.HSet(model.Id, model)
}))

func BusinessUtilityEvaluationExploration() {
	keyBusinessUtilityEvaluation.HSet("default", &BusinessUtilityEvaluation{Id: "default", BusinessUtilityEvaluation: "UtilityFunction = exp(WeightMarketSizeln(MarketSize) + 0.18ln(MarketGrowthRate) + 0.22ln(ExpectedReturn) + 0.10ln(TechnicalFeasibility) + 0.15ln(InnovationPotential) + 0.080ln(ResourceAllocation) - 0.12ln(ProjectRisk + 1) - 0.080ln(CompetitionIntensity) - 0.10ln(ImplementationDifficulty) + 0.060ln(TimeToMarket) + 0.040ln(TeamExperience) + 0.050ln(PolicySupport))"})
	const MaxThreads = 16
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	for i, TotalTasks := 0, 1000*1000; i < TotalTasks; i++ {
		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		go func() {
			defer func() { <-MaxThreadsSemaphore }()
			best, _ := keyBusinessUtilityEvaluation.HGetAll()
			//put new nodes to the top
			NodesWithPriority := lo.Filter(lo.Values(best), func(v *BusinessUtilityEvaluation, i int) bool {
				return v.ScoreAccessor() > 996 && v.ScoreAccessor() < 1004
			})
			selected := lo.Union(NodesWithPriority, lo.Values(best))
			mutable.Shuffle(selected[len(NodesWithPriority):])
			param := map[string]any{"ItemList": selected[:min(6, len(selected))],
				"Model":       []*models.Model{models.Qwq32B, models.DeepSeekR132B}[i%2],
				"TotoalNodes": len(best),
			}
			err := AgentBusinessUtilityEvaluation.WithModel(models.Qwq32B).Call(context.Background(), param)
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

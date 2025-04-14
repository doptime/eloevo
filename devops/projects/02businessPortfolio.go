package projects

import (
	"context"
	"fmt"
	"math"
	"slices"
	"sync"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

const TOP_N_INDIVIDUALS = 4000

type BusinessPlan struct {
	Id     string `description:"required, Unique identifier for the BusinessPlan"`
	Name   string `description:"required, Name of the BusinessPlan"`
	Detail string `description:"required, Detail of the BusinessPlan"`

	MarketSize                            float64 `description:"required, Log scale based on revenue in USD from government stats and third-party surveys"`
	BusinessUtilization                   float64
	MarketGrowthRate                      float64            `description:"required, Past 5 years' Compound Annual Growth Rate (CAGR)"`
	ExpectedReturn                        float64            `description:"required, Internal Rate of Return (IRR) or Net Present Value (NPV)"`
	TechnicalFeasibility                  float64            `description:"required, 0-1 scale based on AI assessment of technology maturity"`
	InnovationPotential                   float64            `description:"required, 0-1 scale based on patent filings and technological uniqueness"`
	ResourceAllocation                    float64            `description:"required, Budget in USD as a percentage of total company budget"`
	ProjectRisk                           float64            `description:"required, 0-1 scale based on risk factor assessment"`
	CompetitionIntensity                  float64            `description:"required, Herfindahl-Herschman Index (HHI) for market concentration"`
	ImplementationDifficulty              float64            `description:"required, 0-1 scale based on task complexity and resource availability"`
	TimeToMarket                          float64            `description:"required, Expected time to market in months"`
	TeamExperience                        float64            `description:"required, Average years of experience of team members"`
	PolicySupport                         float64            `description:"required, Government policy support score (1-10)"`
	BusinessUtilityAttributionCoefficient map[string]float64 `description:"required, Proportion of current BusinessPlan‘s Utility that should contributes to the other BusinessPlan。 with key the destination BusinessPlan Id and value the weight the attribution coefficient "`
}

func (u *BusinessPlan) GetId() string {
	return u.Id
}
func (b *BusinessPlan) Utility() float64 {
	return math.Exp(
		0.25*math.Log(b.MarketSize) +
			0.18*math.Log(b.MarketGrowthRate) +
			0.22*math.Log(b.ExpectedReturn) +
			0.10*math.Log(b.TechnicalFeasibility) +
			0.15*math.Log(b.InnovationPotential) +
			0.08*math.Log(b.ResourceAllocation) -
			0.12*math.Log(b.ProjectRisk+1) -
			0.08*math.Log(b.CompetitionIntensity) -
			0.10*math.Log(b.ImplementationDifficulty) +
			0.06*math.Log(b.TimeToMarket) +
			0.04*math.Log(b.TeamExperience) +
			0.05*math.Log(b.PolicySupport),
	)
}
func (b *BusinessPlan) AdjustedUtility(plans map[string]*BusinessPlan) float64 {
	base := b.Utility()
	var prSum float64 = b.BusinessUtilization

	// 计算其他 BusinessPlan 对当前 BusinessPlan 的效用贡献
	for _, plan := range plans {
		if plan.BusinessUtilityAttributionCoefficient != nil {
			if weight, exists := plan.BusinessUtilityAttributionCoefficient[b.Id]; exists {
				prSum += weight * plan.Utility() // 累加其他计划对当前计划的效用贡献
			}
		}
	}

	return base + prSum
}
func PruneOnePlan() {
	if len(BusinessPlanMap) < TOP_N_INDIVIDUALS {
		return
	}

	//remove the worst project
	playersOrderByElo := lo.Values(BusinessPlanMap)
	//sort the players by elo
	slices.SortFunc(playersOrderByElo, func(i, j *BusinessPlan) int {
		return int(i.AdjustedUtility(BusinessPlanMap) - j.AdjustedUtility(BusinessPlanMap))
	})
	//remove the worst 1
	fieldsToRemove := lo.Map(playersOrderByElo[:1], func(v *BusinessPlan, i int) string { return v.Id })
	mapToRemove := map[string]*BusinessPlan{}
	for _, v := range fieldsToRemove {
		mapToRemove[v] = BusinessPlanMap[v]
	}
	keyBusinessPlanExpired.HMSet(mapToRemove)
	for _, v := range fieldsToRemove {
		delete(BusinessPlanMap, v)
	}
	keyBusinessPlan.HDel(fieldsToRemove...)
	//save other UtilityFunctionExploration
	keyBusinessPlan.HMSet(BusinessPlanMap)

	slices.Reverse(playersOrderByElo)
	for i, v := range playersOrderByElo {
		fmt.Println("Best Model,top ", i+1, v.Id, "Elo")
	}
}

var AgentGenNicheMarketOpportunity = agent.NewAgent(template.Must(template.New("question").Parse(`
请作为一位创新思维专家, 全面探索AGI时代，一切都即将开始无人化运营后，将要出现的应用领域及其盈利机会.

生成的商业函数将采用这个效用函数:	UtilityFunction = exp(WeightMarketSizeln(MarketSize) + 0.18ln(MarketGrowthRate) + 0.22ln(ExpectedReturn) + 0.10ln(TechnicalFeasibility) + 0.15ln(InnovationPotential) + 0.080ln(ResourceAllocation) - 0.12ln(ProjectRisk + 1) - 0.080ln(CompetitionIntensity) - 0.10ln(ImplementationDifficulty) + 0.060ln(TimeToMarket) + 0.040ln(TeamExperience) + 0.050ln(PolicySupport))

These are current Business plans:
{{range .BusinessPlans}}
	Id:{{.Id}}
	Name:{{.Name}}
	Detail:{{.Detail}}
	BusinessUtilization:{{.BusinessUtilization}}
{{end}}

步骤：
	总说明：按照Functional Tools:UtilityFunctionExplorationFinal 中的效用函数中的评估标准对步骤1、2中的分析结果进行保存。
	1. 请仔细深度探讨现有机会方案的分布,如果现有商业机会的BusinessUtilityAttributionCoefficient 估计严重失误。那么请对现有商业机会重新修正评估。
	2. 如果不需要修复现有商业机会的评估。那么请提出一个新的商业机会。新提出的商业方案和现有机会高度相关，和现有商业方案共同构成联合商业效用最大化。也就是说，从价值链的视角，要求这个新提出的商业机会能够最大化地依赖现有的商业机会，或者现有的机会能够最大化地依赖新的商业机会。使得全部的商业机会集合的效用最大化。
	3. 调用UtilityFunctionExplorationFinal，整理保持1、2步骤中的商业机会和细节。
	

   `))).WithTools(tool.NewTool("UtilityFunctionExplorationFinal", "Extract UtilityFunctionExploration struct from previous response", func(model *BusinessPlan) {
	if len(model.Name) == 0 || len(model.Detail) == 0 {
		return
	}

	if current, ok := BusinessPlanMap[model.Id]; !ok {
		model.Id = redisdb.NanoId(8)
		model.BusinessUtilityAttributionCoefficient = map[string]float64{}
	} else {
		//merge the BusinessUtilityAttributionCoefficient
		learningRate := 0.3
		for id, w := range model.BusinessUtilityAttributionCoefficient {
			//gradient descent
			if curWeight, ok := current.BusinessUtilityAttributionCoefficient[id]; ok {
				model.BusinessUtilityAttributionCoefficient[id] = curWeight + (w-curWeight)*learningRate
			}
		}
		model.MarketSize = current.MarketSize + (model.MarketSize-current.MarketSize)*learningRate
		model.MarketGrowthRate = current.MarketGrowthRate + (model.MarketGrowthRate-current.MarketGrowthRate)*learningRate
		model.ExpectedReturn = current.ExpectedReturn + (model.ExpectedReturn-current.ExpectedReturn)*learningRate
		model.TechnicalFeasibility = current.TechnicalFeasibility + (model.TechnicalFeasibility-current.TechnicalFeasibility)*learningRate
		model.InnovationPotential = current.InnovationPotential + (model.InnovationPotential-current.InnovationPotential)*learningRate
		model.ResourceAllocation = current.ResourceAllocation + (model.ResourceAllocation-current.ResourceAllocation)*learningRate
		model.ProjectRisk = current.ProjectRisk + (model.ProjectRisk-current.ProjectRisk)*learningRate
		model.CompetitionIntensity = current.CompetitionIntensity + (model.CompetitionIntensity-current.CompetitionIntensity)*learningRate
		model.ImplementationDifficulty = current.ImplementationDifficulty + (model.ImplementationDifficulty-current.ImplementationDifficulty)*learningRate
		model.TimeToMarket = current.TimeToMarket + (model.TimeToMarket-current.TimeToMarket)*learningRate
		model.TeamExperience = current.TeamExperience + (model.TeamExperience-current.TeamExperience)*learningRate
		model.PolicySupport = current.PolicySupport + (model.PolicySupport-current.PolicySupport)*learningRate
		model.BusinessUtilization = current.BusinessUtilization + (model.BusinessUtilization-current.BusinessUtilization)*learningRate

	}

	for id := range model.BusinessUtilityAttributionCoefficient {
		if _, ok := BusinessPlanMap[id]; !ok {
			delete(model.BusinessUtilityAttributionCoefficient, id)
		}
	}
	BusinessPlanMap[model.Id] = model
	keyBusinessPlan.HSet(model.Id, model)

	PruneOnePlan()

}))

var BusinessPlanMap = map[string]*BusinessPlan{}
var keyBusinessPlan = redisdb.NewHashKey[string, *BusinessPlan](redisdb.WithRds("projects"))
var keyBusinessPlanExpired = redisdb.NewHashKey[string, *BusinessPlan](redisdb.WithKey("BusinessPlanExpired"), redisdb.WithRds("projects"))

// GenNicheMarketOpportunityParallel calls GenNicheMarketOpportunity 1000 times in 16 parallel threads.
func GenBusinessPlanParallel() {
	const numThreads = 1
	const numCallsPerThread = 4000

	var wg sync.WaitGroup
	wg.Add(numThreads)

	BusinessPlanMap, _ = keyBusinessPlan.HGetAll()

	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				param := map[string]any{"BusinessPlans": utils.GetRandomNValueFromMap(BusinessPlanMap, 10)}
				err := AgentGenNicheMarketOpportunity.WithModels(models.FuseO1).Call(context.Background(), param)
				if err != nil {
					fmt.Println(err)
				}
			}
		}()
	}
	wg.Wait()
}

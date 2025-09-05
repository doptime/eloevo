package projects

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/elo"
	"github.com/doptime/eloevo/mixincached"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
)

// FactorMeasurement 定义了用于评估项目的因素的度量标准。
type FactorMeasurement struct {
	MarketSize               string `json:"marketSize"`
	MarketGrowthRate         string `json:"marketGrowthRate"`
	ExpectedReturn           string `json:"expectedReturn"`
	TechnicalFeasibility     string `json:"technicalFeasibility"`
	InnovationPotential      string `json:"innovationPotential"`
	ResourceAllocation       string `json:"resourceAllocation"`
	ProjectRisk              string `json:"projectRisk"`
	CompetitionIntensity     string `json:"competitionIntensity"`
	ImplementationDifficulty string `json:"implementationDifficulty"`
	TimeToMarket             string `json:"timeToMarket"`
	TeamExperience           string `json:"teamExperience"`
	PolicySupport            string `json:"policySupport"`
}

func (f *FactorMeasurement) String() string {
	return fmt.Sprintf("MarketSize: %v, MarketGrowthRate: %v, ExpectedReturn: %v, TechnicalFeasibility: %v, InnovationPotential: %v, ResourceAllocation: %v, ProjectRisk: %v, CompetitionIntensity: %v, ImplementationDifficulty: %v, TimeToMarket: %v, TeamExperience: %v, PolicySupport: %v", f.MarketSize, f.MarketGrowthRate, f.ExpectedReturn, f.TechnicalFeasibility, f.InnovationPotential, f.ResourceAllocation, f.ProjectRisk, f.CompetitionIntensity, f.ImplementationDifficulty, f.TimeToMarket, f.TeamExperience, f.PolicySupport)
}

// FactorWeights 定义了各个评估因素的权重。
type FactorWeights struct {
	WeightMarketGrowthRate         float32
	WeightExpectedReturn           float32
	WeightTechnicalFeasibility     float32
	WeightInnovationPotential      float32
	WeightResourceAllocation       float32
	WeightProjectRisk              float32
	WeightCompetitionIntensity     float32
	WeightImplementationDifficulty float32
	WeightTimeToMarket             float32
	WeightTeamExperience           float32
	WeightPolicySupport            float32
}

func (f *FactorWeights) String() string {
	return fmt.Sprintf("WeightMarketGrowthRate: %v, WeightExpectedReturn: %v, WeightTechnicalFeasibility: %v, WeightInnovationPotential: %v, WeightResourceAllocation: %v, WeightProjectRisk: %v, WeightCompetitionIntensity: %v, WeightImplementationDifficulty: %v, WeightTimeToMarket: %v, WeightTeamExperience: %v, WeightPolicySupport: %v", f.WeightMarketGrowthRate, f.WeightExpectedReturn, f.WeightTechnicalFeasibility, f.WeightInnovationPotential, f.WeightResourceAllocation, f.WeightProjectRisk, f.WeightCompetitionIntensity, f.WeightImplementationDifficulty, f.WeightTimeToMarket, f.WeightTeamExperience, f.WeightPolicySupport)
}

// EvaluationSchema 代表一个项目评估模式，包含度量标准、权重、以及 Elo 评分。
type EvaluationSchema struct {
	Id                      string
	FactorMeasurement       *FactorMeasurement
	FactorWeights           *FactorWeights
	PresviousProposalSorted []string
}

func (u *EvaluationSchema) String() string {
	return fmt.Sprintf("Id: %v, FactorMeasurement: %v, FactorWeights: %v", u.Id, u.FactorMeasurement, u.FactorWeights)
}

// GetId 获取 EvaluationSchema 的 ID。
func (u *EvaluationSchema) GetId() string {
	return u.Id
}

// Rating 更新 EvaluationSchema 的 Elo 评分。
func (u *EvaluationSchema) ScoreAccessor(delta ...int) int {
	var eloDelta int = append(delta, 0)[0]
	return mixincached.WithElo("projects", "EvaluationSchema", 1200).ScoreAccessor(u.Id, float64(eloDelta))
}

// activeSchemaKey 是用于存储活跃 EvaluationSchema 的 Redis Hash Key。
var activeSchemaKey = redisdb.NewHashKey[string, *EvaluationSchema](redisdb.Opt.Rds("projects"))

// expiredSchemaKey 是用于存储过期 EvaluationSchema 的 Redis Hash Key。
var expiredSchemaKey = redisdb.NewHashKey[string, *EvaluationSchema](redisdb.Opt.Rds("projects").Key("UtilityFunctionSchemaExpired"))

// evaluationSchemaMutex 用于保护 evaluationSchemas 的并发访问。
var evaluationSchemaMutex sync.Mutex = sync.Mutex{}

// evaluationSchemaAgent 是负责生成和改进 EvaluationSchema 的 Agent。
var evaluationSchemaAgent = agent.NewAgent().WithTemplate(template.Must(template.New("evaluationSchemaPrompt").Parse(`
这是一个AGI 给出的时代商业项目的评估函数：
UtilityFunction = exp(w1*ln(MarketSize) + WeightMarketGrowthRate*ln(MarketGrowthRate) + WeightExpectedReturn*ln(ExpectedReturn) + WeightTechnicalFeasibility*ln(TechnicalFeasibility) + WeightInnovationPotential*ln(InnovationPotential) + WeightResourceAllocation*ln(ResourceAllocation) - WeightProjectRisk*ln(ProjectRisk + 1) - WeightCompetitionIntensity*ln(CompetitionIntensity) - WeightImplementationDifficulty*ln(ImplementationDifficulty) + WeightTimeToMarket*ln(TimeToMarket) + WeightTeamExperience*ln(TeamExperience) + WeightPolicySupport*ln(PolicySupport))

UtilityFunction 用以在一个以无人化的方式运营几乎一切的业务的世界中，确定运营业务的优先序。

你需要讨论并且提出这个评估函数中不同的参数的度量办法 和参数的权重。
期望的效果：
	- 效用函数带有权重参数的，需要仔细思考并改进参数值的合理性
	- 所提出参数的评估标准（FactorMeasurement）要包括明确清晰的量化标准，以便所有的项目都可以不依赖外部专家，而是可以依赖该标准，无需依赖第三方，直接进行清晰的量化评估

这是现有的若干评估商业效用 的 FactorMeasurement+FactorWeights
{{.previousSchemaProposals}}

ToDoList:
1. 现在请你仔细讨论这些方案，将他们从最好到最坏排序(用Id列表表示)。
2. 尝试提出一个更好的，新的适用这个UtilityFunction 的 FactorMeasurement+FactorWeights 方案。 度量办法要求任何的方案，可以依据标准直接做出判断。

在完成讨论和提出新的评估方案后，请将ToDoList中的两项工作（方案排序和新方案）， 通过 Toolcall SaveEvaluationSchema 保存。
`))).WithTools(tool.NewTool("SaveEvaluationSchema", "Extract UtilityFunctionSchemaExploration struct according to analysis", func(model *EvaluationSchema) {
	if model.FactorMeasurement == nil || model.FactorWeights == nil {
		return
	}
	if len(model.PresviousProposalSorted) == 0 {
		return
	}
	model.Id = redisdb.NanoId(8)
	evaluationSchemas[model.Id] = model
	activeSchemaKey.HSet(model.Id, model)
	evaluationSchemaMutex.Lock()
	defer evaluationSchemaMutex.Unlock()

	playersRanked := []elo.Elo{}
	for _, v := range model.PresviousProposalSorted {
		if _, ok := evaluationSchemas[v]; ok {
			playersRanked = append(playersRanked, evaluationSchemas[v])
		}
	}

	elo.BatchUpdateRanking(playersRanked...)
	if len(evaluationSchemas) > topSchemaCount*4 {

		// 移除最差的 10% 的项目 (这里简化为移除最差的 2 个)
		playersOrderByElo := lo.Values(evaluationSchemas)
		// 根据 Elo 评分排序
		slices.SortFunc(playersOrderByElo, func(i, j *EvaluationSchema) int {
			return int(i.ScoreAccessor() - j.ScoreAccessor())
		})
		// 移除最差的 2 个
		fieldsToRemove := lo.Map(playersOrderByElo[:2], func(v *EvaluationSchema, i int) string { return v.Id })
		mapToRemove := map[string]*EvaluationSchema{}
		for _, v := range fieldsToRemove {
			mapToRemove[v] = evaluationSchemas[v]
		}

		expiredSchemaKey.HMSet(mapToRemove)
		for _, v := range fieldsToRemove {
			delete(evaluationSchemas, v)
		}
		activeSchemaKey.HDel(fieldsToRemove...)
		// 保存其他的 EvaluationSchema
		activeSchemaKey.HMSet(evaluationSchemas)

		slices.Reverse(playersOrderByElo)
		for i, v := range playersOrderByElo {
			fmt.Println("Best Model,top ", i+1, v.Id, "Elo", v.ScoreAccessor())
		}
	}

}))

// evaluationSchemas 存储 EvaluationSchema 的集合。
var evaluationSchemas = make(map[string]*EvaluationSchema)

// topSchemaCount 定义了保留的最佳方案数量。
const topSchemaCount = 10 // 假设保留最佳方案数量为 10

// RefineEvaluationSchemas 负责定期进化和改进 EvaluationSchema。
func RefineEvaluationSchemas() {
	const numThreads = 12
	const numCallsPerThread = 1000 * 1000 / numThreads

	evaluationSchemas, _ = activeSchemaKey.HGetAll()

	var wg sync.WaitGroup
	wg.Add(numThreads)
	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				items := lo.Values(evaluationSchemas)
				mutable.Shuffle(items)
				if len(items) > topSchemaCount {
					items = items[:topSchemaCount]
				}
				var ret strings.Builder
				for i, v := range items {
					ret.WriteString(fmt.Sprintf("第%v个评估方案：\n%v\n", i, v.String()))
				}
				err := evaluationSchemaAgent.WithModels(models.FuseO1).Call(context.Background(), map[string]any{
					"previousSchemaProposals": ret.String(), // 使用重命名后的变量名
					"NewId":                   redisdb.NanoId(8),
				})
				if err == nil {
					continue
				}
				fmt.Printf("Agent call failed: %v\n", err)
			}
		}()
	}

	wg.Wait()
}

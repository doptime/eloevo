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

// UtilityFunction = exp(WeightMarketSizeln(MarketSize) + 0.18ln(MarketGrowthRate) + 0.22ln(ExpectedReturn) + 0.10ln(TechnicalFeasibility) + 0.15ln(InnovationPotential) + 0.080ln(ResourceAllocation) - 0.12ln(ProjectRisk + 1) - 0.080ln(CompetitionIntensity) - 0.10ln(ImplementationDifficulty) + 0.060ln(TimeToMarket) + 0.040ln(TeamExperience) + 0.050ln(PolicySupport))
type UtilityFunctionExploration struct {
	Id                      string
	ProjectDescription      map[string]string
	UtilityFunction         string
	PresviousProposalSorted []string
}

func (u *UtilityFunctionExploration) GetId() string {
	return u.Id
}
func (u *UtilityFunctionExploration) ScoreAccessor(delta ...int) int {
	var eloDelta int = append(delta, 0)[0]
	return mixincached.WithElo("projects", "UtilityFunctionExploration", 1000).ScoreAccessor(u.Id, float64(eloDelta))
}

func (u *UtilityFunctionExploration) String() string {
	var ProjectDescription strings.Builder
	for k, v := range u.ProjectDescription {
		ProjectDescription.WriteString(fmt.Sprintf("\t%v: \"%v\",\n", k, v))
	}
	ProjectDescriptionString := ProjectDescription.String()
	return fmt.Sprintf("Id: %v, ProjectDescription: \t{\n%s\t}, UtilityFunction: %s", u.Id, ProjectDescriptionString, u.UtilityFunction)
}
func UtilityFunctionExplorationList(us []*UtilityFunctionExploration) string {
	var ret strings.Builder
	for i, v := range us {
		ret.WriteString(fmt.Sprintf("第%v个评估函数：\n%v\n", i, v.String()))
	}
	return ret.String()
}

const ProjectsUtilityFunctionTopN = 10

var keyProjectsUtilityFunction = redisdb.NewHashKey[string, *UtilityFunctionExploration](redisdb.Opt.Rds("projects"))
var AgentBusinessUtilityFunctionGen = agent.NewAgent().WithTemplate(template.Must(template.New("utilifyFunction").Parse(`
现在你要设计AGI时代商业项目的评估函数。目标是在一个以无人化的方式运营几乎一切的业务的世界中，确定运营业务的优先序。

你需要讨论并且提出这个评估函数的数据结构，和评估函数的具体表达式。

效用函数带有权重参数的，那你需要提供合理的权重参数值。如果已经有权重参数值，那你需要进一步改进这些参数值。

这是现有的若干评估商业效用 的 数据结构+函数的描述结构体：
{{.UtilityFunctionExplorations}}



ToDoList:
1. 现在请你仔细讨论这些方案，将他们从最好到最坏排序(用Id列表表示)。
2. 尝试提出一个更好的，新的评估方案的数据结构和评估函数的具体表达式。

在完成讨论和提出新的评估方案后，请将ToDoList中的两项工作（方案排序和新方案）， 通过 Toolcall: UtilityFunctionExplorationFinal 保存。
`))).WithToolCallMutextRun().WithTools(tool.NewTool("UtilityFunctionExplorationFinal", "Extract UtilityFunctionExploration struct from previous response", func(model *UtilityFunctionExploration) {
	if len(model.ProjectDescription) == 0 || len(model.UtilityFunction) == 0 {
		return
	}
	if len(model.PresviousProposalSorted) == 0 {
		return
	}
	model.Id = redisdb.NanoId(8)
	projectFunctions[model.Id] = model
	keyProjectsUtilityFunction.HSet(model.Id, model)
	playersRanked := []elo.Elo{}
	for _, v := range model.PresviousProposalSorted {
		if _, ok := projectFunctions[v]; ok {
			playersRanked = append(playersRanked, projectFunctions[v])
		}
	}
	elo.BatchUpdateRanking(playersRanked...)
	if len(projectFunctions) > ProjectsUtilityFunctionTopN*2 {
		//remove the worst 10% of the projects
		playersOrderByElo := lo.Values(projectFunctions)
		//sort the players by elo
		slices.SortFunc(playersOrderByElo, func(i, j *UtilityFunctionExploration) int {
			return int(i.ScoreAccessor() - j.ScoreAccessor())
		})
		//remove the worst 2
		fieldsToRemove := lo.Map(playersOrderByElo[:2], func(v *UtilityFunctionExploration, i int) string { return v.Id })
		mapToRemove := map[string]*UtilityFunctionExploration{}
		for _, v := range fieldsToRemove {
			mapToRemove[v] = projectFunctions[v]
		}
		keyProjectsUtilityFunction.ConcatKey("Expired").HMSet(mapToRemove)
		for _, v := range fieldsToRemove {
			delete(projectFunctions, v)
		}
		keyProjectsUtilityFunction.HDel(fieldsToRemove...)
		//save other UtilityFunctionExploration
		keyProjectsUtilityFunction.HMSet(projectFunctions)

		slices.Reverse(playersOrderByElo)
		for i, v := range playersOrderByElo {
			fmt.Println("Best Model,top ", i+1, v.Id, "Elo", v.ScoreAccessor())
		}
	}

}))
var projectFunctions = make(map[string]*UtilityFunctionExploration)

func EvoUtilityFunctionExploration() {
	const numThreads = 12
	const numCallsPerThread = 1000 * 1000 / numThreads
	best, _ := keyProjectsUtilityFunction.HGet("GDvv1dsZ")
	projectFunctions[best.Id] = best
	var wg sync.WaitGroup
	wg.Add(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				items := lo.Values(projectFunctions)
				mutable.Shuffle(items)
				if len(items) > ProjectsUtilityFunctionTopN {
					items = items[:ProjectsUtilityFunctionTopN]
				}
				err := AgentBusinessUtilityFunctionGen.WithModels(models.FuseO1).Call(context.Background(), map[string]any{
					"UtilityFunctionExplorations": UtilityFunctionExplorationList(items),
					"NewId":                       redisdb.NanoId(8),
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

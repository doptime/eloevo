package devops

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"text/template"

	"github.com/cespare/xxhash/v2"
	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/elo"
	"github.com/doptime/eloevo/prototype"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

type UseCase struct {
	Id    string  `description:"Id, string, unique"`
	Score float64 `msgpack:"Elo" description:"-"`
	Item  string  `description:"Item,string"`
}

func (u *UseCase) GetId() string {
	return u.Id
}
func (u *UseCase) Quantile(delta ...float64) float64 {
	const learningRate = 0.2
	if eloDelta := append(delta, 0)[0]; eloDelta != 0 {
		u.Score += learningRate * eloDelta
		keyUseCase.HSet(u.Id, u)
	}
	return u.Score
}

var keyUseCase = redisdb.NewHashKey[string, *UseCase](redisdb.Opt.Rds("usecase").Key("法规遵从AI助手服务"))

// 为什么Qwen能自我改进推理，Llama却不行 https://mp.weixin.qq.com/s/OvS61OrDp6rB-R5ELg48Aw
// 并且每次就一个确切的改进方向进行深度分析，反代的深度分析第一性原理之上的需求，深度创新以做出实质的改进。要痛恨泛泛而谈的内容，重复空洞的内容，因为现在是在开发世界级的工具。
var AgentUseCaseGen = agent.NewAgent(template.Must(template.New("utilifyFunction").Parse(`
现在我们要从软件工程的角度，完成对一个商业项目的需求建模。要求生成完备的覆盖全部需求场景的用例集合 Use Cases。
场景必须覆盖以下几个方面，这些方面的权重重要性数值已经给出。用例的描述必须符合这些方面的权重重要性数值。
{
  "市场需求: 确保项目有明确的市场需求和增长潜力": 0.15,
  "产品/服务: 产品的独特性、技术可行性和用户体验至关重要": 0.15,
  "商业模式: 盈利模式和成本结构的合理性直接影响项目的可持续性": 0.15,
  "团队能力: 强大的团队是执行和应对挑战的基础": 0.15,
  "竞争与壁垒: 了解竞争环境并建立竞争优势，确保项目能够在市场中立足": 0.1,
  "财务与资源: 充足的资金和资源是项目启动和扩展的保障": 0.1,
  "法律与合规: 确保项目在法律框架内运作，避免潜在的法律风险": 0.05,
  "市场进入策略: 有效的市场进入策略能够加速项目的市场渗透": 0.05,
  "技术与创新: 技术优势和创新能力能够提升项目的竞争力": 0.05,
  "用户获取与营销: 高效的用户获取和营销策略有助于快速扩大用户基础": 0.05,
  "可持续性与社会影响: 确保项目具备长期可持续发展，并产生积极的社会影响": 0.03,
  "风险管理: 有效的风险管理能够降低项目失败的可能性": 0.02
}

现有的项目是：
为特定国家或地区的企业和个人提供本地化、定制化的法规遵从AI助手服务。 用户可以动态追加与法规相关的需求描述。系统会自动给出最佳的法规遵从建议。

这是现有的用例方案/(用例列表)：
{{range  $index, $item := .ItemList}}
"Id":"{{$item.Id}}"
"Item":"{{$item.Item}}"
{{end}}

ToDo:
步骤1. 对现有的选项进行思考和评估：
	1、对回溯或在检测到错误进行显式修改；
	2、验证或系统地检查中间结果；
	3、子目标设定，即将复杂问题分解为可管理的步骤
	4、逆向思考，即在目标导向的推理问题中，从期望的结果出发，逐步向后推导，找到解决问题的路径。

步骤2. 根据上面流程对现有方案进行讨论.
	- 在讨论的基础上，如果发现方案存在被改进的可能，请进一步深度推理，以便提出一个（注意有且仅有一个）更好的Use Case。其内容稍后保存到NewProposedItemContent字段中。
	- 选出必须淘汰的Use case条目，比如冗余，劣质，条目，形成淘汰列表ItemsShouldRemoveFromSolutionSorted，按淘汰优先级别列出。
	- 选出目标方案条目，形成解决方案列表ItemsShouldKeptInSolutionSorted，按保留优先级列出。
最后调用FunctionCall:SolutionRefine 保存排序结果。

`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionRefine", "Save sorted Items, Items represented as Id list.", func(model *prototype.SolutionRefinement_v250319) {
	if model == nil || len(model.ItemIdsShouldKeptInSolution) == 0 {
		return
	}
	all, _ := keyUseCase.HGetAll()
	sortedElos := elo.ToQuantileSlice(all).Sort()
	p1 := sortedElos.TakeByIds(model.ItemIdsShouldKeptInSolution...)
	slices.Reverse(model.ItemIdsShouldRemoveFromSolution)
	p3 := sortedElos.TakeByIds(model.ItemIdsShouldRemoveFromSolution...)
	p2, _ := lo.Difference(sortedElos, p1)
	p2, _ = lo.Difference(p2, p3)
	elo.UpdateQuantile(lo.Uniq(append(p1, p2...))...)
	elo.UpdateQuantilTo(1.2, p3...)

	if model.NewProposedItemContent != "" {
		item := &UseCase{Score: 1 - 0.15, Item: model.NewProposedItemContent}
		item.Id = big.NewInt(int64(xxhash.Sum64String(model.NewProposedItemContent))).Text(62)[:6]
		keyUseCase.HSet(item.Id, item)
	}

}))

func UseCaseExploration() {

	const MaxThreads = 1
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	for i, TotalTasks := 0, 1000*1000; i < TotalTasks; i++ {
		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		best, _ := keyUseCase.HGetAll()
		listSorted := elo.ToQuantileSlice(lo.Values(best)).Sort()
		//print the lefts
		for i, v := range listSorted {
			fmt.Println("Rank", i+1, " ", v.GetId(), "‰", int(1000*v.Quantile()), " ", v.(*UseCase).Item)
		}
		//remove the worst 1
		if len(listSorted) > 1 {
			if worst := listSorted[len(listSorted)-1]; worst.Quantile() > 1.0 {
				keyUseCase.ConcatKey("Expired").HSet(worst.GetId(), worst)
				keyUseCase.HDel(worst.GetId())
				listSorted = listSorted[:len(listSorted)-1]
			}
		}

		param := map[string]any{
			"ItemList":    listSorted,
			"TotoalNodes": len(best),
		}
		go func(param map[string]any) {
			defer func() { <-MaxThreadsSemaphore }()
			err := AgentUseCaseGen.Call(context.Background(), param)
			if err != nil {
				fmt.Printf("Agent call failed: %v\n", err)
			}
		}(param)
	}
	// Wait for all the goroutines to finish)
	for i := 0; i < MaxThreads; i++ {
		MaxThreadsSemaphore <- struct{}{}
	}

}

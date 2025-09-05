package projects

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/prototype"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
)

type BusinessPlansEdTech struct {
	Id       string `description:"Id, string, unique"`
	Votes    int64  `description:"-"`
	Item     string `description:"item of the solution"`
	ParentId string `description:"Parent Id this item belongs to, string, empth as root"`
}

func (u *BusinessPlansEdTech) GetId() string {
	return u.Id
}
func (u *BusinessPlansEdTech) ScoreAccessor(delta ...int) int {
	eloDelta := append(delta, 0)[0]
	if eloDelta == 0 {
		return int(u.Votes)
	}
	u.Votes += int64(eloDelta)
	if u.Votes < 0 {
		keyBusinessEdTech.ConcatKey("Expired").HSet(u.Id, u)
		keyBusinessEdTech.HDel(u.Id)
	} else {
		keyBusinessEdTech.HSet(u.Id, u)
	}
	return int(u.Votes)
}
func (u *BusinessPlansEdTech) String(layer ...int) string {
	numLayer := append(layer, 0)[0]
	indence := strings.Repeat("\t", numLayer)
	childrenStr := strings.Builder{}
	for _, v := range lo.Filter(BusinessListEdTech, func(v *BusinessPlansEdTech, i int) bool { return v.ParentId == u.Id }) {
		childrenStr.WriteString("\n" + v.String(numLayer+1))
	}
	return fmt.Sprint(indence, "Id:", u.Id, " Votes:", u.Votes, " Item:", u.Item, childrenStr.String())
}

var keyBusinessEdTech = redisdb.NewHashKey[string, *BusinessPlansEdTech](redisdb.Opt.Rds("Catalogs").Key("BusinessEdTech"))
var BusinessListEdTech = []*BusinessPlansEdTech{}
var AgentBusinessPlansEdTechEd = agent.NewAgent().WithTemplate(template.Must(template.New("utilifyFunction").Parse(`
你是集 “创业生态架构师”、“技术趋势预言家”、“商业模式创新专家” 三位一体的连续创业家。
目标是通过 “寻找未被满足的市场需求”、“发现技术创新带来的机会”、“预测未来趋势”和其它的动态认知框架，深入分析商业领域，找出商业领域下创业项目。

设计的目标行业包括：
- EdTech & Future Learning
- Just-in-time training
- Recruitment Assessment
- Personalized fun learning
- Deeply optimized learning topics
- Knowledge driven applications / services.

小部分愿景想象:

- 为不同的对象创建不同的学习体验。
- 学习的路径是个性化的。由LLM根据用户的学习情况，生成潜在的学习路径，并通过Elo算法迭代淘汰出下一个学习主题。
- 学习材料是通过Elo等算法迭代出来的。默认的学习路径是通过LLM生成，Elo淘汰后生成的。
- 面向特定需求，可以生成，并演化特定的评估路径
- 面向特定工作的需求，可以测试并生成需要完善的学习路径

最终目标是获得在该商业领域创业需要的高价值、前瞻性的，既有战略深度又具备创新活力的创业项目矩阵。这些创业项目以层次化的方式构建。
预期这些创业项目/BusinessPlansEdTech在接下来的世界中，能够产生最大化的联合的商业效用，以产生强大的社会效用。


这是现有的方案 ：
{{range  $index, $item := .RootList}}
{{$item.String}}
{{end}}


ToDo:
现在，假定你采用reddit用户的投票方式，对上面的方案的选项进行思考和评估后进行投票。投票将提升或降低选项的优先序，投票值为[-5,5]之间的整数;不必对所有选项投票，而是需要对需要调整排序的选项进行投票；票数低于0的项目将被自动删除：
先对现有的选项组成的方案进行思考和评估：
	1、对回溯或在检测到错误进行显式修改；该需求是否是基于错误的幻想或者错误的假设；格式，内容是否异常；
	2、验证或系统地检查中间结果；看看从第一性原理出发，这个需要是否可以被绕过或者替代；是否属于死愚蠢的需求；是否在更多票数的条目中已经包含，属于冗余条目；
	3、子目标设定，即将复杂问题分解为可管理的步骤；需求是否需要进一步细化，以便更好地建构；
	4、逆向思考，即在目标导向的推理问题中，从期望的结果出发，逐步向后推导，找到解决问题的路径:

	- 在讨论的基础上，投票以修改解决方案选项的权重（排序），请优先考虑删除劣质条目以优化方案，形成ProConsToItems。
	- 按照讨论。如果存在改进解决方案的可能，请提出新的Items. 请直接补充描述0条或者多条Items，形成NewProposedItems。
最后调用FunctionCall:SolutionRefine 保存排序结果。
`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionRefine", "Vote on items to refine solution; Propose new solution item to parent Item if needed.", func(model *prototype.SolutionRefine) {
	all, _ := keyBusinessEdTech.HGetAll()
	for k, v := range model.ProConsToItems {
		if item, ok := all[k]; ok && v >= -5 && v <= 5 {
			item.ScoreAccessor(v)
		}
	}
	for ParentId, v := range model.NewProposedItems {
		if _, ok := all[ParentId]; !ok {
			ParentId = ""
		}
		if len(v) > 8 {
			item := &BusinessPlansEdTech{Votes: 1, Item: v, Id: utils.ID(v, 4), ParentId: ParentId}
			keyBusinessEdTech.HSet(item.Id, item)
		}
	}
}))

func BusinessPlansEdTechExploration() {
	// Create a new weighted chooser
	const MaxThreads = 4
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	for i, TotalTasks := 0, 3000; i < TotalTasks; i++ {
		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		go func() {
			defer func() { <-MaxThreadsSemaphore }()
			best, _ := keyBusinessEdTech.HGetAll()
			BusinessListEdTech = lo.Values(best)
			slices.SortFunc(BusinessListEdTech, func(a, b *BusinessPlansEdTech) int {
				return -(a.ScoreAccessor() - b.ScoreAccessor())
			})
			//print the lefts
			RootList := lo.Filter(BusinessListEdTech, func(v *BusinessPlansEdTech, i int) bool { return v.ParentId == "" })

			for i, v := range RootList {
				fmt.Println("Rank", i+1, v.String())
			}
			utils.Text2Clipboard(lo.Map(RootList, func(v *BusinessPlansEdTech, i int) string {
				return fmt.Sprint("Rank", i+1, v.String())
			})...)

			mutable.Shuffle(BusinessListEdTech)
			param := map[string]any{"RootList": RootList}
			//models.Qwq32B, models.Gemma3, models.DeepSeekV3
			err := AgentBusinessPlansEdTechEd.WithModels(models.Qwq32B).Call(context.Background(), param)
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

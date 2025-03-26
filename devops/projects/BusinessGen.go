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
)

type BusinessPlans struct {
	Id       string         `description:"Id, string, unique"`
	Votes    int64          `description:"-"`
	Item     string         `description:"item of the solution"`
	ParentId string         `description:"Parent Id this item belongs to, string, empth as root"`
	Extra    map[string]any `msgpack:"-" json:"-" description:"-"`
}

func (u *BusinessPlans) GetId() string {
	return u.Id
}
func (u *BusinessPlans) ScoreAccessor(delta ...int) int {
	eloDelta := append(delta, 0)[0]
	if eloDelta >= -5 && eloDelta <= 5 && eloDelta != 0 {
		u.Votes += int64(eloDelta)
		if key, ok := u.Extra["Key"].(*redisdb.HashKey[string, *BusinessPlans]); ok {
			if u.Votes < 0 {
				key.ConcatKey("Expired").HSet(u.Id, u)
				key.HDel(u.Id)
			} else {
				key.HSet(u.Id, u)
			}
		}
	}
	return int(u.Votes)
}
func (u *BusinessPlans) String(layer ...int) string {
	numLayer := append(layer, 0)[0]
	indence := strings.Repeat("\t", numLayer)
	childrenStr := strings.Builder{}
	itemList, _ := u.Extra["ItemList"].([]*BusinessPlans)
	for _, v := range lo.Filter(itemList, func(v *BusinessPlans, i int) bool { return v.ParentId == u.Id }) {
		childrenStr.WriteString("\n" + v.String(numLayer+1))
	}
	return fmt.Sprint(indence, "Id:", u.Id, " Votes:", u.Votes, " Item:", u.Item, childrenStr.String())
}

var keyBusinessPlans = redisdb.NewHashKey[string, *BusinessPlans](redisdb.WithRds("Catalogs"))

var AgentBusinessPlans = agent.NewAgent(template.Must(template.New("utilifyFunction").Parse(`
你是集 “创业生态架构师”、“技术趋势预言家”、“商业模式创新专家” 三位一体的连续创业家。
目标是通过 “寻找未被满足的市场需求”、“发现技术创新带来的机会”、“预测未来趋势”和其它的动态认知框架，深入分析商业领域 “{{.Topic}}”，找出商业领域(Industry) "{{.Topic}}"下创业项目。
最终目标是获得在该领域创业需要的高价值、前瞻性的，既有战略深度又具备创新活力的创业项目矩阵。这些创业项目以层次化的方式构建。
预期这些创业项目/BusinessPlans在接下来的世界中，在商业领域"{{.Topic}}"能够产生最大化的联合的商业效用，以捕获该行业中的主要机会。


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
	基于BusinessUtilityFunction=exp(WeightMarketSizeln(MarketSize) + 0.18ln(MarketGrowthRate) + 0.22ln(ExpectedReturn) + 0.10ln(TechnicalFeasibility) + 0.15ln(InnovationPotential) + 0.080ln(ResourceAllocation) - 0.12ln(ProjectRisk + 1) - 0.080ln(CompetitionIntensity) - 0.10ln(ImplementationDifficulty) + 0.060ln(TimeToMarket) + 0.040ln(TeamExperience) + 0.050ln(PolicySupport))  方案是否有重大缺陷。


	- 在讨论的基础上，投票以修改解决方案选项的权重（排序），请优先考虑删除劣质条目以优化方案，形成ProConsToItems。
	- 按照讨论。如果存在改进解决方案的可能，请提出新的Items. 请直接补充描述0条或者多条Items，形成NewProposedItems。
最后调用FunctionCall:SolutionRefine 保存排序结果。
`))).WithToolCallLocked().WithTools(tool.NewTool("SolutionRefine", "Vote on items to refine solution; Propose new solution item to parent Item if needed.", func(model *prototype.SolutionRefine) {
	hashKey, ok := model.Extra["Key"].(*redisdb.HashKey[string, *BusinessPlans])
	if !ok || model == nil {
		return
	}
	all, _ := hashKey.HGetAll()
	for k, v := range model.ProConsToItems {
		if item, ok := all[k]; ok {
			item.Extra = model.Extra
			item.ScoreAccessor(v)
		}
	}
	for ParentId, v := range model.NewProposedItems {
		if _, ok := all[ParentId]; !ok || ParentId == "root" {
			ParentId = ""
		}
		if len(v) > 8 {
			item := &BusinessPlans{Votes: 1, Item: v, Id: utils.ID(v), ParentId: ParentId}
			hashKey.HSet(item.Id, item)
		}
	}
}))

func BusinessPlansExploration() {
	// Create a new weighted chooser
	const MaxThreads = 1
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	for i, TotalTasks := 0, 1000*1000; i < TotalTasks; i++ {
		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		go func() {
			//industry := IndustryList[rand.Intn(len(IndustryList))]
			industry := "Synthetic Biology Platforms"
			defer func() { <-MaxThreadsSemaphore }()
			key := keyBusinessPlans.ConcatKey(industry)
			best, _ := key.HGetAll()
			// for k, v := range best {
			// 	var k1 string
			// 	if err := msgpack.Unmarshal([]byte(k), &k1); err == nil {
			// 		key.HDel(k)
			// 		key.HSet(k1, v)
			// 	}
			// }
			listSorted := lo.Values(best)
			slices.SortFunc(listSorted, func(a, b *BusinessPlans) int {
				return -(a.ScoreAccessor() - b.ScoreAccessor())
			})
			RootList := lo.Filter(listSorted, func(v *BusinessPlans, i int) bool {
				return v.ParentId == ""
			})
			param := map[string]any{
				"RootList": RootList,
				"Key":      key,
				"Topic":    industry,
				"ItemList": listSorted,
				//"Model":       models.LoadbalancedPick(models.Qwq32B,models.Gemma3) ,
				//"Model": models.LoadbalancedPick(models.Qwq32B),
				"Model":       models.LoadbalancedPick(models.Qwq32B, models.Gemma3),
				"TotoalNodes": len(best),
			}
			for _, v := range listSorted {
				v.Extra = param
			}
			//print the lefts
			for i, v := range RootList {
				fmt.Println("Rank", i+1, v.String())
			}
			err := AgentBusinessPlans.WithModel(models.Qwq32B).Call(context.Background(), param)
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

var IndustryList = []string{
	"AI Core Technologies",
	"Biotech & Medical Innovations",
	"Cybersecurity Infrastructure & Quantum Security",
	"Hydrogen & Clean Energy Solutions",
	"Quantum Computing & Sensing",
	"Healthcare & Longevity (合并医疗创新、药物发现、个性化医疗等)",
	"Advanced Materials & Nanotechnology",
	"Synthetic Biology & Bioengineering",
	"Advanced Battery Technology & Energy Storage",
	"Space Commerce & Satellite Techtop",
	"Smart Manufacturing & Industry 4.0",
	"Circular Economy & Resource Management (整合回收、材料科学、可持续制造)",
	"Generative AI and Foundation Models",
	"Carbon Capture, Utilization, and Storage (CCUS)",
	"Smart Energy & Grid Management",
	"Smart Infrastructure (城市、交通、水资源、能源网格的智能化整合)",
	"Bioeconomy (合成生物、生物制造、可持续材料等)",
	"Smart Agriculture & Farming Tech",
	"Renewable Energy Infrastructure Development",
	"Sustainable Materials Science",
	"AI-Driven Climate & Environmental Modeling",
	"Sustainable Water Management & Purification",
	"Fusion Power",
	"Carbon Removal Technologies",
	"Climate Change Adaptation Technologies",
	"Decentralized Renewable Energy Grids",
	"Sustainable Agriculture & Regenerative Farming",
	"Cybersecurity for Critical Infrastructure",
	"Sustainable Chemistry & Green Materials",
	"Neurotechnology & Brain-Computer Interfaces",
	"Sustainable Food Systems & Alternative Proteins",
	"Precision Robotics & Automation (for Healthcare, Manufacturing, Logistics)",
	"Space-Based Solar Power (SBSP)",
	"Synthetic Biology Platforms",
	"Digital Twins & Predictive Modeling (across industries)",
	"Energy Storage Solutions (Beyond Batteries)",
}

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
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

// {"Id":"H3URX2jY","BusinessClusters":{"AI-Driven Public Policy Analysis":2500,"AI-Driven Fashion & Wearable Tech":5200,"AI in Public Services & Governance":2500,"Advanced Materials & Nanotechnology":24000,"Virtual & Digital Asset Management":15200,"AI-Driven Customer Experience":6200,"AI-Driven Agricultural & Farm Management":16100,"AI in Waste Management & Circular Economy":45000,"Green Energy & Renewable Tech":124000,"AI-Driven Tourism & Hospitality":5500,"AI-Driven Automotive & Mobility Solutions":6800,"AI in Disaster Management & Safety":2500,"AI-Driven Supply Chain & Inventory Optimization":7500,"AI-Driven Urban Planning & Development":5000,"AI in Climate & Environmental Science":124000,"AI in Energy & Utilities":124000,"AI-Driven Analytics & Big Data":22500,"AI-Driven Real Estate & Architecture":9100,"AI-Driven Fraud Detection & Compliance":7300,"AI Core Technologies":45000,"AI in Agriculture & Food Tech":16100,"Extended Reality (XR) & Virtual Worlds":9800,"Health Informatics & Telemedicine":41000,"AI in Legal & Compliance":11100,"AI in Telecommunications & Networks":9900,"Digital Payment Systems & Fintech":89000,"AI-Driven Cultural & Heritage Preservation":3600,"AI in Media & Entertainment":2000,"Smart Manufacturing & Industry 4.0":34000,"Space Commerce & Satellite Tech":10500,"AI-Driven Risk Management & Insurance":4800,"AI-Driven Legal Document Analysis":7900,"AI-Driven Journalism & Media Analysis":3200,"AI in Manufacturing & Industry 4.0":34000,"AI in Marketing & Advertising":6200,"AI in Blockchain & Digital Assets":94100,"Neural Interfaces & Bioelectronics":9300,"Smart Home & IoT Solutions":5800,"AI in Education & EdTech":21000,"AI-Driven Personalized Learning":21000,"AI-Driven Sports & Fitness Tech":5200,"AI-Driven Public Safety & Defense Systems":2500,"AI in Water & Environmental Tech":6500,"AI-Driven Mining & Resources":9200,"AI-Driven Disaster Response Systems":2500,"AI-Driven Security & Surveillance Systems":2500,"AI in Transportation & Logistics":15000,"Drone & Autonomous Systems":38000,"Smart Agriculture & Farming Tech":16500,"Smart Transportation & Urban Mobility":13300,"AI-Driven Media & Content Creation":3200,"Quantum Computing & Sensing":10000,"Smart Energy & Grid Management":9900,"Synthetic Biology & Bioengineering":6700,"AI-Driven Public Infrastructure Management":8900,"AI in Retail & E-commerce":15000,"AI in Cultural Preservation & Heritage":3600,"AI in Robotics & Automation":42000,"Hydrogen & Clean Energy Solutions":34200,"AI-Driven Telecom & Network Optimization":9900,"AI-Driven Customer Support & Chatbots":5800,"AI in Healthcare & Biotech":45000,"AI in Smart Cities & Urban Tech":27500,"Cybersecurity Infrastructure & Quantum Security":38000,"AI-Driven Virtual Asset Trading":15200,"AI-Driven Climate & Environmental Modeling":7200,"AI in Finance & Banking":22500,"AI in Cybersecurity & Digital Defense":38000,"Biotech & Medical Innovations":54000,"AI-Driven Emergency Response":2500,"AI-Driven Energy Efficiency Solutions":9500}}
type BusinessClusterItem struct {
	Id     string `description:"Id, string, unique"`
	Votes  int64  `description:"-"`
	Market int64  `description:"Market, integer, Market size in million dollars"`
	Item   string `description:"Item,string" msgpack:"ClusterDescription"`
}

func (u *BusinessClusterItem) GetId() string {
	return u.Id
}
func (u *BusinessClusterItem) ScoreAccessor(delta ...int) int {
	if eloDelta := append(delta, 0)[0]; eloDelta != 0 {
		u.Votes += int64(eloDelta)
		if u.Votes < 0 {
			keyBusinessClustering.ConcatKey("Expired").HSet(u.Id, u)
			keyBusinessClustering.HDel(u.Id)
		} else {
			keyBusinessClustering.HSet(u.Id, u)
		}
	}
	return int(u.Votes)
}

var keyBusinessClustering = redisdb.NewHashKey[string, *BusinessClusterItem](redisdb.WithRds("projects"), redisdb.WithKey("BusinessClusteringRdit"))

// 为什么Qwen能自我改进推理，Llama却不行 https://mp.weixin.qq.com/s/OvS61OrDp6rB-R5ELg48Aw
// 并且每次就一个确切的改进方向进行深度分析，反代的深度分析第一性原理之上的需求，深度创新以做出实质的改进。要痛恨泛泛而谈的内容，重复空洞的内容，因为现在是在开发世界级的工具。
var AgentBusinessClustering = agent.NewAgent(template.Must(template.New("utilifyFunction").Parse(`
现在我们要对世界上的商业活动进行分类 (创建商业活动的类目)（注意，不是创建具体的商业项目）。

要关注类目的分布的覆盖性,能够覆盖世界上绝大多数的商业活动，并且按照重要性排序。

我们的最终目标是寻找具有长期市场重要性的商业活动类目。

这是现有的商业活分类方案 ：
{{range  $index, $item := .ItemList}}
"Id":"{{$item.Id}}" "Votes":"{{$item.Votes}}" 
{{$item.Item}}
{{end}}

ToDo:
现在，假定你采用reddit用户的投票方式，对上面的方案的选项进行思考和评估后进行投票。投票将提升或降低选项的优先序，投票值为[-5,5]之间的整数;不必对所有选项投票，而是需要对需要调整排序的选项进行投票；票数低于0的项目将被自动删除：
先对现有的选项组成的方案进行思考和评估：
	1、对回溯或在检测到错误进行显式修改；
	2、验证或系统地检查中间结果；
	3、子目标设定，即将复杂问题分解为可管理的步骤
	4、逆向思考，即在目标导向的推理问题中，从期望的结果出发，逐步向后推导，找到解决问题的路径。

	- 在讨论的基础上，投票以修改解决方案选项的权重（排序），形成ProConsToItems。
	- 按照讨论。如果存在改进解决方案的可能，请提出新的Items. 请直接补充描述0条或者多条Items，形成NewProposedItems。
最后调用FunctionCall:SolutionRefine 保存排序结果。

`))).WithToolCallLocked().WithTools(tool.NewTool("SolutionRefine", "Save sorted Items, Items represented as Id list.", func(model *prototype.SolutionRefinement) {
	if model == nil {
		return
	}
	all, _ := keyBusinessClustering.HGetAll()
	for k, v := range model.ProConsToItems {
		if item, ok := all[k]; ok {
			item.ScoreAccessor(v)
		}
	}
	for _, v := range model.NewProposedItems {
		item := &BusinessClusterItem{Votes: 1, Item: v, Id: utils.ID(v)}
		keyBusinessClustering.HSet(item.Id, item)
	}
}))

func BusinessClusteringExploration() {
	// var businessCluster = map[string]uint{
	// 	"AI-Driven Public Policy Analysis":                2500,
	// 	"AI-Driven Fashion & Wearable Tech":               5200,
	// 	"AI in Public Services & Governance":              2500,
	// 	"Advanced Materials & Nanotechnology":             24000,
	// 	"Virtual & Digital Asset Management":              15200,
	// 	"AI-Driven Customer Experience":                   6200,
	// 	"AI-Driven Agricultural & Farm Management":        16100,
	// 	"AI in Waste Management & Circular Economy":       45000,
	// 	"Green Energy & Renewable Tech":                   124000,
	// 	"AI-Driven Tourism & Hospitality":                 5500,
	// 	"AI-Driven Automotive & Mobility Solutions":       6800,
	// 	"AI in Disaster Management & Safety":              2500,
	// 	"AI-Driven Supply Chain & Inventory Optimization": 7500,
	// 	"AI-Driven Urban Planning & Development":          5000,
	// 	"AI in Climate & Environmental Science":           124000,
	// 	"AI in Energy & Utilities":                        124000,
	// 	"AI-Driven Analytics & Big Data":                  22500,
	// 	"AI-Driven Real Estate & Architecture":            9100,
	// 	"AI-Driven Fraud Detection & Compliance":          7300,
	// 	"AI Core Technologies":                            45000,
	// 	"AI in Agriculture & Food Tech":                   16100,
	// 	"Extended Reality (XR) & Virtual Worlds":          9800,
	// 	"Health Informatics & Telemedicine":               41000,
	// 	"AI in Legal & Compliance":                        11100,
	// 	"AI in Telecommunications & Networks":             9900,
	// 	"Digital Payment Systems & Fintech":               89000,
	// 	"AI-Driven Cultural & Heritage Preservation":      3600,
	// 	"AI in Media & Entertainment":                     2000,
	// 	"Smart Manufacturing & Industry 4.0":              34000,
	// 	"Space Commerce & Satellite Tech":                 10500,
	// 	"AI-Driven Risk Management & Insurance":           4800,
	// 	"AI-Driven Legal Document Analysis":               7900,
	// 	"AI-Driven Journalism & Media Analysis":           3200,
	// 	"AI in Manufacturing & Industry 4.0":              34000,
	// 	"AI in Marketing & Advertising":                   6200,
	// 	"AI in Blockchain & Digital Assets":               94100,
	// 	"Neural Interfaces & Bioelectronics":              9300,
	// 	"Smart Home & IoT Solutions":                      5800,
	// 	"AI in Education & EdTech":                        21000,
	// 	"AI-Driven Personalized Learning":                 21000,
	// 	"AI-Driven Sports & Fitness Tech":                 5200,
	// 	"AI-Driven Public Safety & Defense Systems":       2500,
	// 	"AI in Water & Environmental Tech":                6500,
	// 	"AI-Driven Mining & Resources":                    9200,
	// 	"AI-Driven Disaster Response Systems":             2500,
	// 	"AI-Driven Security & Surveillance Systems":       2500,
	// 	"AI in Transportation & Logistics":                15000,
	// 	"Drone & Autonomous Systems":                      38000,
	// 	"Smart Agriculture & Farming Tech":                16500,
	// 	"Smart Transportation & Urban Mobility":           13300,
	// 	"AI-Driven Media & Content Creation":              3200,
	// 	"Quantum Computing & Sensing":                     10000,
	// 	"Smart Energy & Grid Management":                  9900,
	// 	"Synthetic Biology & Bioengineering":              6700,
	// 	"AI-Driven Public Infrastructure Management":      8900,
	// 	"AI in Retail & E-commerce":                       15000,
	// 	"AI in Cultural Preservation & Heritage":          3600,
	// 	"AI in Robotics & Automation":                     42000,
	// 	"Hydrogen & Clean Energy Solutions":               34200,
	// 	"AI-Driven Telecom & Network Optimization":        9900,
	// 	"AI-Driven Customer Support & Chatbots":           5800,
	// 	"AI in Healthcare & Biotech":                      45000,
	// 	"AI in Smart Cities & Urban Tech":                 27500,
	// 	"Cybersecurity Infrastructure & Quantum Security": 38000,
	// 	"AI-Driven Virtual Asset Trading":                 15200,
	// 	"AI-Driven Climate & Environmental Modeling":      7200,
	// 	"AI in Finance & Banking":                         22500,
	// 	"AI in Cybersecurity & Digital Defense":           38000,
	// 	"Biotech & Medical Innovations":                   54000,
	// 	"AI-Driven Emergency Response":                    2500,
	// 	"AI-Driven Energy Efficiency Solutions":           9500,
	// }
	// for k, v := range businessCluster {
	// 	idStr := fmt.Sprintf("%x", xxhash.Sum64String(k))
	// 	keyBusinessClustering.HSet(idStr, &BusinessClusterItem{Id: idStr, Votes: int64(rand.Intn(5) + 3), Market: int64(v)})
	// }
	const MaxThreads = 24
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	for i, TotalTasks := 0, 1000*1000; i < TotalTasks; i++ {
		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		best, _ := keyBusinessClustering.HGetAll()
		listSorted := lo.Values(best)
		slices.SortFunc(listSorted, func(a, b *BusinessClusterItem) int {
			return -(a.ScoreAccessor() - b.ScoreAccessor())
		})
		//print the lefts
		for i, v := range listSorted {
			fmt.Println("Rank", i+1, v.GetId(), "Score", int(1000*v.ScoreAccessor()), v.Item)
		}

		param := map[string]any{
			"ItemList": listSorted,
			//"Model":       models.LoadbalancedPick(models.Qwq32B,models.Gemma3) ,
			"Model":       models.LoadbalancedPick(models.Gemma3),
			"TotoalNodes": len(best),
		}
		go func(param map[string]any) {
			defer func() { <-MaxThreadsSemaphore }()
			err := AgentBusinessClustering.Call(context.Background(), param)
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

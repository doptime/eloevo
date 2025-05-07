package projects

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"

	// "github.com/yourbasic/graph"

	"github.com/doptime/qmilvus"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
)

type BusinessPlans struct {
	Id   string `description:"Required when update. Id, string, unique." milvus:"PK,in,out"`
	Item string `description:"Required when create. item of the solution. Bullet Name of Module, Constraints, Guidelines, Architecturals, Nexus or Specifications."`

	SuperEdge      bool     `description:"bool,true if the item is super edge of the solution graph. super edge 描述节点之间的协议,约定,约束,标准,规范,想法,技术路线,时间限制,资源限制,法律客户需求,反馈限制、层次化约束等 "`
	SuperEdgeNodes []string `description:"array of Ids. If this node is super edge. here lists the child nodes that belongs to this SuperEdge. SuperEdgeNodes不能包含超边节点，因为超边节点实际上是图的边而不是图的节点，超边包含超边会破坏图结构. \nRequired by SuperEdge item. update each time super edge revised. "`

	Importance int64 `description:"int, value >=- 1 and value<= 10, Importance. \nRequired when create; optional when update. making Importance < 0  to Remove the item."`
	Priority   int64 `description:"int, value >= 0 and value <= 10 . \n Required for module node. use in Gatt chart to determin the priority of the item. the lower the higher the priority."`

	EmbedingVector []float32 `description:"-" milvus:"dim=1024,index" `

	//初始添加的时候得分为0，Elo 后产生Elo分数
	Elo      float64                   `description:"-"`
	AllItems map[string]*BusinessPlans `description:"-" msgpack:"-"` //所有的条目
	//被人类专家标记为锁定的条目。Locked = true. 不能被删除和修改
	Locked bool `description:"-"`
}

func (u *BusinessPlans) ScoreAccessor(delta ...int) int {
	if len(delta) > 0 {
		u.Elo += float64(delta[0])
		keyBusinessDronebot.HSet(u.Id, u)
	}

	return int(u.Elo)
}
func (u *BusinessPlans) GetId() string {
	return u.Id
}

var milvusCollection = qmilvus.NewCollection[*BusinessPlans]("milvus.lan:19530").CreateCollection()

func (u *BusinessPlans) Embed(embed ...[]float32) []float32 {
	if len(embed) > 0 {
		u.EmbedingVector = embed[0]
		keyBusinessDronebot.HSet(u.Id, u)
	}
	return u.EmbedingVector
}

func (u *BusinessPlans) String(layer ...int) string {
	numLayer := append(layer, 0)[0]
	indence := strings.Repeat("\t", numLayer)
	var elo = ""
	if u.Elo > 0 && !u.Locked {
		elo = fmt.Sprintf(" Elo:%.2f", u.Elo)
	}
	return fmt.Sprint(indence, u.Item, " [Id:", u.Id, lo.Ternary(u.SuperEdge, " SuperEdge", ""), " importance:", strconv.Itoa(int(u.Importance)), " priority:", strconv.Itoa(int(u.Priority)), elo, "]\n")
	//return fmt.Sprint(indence, "Id:", u.Id, " Importance:", u.Importance, communityCore, " Priority:", u.Priority, "\n", u.Item, "\n\n")
}

// var keyBusinessDronebotbak = redisdb.NewHashKey[string, *BusinessPlans](redisdb.Opt.Rds("Catalogs").Key("BusinessDronebot250412"))
var keyBusinessDronebot = redisdb.NewHashKey[string, *BusinessPlans](redisdb.Opt.Rds("Catalogs").Key("BusinessDronebotExploration"))
var ForbiddenWords = []string{"区块链", "量子", "氢燃料", "纠缠", "quantum", "blockchain", "hydrogen", "entanglement", "co2", "carbon sequestration"}

type SuperEdgePlannedForNextLoop struct {
	TopicToDiscuss string   `description:"Topic to discuss, string"`
	SuperEdgeIds   []string `description:"Super edge Ids, []string"`
}

var keyIterPlannedDrone = redisdb.NewListKey[*SuperEdgePlannedForNextLoop](redisdb.Opt.Rds("Catalogs"))

// ## 已明确的全局技术路线:
// - 采用上市的高放电的电池. 不考虑无线充电。
// - 超低成本模块化无人机。机身(包裹)可拆卸，可以动态装配到固定翼和多旋翼无人机上
// - 分布式的高塔式无人机中继站。提供机身维护和电池充电，包裹路由服务。
// - 低速、低空、超高滑翔比的飞机。
// - 不采用操作系统。不采用软件。只采用硬件运行大型神经网络模型，只有Wifi,GPS，摄像头和姿控等最少的输入信号。
// - 仅考虑 PayLoad < 100kg 的无人机，不考虑载人或大型机。

var AgentBusinessPlans = agent.NewAgent(template.Must(template.New("AgentBusinessPlansDrone").Parse(`
系统愿景:实现AI时代，无人机作为基础物流平台。确保方案和应用的极致的简单、可靠、低成本、高效用。
你的核心目标是基于第一性原理的工程学实现，构建一个在无人机平台和机器人应用领域具有高价值、高可行性的项目模块矩阵，这些项目应能在未来的世界中产生最大的联合商业效用和社会效用。

## 涉及的目标行业包括：
- AI-Driven Bobotic Development
- Robotic As a Service
- Drone Technology & Solution
- Suppy Chain & Drone & Logistics Technology
    -如从农业产地直供最终零售点
- Sustainable Packaging Technology
- Sustainable Transportation Infrastructure

## 部分愿景:
- 借助外接电源或超高速放电电池，垂直起飞的固定翼无人机
- 它是一个非常便利的载具平台。可以提供各种机器人的投送和收回服务
- 由于极高的滑翔比。它的物流成本只有汽车的1/10和船运的1/2. 可以在全球内完成有中继的长途运输
- 它可以借助地形和动态风向变化，实现能量的节约。
- 联合多机器人和多飞机。送外卖，入户医疗检查。在户外部署就餐，住宿，岗哨，它能做很多。
- 在未来的世界中，基于无人机平台的全球即时物流和资源分配系统，是最重要的基础设施。
- AI 驱动的自主机器人团队，能够高效协作完成复杂任务。

## 本系统采用迭代方式来渐进得完成愿景，当前的迭代会持续数十万次，直至最终目标实现。每一轮的迭代中，请通过一系列的Funtioncall 调用，来完善或改进方案的实现。
	整个解决方案被建模为顶点和边的图。其中的边为超边(超边是连接多个顶点的边)可以连接两个或两个以上的模块节点。 现有的方案由两类节点构成:
	1) 超边节点
		超边节点就是图的边。但由于可以连接多个顶点，所以是超边。作为一种图结构处理技巧，我们把超边看做是非普通节点也叫超边节点。
		超边节点用以处理解决方案的必要约束条件，包括反馈、技术路径、规范、意图，资源限制、层次化聚类等用来影响模块节点生成和调整的节点。超边节点被构建用以驱动系统的变化。
		超边应显式设置SuperEdge=true。
	2) 模块节点
		模块节点是直接给出解决方案的节点
		模块只能通过超边和其它的模块节点完成耦合。也就是通过实现超边节点约束来自身定义
- 解决方案被拆分成为多个超边相关的Community。每次迭代处理若干(2-5)Community。

{{if eq .task "gen"}}
首先，请深度分析,并提出解决方案的改进方案。以使得最终的愿景能够以非常可行的方式落地。
通过多次调用 FunctionCall:SolutionItemRefine 来保存方案改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Id,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)通过指定Id,修改Importance = -1 来删除无效节点:

## 这是现有的所有SuperEdges：
{{.SuperEdges}}

## 这是当前解决方案(SuperEdgeCommunities):
{{.Communities}}

TopicToDiscuss:
{{.TopicToDiscuss}} 

最后请为下一轮的迭代进行规划。讨论下一轮需要重点完善的关注方向，围绕这个方向选择现有的超边集合，并且调用 FunctionCall:SuperEdgePlannedForNextLoop 进行保存。

{{else if eq .task "batchElo"}}

## 这是当前解决方案(SuperEdgeCommunities):
{{.CommunitiesAll}}

这是新增的超边和节点的集合:
{{.NewNodes}}

现在要使得**现有解决方案**以 **最优的方式** 向 **期望的终态目标系统** 演化。
为此现采用dijkstra算法来实现这个目标。为此，1)我们采用Batch Elo算法来挑选出最优的节点,作为动作（添加、修改或删除）; 2)以一次一个动作的方式，提交最优的动作（设置Locked）来逼近期望的终态目标系统。
请讨论现有的新增节点,以便将它们从好到坏，完成排序。并使用BatchEloResults保存排序结果。
{{end}}


`))).WithToolCallMutextRun()
var businessPlans map[string]*BusinessPlans

func EdgeCommunitiesWithExploration(allNodes map[string]*BusinessPlans, centerEdge *BusinessPlans, c []*BusinessPlans, RandomExplorationNum int) (ret []*BusinessPlans) {
	centerEdge = allNodes[centerEdge.Id]
	//from milvus to redis data
	c = lo.Map(c, func(v *BusinessPlans, _ int) *BusinessPlans {
		return allNodes[v.Id]
	})
	//shouble be valid
	c = lo.Filter(c, func(v *BusinessPlans, _ int) bool {
		return v != nil && v.Importance > 0 && !v.SuperEdge
	})

	//上一轮的EdgeNodes
	SuperEdgeNodes := lo.Filter(centerEdge.SuperEdgeNodes, func(v string, _ int) bool { return allNodes[v] != nil })
	ret = append([]*BusinessPlans{centerEdge}, lo.Map(SuperEdgeNodes, func(id string, i int) *BusinessPlans { return allNodes[id] })...)
	left1, _ := lo.Difference(c, ret)
	return append(ret, left1[:min(RandomExplorationNum, len(left1))]...)

}
func NodeListToString(nodes []*BusinessPlans, uniq, LockedOnly bool) (list string) {
	nodes = lo.Ternary(uniq, lo.Uniq(nodes), nodes)
	if LockedOnly {
		nodes = lo.Filter(nodes, func(v *BusinessPlans, _ int) bool { return v.Locked })
	}
	return strings.Join(lo.Map(nodes, func(v *BusinessPlans, _ int) string { return v.String() }), "\n")
}
func LoadResultsToRedis() {
	//read all files in /Users/yang/Desktop/projects/doptime/evolab/main that starts with asw* and ends with .md
	files, err := filepath.Glob("/Users/yang/Desktop/projects/doptime/evolab/main/asw*.md")
	if err != nil {
		fmt.Println("Error reading files:", err)
		return
	}
	contents := lo.Map(files, func(v string, _ int) string {
		content, _ := os.ReadFile(v)
		return string(content)
	})

	AgentBusinessPlans.WithModels(models.GeminiTB).CallWithResponseString(strings.Join(contents, "\n\n"))

}

func BusinessPlansDronebotExploration() {
	// var Items = []string{"飞机采用上市的先进电池. 不考虑无线充电。燃油引擎。", "超低成本模块化无人机。", "机身(包裹)可拆卸，可以动态装配到固定翼和多旋翼无人机上。",
	// 	"分布式的无人机中继站。提供机身维护和电池换电，包裹路由服务。", "低速、低空、超高滑翔比的飞机。", "仅采用手机主板完成整个控制系统。", "采用迷你神经网络作为核心控制系统。",
	// 	"仅考虑 PayLoad < 25kg 的无人机，不考虑载人或大型机。"}
	// for _, item := range Items {
	// 	id := utils.ID(item, 3)
	// 	_item := &BusinessPlans{Id: id, Item: item, SuperEdge: true, Importance: 9, Locked: true}
	// 	keyBusinessDronebot.HSet(_item.Id, _item)
	// }
	// Create a new weighted chooser
	const MaxThreads = 1
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)
	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		businessPlans, _ = keyBusinessDronebot.HGetAll()
		AgentBusinessPlans.ShareMemoryUpdate("AllItems", businessPlans)
		//commit remove where remove >= 5
		for _, key := range lo.Keys(businessPlans) {
			item := businessPlans[key]

			if utils.HasForbiddenWords(strings.ToLower(item.Item), ForbiddenWords) || item.Importance < 0 {
				keyBusinessDronebot.ConcatKey("Archive").HSet(item.Id, item)
				keyBusinessDronebot.HDel(item.Id)
				delete(businessPlans, item.Id)
				if err := milvusCollection.Remove(item); err != nil {
					fmt.Println("Error removing item from Milvus:", err)
				}
			}
			// 非超边不能有超边节点
			if !item.SuperEdge && len(item.SuperEdgeNodes) > 0 {
				item.SuperEdgeNodes = []string{}
				keyBusinessDronebot.HSet(item.Id, item)
			}
			// 超边节点仅连接到普通节点
			if item.SuperEdge && len(item.SuperEdgeNodes) > 0 {
				SuperEdgeNodesFiltered := lo.Filter(item.SuperEdgeNodes, func(v string, _ int) bool {
					if node, ok := businessPlans[v]; ok {
						return !node.SuperEdge
					}
					return false
				})
				if len(SuperEdgeNodesFiltered) != len(item.SuperEdgeNodes) {
					item.SuperEdgeNodes = SuperEdgeNodesFiltered
					keyBusinessDronebot.HSet(item.Id, item)
				}
			}

		}
		//Upsert to milvus
		var milvusInserts []*BusinessPlans
		for _, item := range lo.Values(businessPlans) {
			if item.Item != "" && len(item.Embed()) != 1024 {
				embed, err := utils.GetEmbedding(item.Item)
				if err == nil {
					item.Embed(embed)
					milvusInserts = append(milvusInserts, item)
				}
			}
		}
		if len(milvusInserts) > 0 {
			milvusCollection.Upsert(lo.Uniq(milvusInserts)...)
		}
		constraints := lo.Filter(lo.Values(businessPlans), func(v *BusinessPlans, _ int) bool {
			return v.SuperEdge
		})

		//save solution to file to visualize
		CommunitiesAllSearched, _, _ := milvusCollection.SearchVectors(lo.Map(constraints, func(v *BusinessPlans, _ int) []float32 { return v.Embed() }), qmilvus.SearchParamsDefault)
		CommunitiesAll := lo.Map(CommunitiesAllSearched, func(community []*BusinessPlans, i int) []*BusinessPlans {
			return EdgeCommunitiesWithExploration(businessPlans, constraints[i], community, 0)
		})
		if i%10 == 0 {
			//with left nodes as a community
			left, _ := lo.Difference(lo.Values(businessPlans), lo.Flatten(CommunitiesAll))
			CommunitiesAll = append(CommunitiesAll, left)

			outputfile, _ := os.Create("BusinessPlans.md")
			childrenStr := strings.Builder{}
			childrenStr.WriteString(fmt.Sprint("条目数量为：", len(businessPlans), "\n"))
			itemReapeated := map[string]int{}
			for i, community := range CommunitiesAll {
				childrenStr.WriteString(fmt.Sprint("\n#### Community", i+1, " size:", len(community), "\n"))
				for _, item := range community {
					if v, ok := itemReapeated[item.Id]; ok {
						childrenStr.WriteString("*R" + strconv.Itoa(v) + "* ")
					}
					childrenStr.WriteString(item.String())
					itemReapeated[item.Id]++
				}
			}
			_, _ = io.WriteString(outputfile, childrenStr.String())
			outputfile.Close()
		}

		// 选择一个用来显示的社区
		var SuperEdgeSelected []*BusinessPlans
		var IterPlan *SuperEdgePlannedForNextLoop
		llen, _ := keyIterPlannedDrone.LLen()
		if llen < 10 && llen > 0 {
			IterPlan, _ = keyIterPlannedDrone.LIndex(rand.Int64N(llen))
		} else if llen >= 10 {
			IterPlan, _ = keyIterPlannedDrone.LPop()
		}
		if IterPlan != nil {
			for _, v := range IterPlan.SuperEdgeIds {
				if p, ok := businessPlans[v]; ok {
					SuperEdgeSelected = append(SuperEdgeSelected, p)
				}
			}
		}
		TopicToDiscuss := "本次迭代的主题是: 通过多次调用 FunctionCall:SolutionItemRefine 来保存对解决方案的改进。"
		if IterPlan != nil && len(IterPlan.TopicToDiscuss) > 0 {
			TopicToDiscuss = IterPlan.TopicToDiscuss
		}
		// 自主规划没有足够的覆盖度。需要随机选择来保证覆盖度
		if len(SuperEdgeSelected) == 0 || rand.IntN(100) < 50 {
			edges := append([]*BusinessPlans{}, constraints...)
			mutable.Shuffle(edges)
			SuperEdgeSelected = edges[:min(3, len(edges))]
		}

		Communities, _, _ := milvusCollection.SearchVectors(lo.Map(SuperEdgeSelected, func(v *BusinessPlans, _ int) []float32 { return v.Embed() }), qmilvus.SearchParamsDefault.WithTopK(50))
		CommunityNodes := lo.Map(Communities, func(community []*BusinessPlans, i int) []*BusinessPlans {
			return EdgeCommunitiesWithExploration(businessPlans, SuperEdgeSelected[i], community, 10)
		})

		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
		newNodes := lo.Filter(lo.Values(businessPlans), func(v *BusinessPlans, _ int) bool { return !v.Locked })
		newNodesSorted := slices.Clone(newNodes)
		slices.SortFunc(newNodesSorted, func(i, j *BusinessPlans) int {
			return -int(i.Elo - j.Elo)
		})
		fmt.Printf("NewNodes Best: %v\n", newNodesSorted[len(newNodesSorted)-1].String())

		go func(Communities, SuperEdges, TopicToDiscuss, CommunitiesAll string, AllItems map[string]*BusinessPlans, newNodes []*BusinessPlans) {
			defer func() { <-MaxThreadsSemaphore }()
			//models.Qwq32B, models.Gemma3, models.DeepSeekV3, models.DeepSeekV3TB,models.Qwen32B,models.GeminiTB,models.Qwen30BA3,models.GLM32B
			// err := AgentBusinessPlans.WithTools(ToolDroneBotIterPlan, ToolDroneBotSolutionItemRefine).WithModels(models.Qwen32B).CopyPromptOnly().Call(context.Background(), map[string]any{
			// 	"Communities":    Communities,
			// 	"SuperEdges":     SuperEdges,
			// 	"TopicToDiscuss": TopicToDiscuss,
			// 	"task":           "batchElo",
			// })
			err := AgentBusinessPlans.WithTools(ToolDroneBatchEloResults).WithModels(models.EloModels.SequentialPick(models.Qwen3B14)).CopyPromptOnly().Call(context.Background(), map[string]any{
				"Communities":    Communities,
				"CommunitiesAll": CommunitiesAll,
				"SuperEdges":     SuperEdges,
				"TopicToDiscuss": TopicToDiscuss,
				"task":           "batchElo",
				"NewNodes":       NodeListToString(newNodes, true, false),
			})
			if err != nil {
				fmt.Printf("Agent call failed: %v\n", err)
			}
		}(NodeListToString(lo.Flatten(CommunityNodes), true, false), NodeListToString(constraints, true, true), TopicToDiscuss, NodeListToString(lo.Flatten(CommunitiesAll), true, true), businessPlans, newNodesSorted)
	}
	// Wait for all the goroutines to finish)
	for i := 0; i < MaxThreads; i++ {
		MaxThreadsSemaphore <- struct{}{}
	}

}

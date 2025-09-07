package projects

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"math/rand/v2"
// 	"os"
// 	"path/filepath"
// 	"slices"
// 	"strconv"
// 	"strings"
// 	"text/template"

// 	// "github.com/yourbasic/graph"

// 	"github.com/doptime/qmilvus"

// 	"github.com/doptime/eloevo/agent"
// 	"github.com/doptime/eloevo/models"
// 	"github.com/doptime/eloevo/utils"
// 	"github.com/doptime/redisdb"
// 	"github.com/samber/lo"
// 	"github.com/samber/lo/mutable"
// )

// // var keyBusinessDronebotbak = redisdb.NewHashKey[string, *BusinessPlans](redisdb.Opt.Rds("Catalogs").Key("BusinessDronebot250412"))
// var KeyBusinessDronebot = redisdb.NewHashKey[string, *SolutionGraphNode](redisdb.Opt.Rds("Catalogs").Key("BusinessDronebotExploration"))
// var ForbiddenWords = []string{"区块链", "量子", "氢燃料", "纠缠", "quantum", "blockchain", "hydrogen", "entanglement", "co2", "carbon sequestration"}

// type SuperEdgePlannedForNextLoop struct {
// 	TopicToDiscuss string   `description:"Topic to discuss, string"`
// 	SuperEdgeIds   []string `description:"Super edge Ids, []string"`
// }

// var keyIterPlannedDrone = redisdb.NewListKey[*SuperEdgePlannedForNextLoop](redisdb.Opt.Rds("Catalogs"))

// // ## 已明确的全局技术路线:
// // - 采用上市的高放电的电池. 不考虑无线充电。
// // - 超低成本模块化无人机。机身(包裹)可拆卸，可以动态装配到固定翼和多旋翼无人机上
// // - 分布式的高塔式无人机中继站。提供机身维护和电池充电，包裹路由服务。
// // - 低速、低空、超高滑翔比的飞机。
// // - 不采用操作系统。不采用软件。只采用硬件运行大型神经网络模型，只有Wifi,GPS，摄像头和姿控等最少的输入信号。
// // - 仅考虑 PayLoad < 100kg 的无人机，不考虑载人或大型机。

// var AgentBusinessPlans = agent.NewAgent(template.Must(template.New("AgentBusinessPlansDrone").Parse(`
// 系统愿景:实现AI时代，无人机作为基础物流平台。确保方案和应用的极致的简单、可靠、低成本、高效用。
// 你的核心目标是基于第一性原理的工程学实现，构建一个在无人机平台和机器人应用领域具有高价值、高可行性的项目模块矩阵，这些项目应能在未来的世界中产生最大的联合商业效用和社会效用。

// ## 涉及的目标行业包括：
// - AI-Driven Bobotic Development
// - Robotic As a Service
// - Drone Technology & Solution
// - Suppy Chain & Drone & Logistics Technology
//     -如从农业产地直供最终零售点
// - Sustainable Packaging Technology
// - Sustainable Transportation Infrastructure

// ## 部分愿景:
// - 借助外接电源或超高速放电电池，垂直起飞的固定翼无人机
// - 它是一个非常便利的载具平台。可以提供各种机器人的投送和收回服务
// - 由于极高的滑翔比。它的物流成本只有汽车的1/10和船运的1/2. 可以在全球内完成有中继的长途运输
// - 它可以借助地形和动态风向变化，实现能量的节约。
// - 联合多机器人和多飞机。送外卖，入户医疗检查。在户外部署就餐，住宿，岗哨，它能做很多。
// - 在未来的世界中，基于无人机平台的全球即时物流和资源分配系统，是最重要的基础设施。
// - AI 驱动的自主机器人团队，能够高效协作完成复杂任务。

// ## 本系统采用迭代方式来渐进得完成愿景，当前的迭代会持续数十万次，直至最终目标实现。每一轮的迭代中，请通过一系列的Funtioncall 调用，来完善或改进方案的实现。
// 	整个解决方案被建模为顶点和边的图。其中的边为超边(超边是连接多个顶点的边)可以连接两个或两个以上的模块节点。 现有的方案由两类节点构成:
// 	1) 超边节点
// 		超边节点就是图的边。但由于可以连接多个顶点，所以是超边。作为一种图结构处理技巧，我们把超边看做是非普通节点也叫超边节点。
// 		超边节点用于实现系统的架构设计，并且通过模块节点用来进一步驱动系统的实现。通过以下维度，来实现系统的架构设计:
// 		 - 解决业务契合度	架构是否真正解决业务痛点？用业务 KPI / 用户场景 / 收益模型倒推技术方案，持续校验“为什么做”。
// 		 - 技术可行性	方案能否落地、运维、扩展？	技术验证（PoC）、性能基准测试、与现有技术栈/团队能力匹配度。
// 		 - 成本–收益比	投入与产出是否平衡？	固定成本（硬件/许可）、可变成本（云资源）、人力/维护，结合收益或风险降低进行 ROI 评估。
// 		 - 风险管理	有哪些风险？怎样缓解？	技术 / 合规 / 安全 / 供应链风险识别 → 减缓措施 → 残余风险可接受性。
// 		 - 治理与可持续性	架构能否迭代、治理？	模块化、接口契约、Observability、版本策略、技术债务控制、文档化。
// 		 - 交付节奏	如何在有限时间内持续交付价值？	与敏捷/DevOps结合的迭代式架构；“Just-Enough Architecture” 概念。
// 		 - 沟通协作	是否与干系方充分参与并达成共识？	理解并响应其它人类用户的需求/专家的反馈/其它AI的评审/Scrum Backlog。
// 		超边应显式设置SuperEdge=true。
// 	2) 解决方案节点/ 模块节点
// 		模块节点是直接给出解决方案的节点
// 		模块节点应显式设置SuperEdge=false。
// 		模块只能通过超边和其它的模块节点完成耦合。也就是模块的耦合应该被显式提出为超边节点。解决方案节点通过实现超边节点约束来完成定义。

// {{if eq .task "super_edge_evolution"}}
// 请通过下面的流程，来完成对系统的增量建构/修改：
// 下面流程以“迭代递增”方式组织，可配合 Scrum Sprint 或阶段性里程碑使用。步骤之间可回溯和并行，只要确保交付物最终一致。
// 业务与目标澄清
// • 干系人访谈、业务流程图、痛点 / 成长目标 / 约束收集
// • 产出：业务目标清单、优先级、可量化 KPI
// 当前状态基线（As-Is）
// • 现有系统拓扑、依赖、痛点、成本
// • 产出：现状架构图、问题清单
// 需求梳理（功能 & 非功能）
// • User Stories、用例、Domain Event
// • NFR：性能、伸缩性、可用性、合规、安全、可观察性
// • 产出：需求规格说明（FRD+NFRD）
// 关键场景与容量预估
// • 选 3–5 个最关键的业务/技术场景做容量 & SLA 预估
// • 产出：容量模型、流量曲线、SLA & SLO
// 架构原则与决策框架
// • 定义指导原则（如：云优先、事件驱动、松耦合、开放标准 …）
// • 确认决策流程（ADR、原则打分、专家评审）
// 高阶方案设计（To-Be View）
// • 架构图（C4 Model、分层、组件、数据流、部署视图）
// • Technology Radar：候选技术、优劣、约束
// • 产出：多视图架构草稿 & 备选方案
// 深度验证（PoC / Spike）
// • 对关键技术/性能/安全点做快速 PoC
// • 产出：PoC 报告、基准数据、Go/No Go 决策
// 详细设计 & Trade-off
// • 接口契约、数据模型、API、时序图
// • 容错、幂等、Observability、CI/CD 流水线
// • 产出：详细设计文档（DDD、数据库 ERD、API 定义 …）
// 风险评估 & 合规审计
// • 威胁建模、隐私评估、软件许可、成本灵敏度分析
// • 产出：风险登记册、缓解计划
// 评审 & 共识
// • 架构评审会（内部 + 外部专家）
// • 记录反馈、确认版本 & 里程碑
// 持续治理 & 迭代
// • ADR/Changelog 持续更新、Observability 指标监控
// • 技术债务看板、能力培训、定期架构回顾
// 在完成对系统的增量建构/修改 后，请使用FunctionallTool 来保存SuprerEdge

// {{else if eq .task "gen"}}
// - 解决方案被拆分成为多个超边相关的Community。每次迭代处理若干(2-5)Community。

// 首先，请深度分析,并提出解决方案的改进方案。以使得最终的愿景能够以非常可行的方式落地。
// 通过多次调用 FunctionCall:SolutionItemRefine 来保存方案改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Id,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)通过指定Id,修改Importance = -1 来删除无效节点:

// ## 这是现有的所有SuperEdges：
// {{.SuperEdges}}

// ## 这是当前解决方案(SuperEdgeCommunities):
// {{.Communities}}

// TopicToDiscuss:
// {{.TopicToDiscuss}}

// 最后请为下一轮的迭代进行规划。讨论下一轮需要重点完善的关注方向，围绕这个方向选择现有的超边集合，并且调用 FunctionCall:SuperEdgePlannedForNextLoop 进行保存。

// {{else if eq .task "batchElo"}}

// ## 这是当前解决方案(SuperEdgeCommunities):
// {{.CommunitiesAll}}

// 这是新增的超边和节点的集合:
// {{.NewNodes}}

// 现在要使得**现有解决方案**以 **最优的方式** 向 **期望的终态目标系统** 演化。
// 为此现采用dijkstra算法来实现这个目标。为此，1)我们采用Batch Elo算法来挑选出最优的节点,作为动作（添加、修改或删除）; 2)以一次一个动作的方式，提交最优的动作（设置Locked）来逼近期望的终态目标系统。
// 请讨论现有的新增节点,以便将它们从好到坏，完成排序。并使用BatchEloResults保存排序结果。
// {{end}}

// `))).WithToolCallMutextRun()
// var businessPlans map[string]*SolutionGraphNode

// func EdgeCommunitiesWithExploration(allNodes map[string]*SolutionGraphNode, centerEdge *SolutionGraphNode, c []*SolutionGraphNode, RandomExplorationNum int) (ret []*SolutionGraphNode) {
// 	centerEdge = allNodes[centerEdge.Id]
// 	//from milvus to redis data
// 	c = lo.Map(c, func(v *SolutionGraphNode, _ int) *SolutionGraphNode {
// 		return allNodes[v.Id]
// 	})
// 	//shouble be valid
// 	c = lo.Filter(c, func(v *SolutionGraphNode, _ int) bool {
// 		return v != nil && v.Importance > 0 && !v.SuperEdge
// 	})

// 	//上一轮的EdgeNodes
// 	SuperEdgeNodes := lo.Filter(centerEdge.SuperEdgeNodes, func(v string, _ int) bool { return allNodes[v] != nil })
// 	ret = append([]*SolutionGraphNode{centerEdge}, lo.Map(SuperEdgeNodes, func(id string, i int) *SolutionGraphNode { return allNodes[id] })...)
// 	left1, _ := lo.Difference(c, ret)
// 	return append(ret, left1[:min(RandomExplorationNum, len(left1))]...)

// }

// func LoadResultsToRedis() {
// 	//read all files in /Users/yang/Desktop/projects/doptime/evolab/main that starts with asw* and ends with .md
// 	files, err := filepath.Glob("/Users/yang/Desktop/projects/doptime/evolab/main/asw*.md")
// 	if err != nil {
// 		fmt.Println("Error reading files:", err)
// 		return
// 	}
// 	contents := lo.Map(files, func(v string, _ int) string {
// 		content, _ := os.ReadFile(v)
// 		return string(content)
// 	})

// 	AgentBusinessPlans.WithModels(models.GeminiTB).CallWithResponseString(strings.Join(contents, "\n\n"))

// }

// func BusinessPlansDronebotExploration() {
// 	// var Items = []string{"飞机采用上市的先进电池. 不考虑无线充电。燃油引擎。", "超低成本模块化无人机。", "机身(包裹)可拆卸，可以动态装配到固定翼和多旋翼无人机上。",
// 	// 	"分布式的无人机中继站。提供机身维护和电池换电，包裹路由服务。", "低速、低空、超高滑翔比的飞机。", "仅采用手机主板完成整个控制系统。", "采用迷你神经网络作为核心控制系统。",
// 	// 	"仅考虑 PayLoad < 25kg 的无人机，不考虑载人或大型机。"}
// 	// for _, item := range Items {
// 	// 	id := utils.ID(item, 3)
// 	// 	_item := &BusinessPlans{Id: id, Item: item, SuperEdge: true, Importance: 9, Locked: true}
// 	// 	keyBusinessDronebot.HSet(_item.Id, _item)
// 	// }
// 	// Create a new weighted chooser
// 	const MaxThreads = 1
// 	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)
// 	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
// 		businessPlans, _ = KeyBusinessDronebot.HGetAll()
// 		for k, v := range businessPlans {
// 			KeyBusinessDronebot.HSet(k, v)
// 		}
// 		AgentBusinessPlans.ShareMemoryUpdate("AllItems", businessPlans)
// 		//commit remove where remove >= 5
// 		for _, key := range lo.Keys(businessPlans) {
// 			item := businessPlans[key]

// 			if utils.HasForbiddenWords(strings.ToLower(item.BulletName), ForbiddenWords) || item.Importance < 0 {
// 				KeyBusinessDronebot.ConcatKey("Archive").HSet(item.Id, item)
// 				KeyBusinessDronebot.HDel(item.Id)
// 				delete(businessPlans, item.Id)
// 				// if err := milvusCollection.Remove(item); err != nil {
// 				// 	fmt.Println("Error removing item from Milvus:", err)
// 				// }
// 			}
// 			// 非超边不能有超边节点
// 			if !item.SuperEdge && len(item.SuperEdgeNodes) > 0 {
// 				item.SuperEdgeNodes = []string{}
// 				KeyBusinessDronebot.HSet(item.Id, item)
// 			}
// 			// 超边节点仅连接到普通节点
// 			if item.SuperEdge && len(item.SuperEdgeNodes) > 0 {
// 				SuperEdgeNodesFiltered := lo.Filter(item.SuperEdgeNodes, func(v string, _ int) bool {
// 					if node, ok := businessPlans[v]; ok {
// 						return !node.SuperEdge
// 					}
// 					return false
// 				})
// 				if len(SuperEdgeNodesFiltered) != len(item.SuperEdgeNodes) {
// 					item.SuperEdgeNodes = SuperEdgeNodesFiltered
// 					KeyBusinessDronebot.HSet(item.Id, item)
// 				}
// 			}

// 		}
// 		//Upsert to milvus
// 		// var milvusCollection = qmilvus.NewCollection[*SolutionGraphNode]("milvus.lan:19530").CreateCollection()
// 		// var milvusInserts []*SolutionGraphNode
// 		// for _, item := range lo.Values(businessPlans) {
// 		// 	if item.BulletDescription!= "" && len(item.Embed()) != 1024 {
// 		// 		embed, err := utils.GetEmbedding(item.BulletName)
// 		// 		if err == nil {
// 		// 			item.Embed(embed)
// 		// 			milvusInserts = append(milvusInserts, item)
// 		// 		}
// 		// 	}
// 		// }
// 		// if len(milvusInserts) > 0 {
// 		// 	milvusCollection.Upsert(lo.Uniq(milvusInserts)...)
// 		// }
// 		var constraints SolutionGraphNodeList = SolutionGraphNodeList(lo.Values(businessPlans)).SuerEdge().PathnameSorted()

// 		//save solution to file to visualize
// 		CommunitiesAllSearched, _, _ := milvusCollection.SearchVectors(lo.Map(constraints, func(v *SolutionGraphNode, _ int) []float32 { return v.Embed() }), qmilvus.SearchParamsDefault)
// 		CommunitiesAll := lo.Map(CommunitiesAllSearched, func(community []*SolutionGraphNode, i int) []*SolutionGraphNode {
// 			return EdgeCommunitiesWithExploration(businessPlans, constraints[i], community, 0)
// 		})
// 		if i%10 == 0 {
// 			//with left nodes as a community
// 			left, _ := lo.Difference(lo.Values(businessPlans), lo.Flatten(CommunitiesAll))
// 			CommunitiesAll = append(CommunitiesAll, left)

// 			outputfile, _ := os.Create("BusinessPlans.md")
// 			childrenStr := strings.Builder{}
// 			childrenStr.WriteString(fmt.Sprint("条目数量为：", len(businessPlans), "\n"))
// 			itemReapeated := map[string]int{}
// 			for i, community := range CommunitiesAll {
// 				childrenStr.WriteString(fmt.Sprint("\n#### Community", i+1, " size:", len(community), "\n"))
// 				for _, item := range community {
// 					if v, ok := itemReapeated[item.Id]; ok {
// 						childrenStr.WriteString("*R" + strconv.Itoa(v) + "* ")
// 					}
// 					childrenStr.WriteString(item.String())
// 					itemReapeated[item.Id]++
// 				}
// 			}
// 			_, _ = io.WriteString(outputfile, childrenStr.String())
// 			outputfile.Close()
// 		}

// 		// 选择一个用来显示的社区
// 		var SuperEdgeSelected []*SolutionGraphNode
// 		var IterPlan *SuperEdgePlannedForNextLoop
// 		llen, _ := keyIterPlannedDrone.LLen()
// 		if llen < 10 && llen > 0 {
// 			IterPlan, _ = keyIterPlannedDrone.LIndex(rand.Int64N(llen))
// 		} else if llen >= 10 {
// 			IterPlan, _ = keyIterPlannedDrone.LPop()
// 		}
// 		if IterPlan != nil {
// 			for _, v := range IterPlan.SuperEdgeIds {
// 				if p, ok := businessPlans[v]; ok {
// 					SuperEdgeSelected = append(SuperEdgeSelected, p)
// 				}
// 			}
// 		}
// 		TopicToDiscuss := "本次迭代的主题是: 通过多次调用 FunctionCall:SolutionItemRefine 来保存对解决方案的改进。"
// 		if IterPlan != nil && len(IterPlan.TopicToDiscuss) > 0 {
// 			TopicToDiscuss = IterPlan.TopicToDiscuss
// 		}
// 		// 自主规划没有足够的覆盖度。需要随机选择来保证覆盖度
// 		if len(SuperEdgeSelected) == 0 || rand.IntN(100) < 50 {
// 			edges := append([]*SolutionGraphNode{}, constraints...)
// 			mutable.Shuffle(edges)
// 			SuperEdgeSelected = edges[:min(3, len(edges))]
// 		}

// 		Communities, _, _ := milvusCollection.SearchVectors(lo.Map(SuperEdgeSelected, func(v *SolutionGraphNode, _ int) []float32 { return v.Embed() }), qmilvus.SearchParamsDefault.WithTopK(50))
// 		CommunityNodes := lo.Map(Communities, func(community []*SolutionGraphNode, i int) []*SolutionGraphNode {
// 			return EdgeCommunitiesWithExploration(businessPlans, SuperEdgeSelected[i], community, 10)
// 		})

// 		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
// 		newNodes := lo.Filter(lo.Values(businessPlans), func(v *SolutionGraphNode, _ int) bool { return !v.Locked })
// 		newNodesSorted := slices.Clone(newNodes)
// 		fmt.Printf("NewNodes Best: %v\n", newNodesSorted[len(newNodesSorted)-1].String())
// 		CommunityNodesStr_ := SolutionGraphNodeList(lo.Flatten(CommunityNodes)).Uniq().FullView()
// 		constraintsStr_ := SolutionGraphNodeList(constraints).Uniq().LockedOnly().FullView()
// 		CommunitiesAllStr := SolutionGraphNodeList(lo.Flatten(CommunitiesAll)).Uniq().LockedOnly().FullView()

// 		go func(Communities, SuperEdges, TopicToDiscuss, CommunitiesAll string, AllItems map[string]*SolutionGraphNode, newNodes SolutionGraphNodeList) {
// 			defer func() { <-MaxThreadsSemaphore }()
// 			//models.Qwq32B, models.Gemma3, models.DeepSeekV3, models.DeepSeekV3TB,models.Qwen32B,models.GeminiTB,models.Qwen30BA3,models.GLM32B
// 			// err := AgentBusinessPlans.WithTools(ToolDroneBotIterPlan, ToolDroneBotSolutionItemRefine).WithModels(models.Qwen32B).CopyPromptOnly().Call( map[string]any{
// 			// 	"Communities":    Communities,
// 			// 	"SuperEdges":     SuperEdges,
// 			// 	"TopicToDiscuss": TopicToDiscuss,
// 			// 	"task":           "batchElo",
// 			// })
// 			err := AgentBusinessPlans.WithTools(ToolDroneBatchEloResults).WithModels(models.EloModels.SequentialPick(models.Gemma3B27)).CopyPromptOnly().Call( map[string]any{
// 				"Communities":    Communities,
// 				"CommunitiesAll": CommunitiesAll,
// 				"SuperEdges":     SuperEdges,
// 				"TopicToDiscuss": TopicToDiscuss,
// 				"task":           "super_edge_evolution", //super_edge_evolution  batchElo gen
// 				"NewNodes":       newNodes.Uniq().FullView(),
// 			})
// 			if err != nil {
// 				fmt.Printf("Agent call failed: %v\n", err)
// 			}
// 		}(CommunityNodesStr_, constraintsStr_, TopicToDiscuss, CommunitiesAllStr, businessPlans, newNodesSorted)
// 	}
// 	// Wait for all the goroutines to finish)
// 	for i := 0; i < MaxThreads; i++ {
// 		MaxThreadsSemaphore <- struct{}{}
// 	}

// }

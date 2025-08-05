package projects

// type PersonalAdmin struct {
// 	Id        string   `description:"Id, string, unique"`
// 	Questions []string `description:"Question,array"`
// }

// func (u *PersonalAdmin) GetId() string {
// 	return u.Id
// }

// func (u *PersonalAdmin) ScoreAccessor(delta ...int) int {
// 	var eloDelta int = append(delta, 0)[0]
// 	return mixincached.WithElo("projects", keyPersonalAdmin.Key, 1000).ScoreAccessor(u.Id, float64(eloDelta))
// }

// var keyPersonalAdmin = redisdb.HashKey[string, *PersonalAdmin](redisdb.Opt.Rds("projects"))

// // 为什么Qwen能自我改进推理，Llama却不行 https://mp.weixin.qq.com/s/OvS61OrDp6rB-R5ELg48Aw
// // 并且每次就一个确切的改进方向进行深度分析，反代的深度分析第一性原理之上的需求，深度创新以做出实质的改进。要痛恨泛泛而谈的内容，重复空洞的内容，因为现在是在开发世界级的工具。
// var AgentPersonalAdmin = agent.NewAgent(template.Must(template.New("utilifyFunction").Parse(`
// 我想要用填写开放问题的方式来最大化一天的时间效用。
// 有哪些问题应当保持在清单上面？
// 避免超过15个问题。否则我将浪费太多时间在填表上。

// 这是现有的方案 ：
// {{range  $index, $item := .ItemList}}
// 第{{$index}}个方案：
// Id:{{$item.Id}}
// Questions:
// {{range $index, $value := $item.Questions}}"{{$value}}}"
// {{end}}
// {{end}}

// ToDo:
// 步骤1. 对现有的选项进行思考和评估：
// 	1、对回溯或在检测到错误进行显式修改（例如，「这种方法行不通，因为...」）；
// 	2、验证或系统地检查中间结果（例如，「让我们通过... 来验证这个结果」）；
// 	3、子目标设定，即将复杂问题分解为可管理的步骤（例如，「要解决这个问题，我们首先需要...」）；
// 	4、逆向思考，即在目标导向的推理问题中，从期望的结果出发，逐步向后推导，找到解决问题的路径。（例如，「要达到 75 的目标，我们需要一个能被... 整除的数字」）。

// {{if gt .TotoalNodes 9}}
// 步骤2. 根据上面的讨论. 将现有方案选项从最好到最坏排序(用Id列表表示)。调用FunctionCall:SaveItemsSorted 保存排序结果。
// {{end}}

// {{if lt .TotoalNodes 10}}
// 步骤2. 在讨论的基础上，进一步提出一个更好的，新的方案草稿。对草稿同样按步骤1改进后，调用FunctionCall:NewPersonalAdmin 保存新的方案。
// {{end}}
// `))).WithToolCallLocked().WithTools(tool.NewTool("SaveItemsSorted", "Save sorted Items, Items represented as Id list.", func(model *prototype.ItemsSorted) {
// 	if model == nil || len(model.ItemsRefById) == 0 {
// 		return
// 	}
// 	all, _ := keyPersonalAdmin.HGetAll()
// 	sortedElos := elo.ToSlice(all).Sort().Reverse()
// 	elo.BatchUpdateRanking(sortedElos.TakeByIds(model.ItemsRefById...)...)
// 	//remove the worst 2
// 	for _, v := range lo.Filter(sortedElos, func(v elo.Elo, i int) bool { return v.ScoreAccessor() < 996 }) {
// 		keyPersonalAdmin.ConcatKey("Expired").HSet(v.GetId(), v)
// 		keyPersonalAdmin.HDel(v.GetId())
// 	}
// 	for i, v := range sortedElos[0:min(10, len(sortedElos))] {
// 		fmt.Println("Best Model,top ", i+1, v.GetId(), "Elo", v.ScoreAccessor())
// 	}

// })).WithTools(tool.NewTool("NewPersonalAdmin", "create a new NewPersonalAdmin plan", func(model *PersonalAdmin) {
// 	model.Id = redisdb.NanoId(8)
// 	keyPersonalAdmin.HSet(model.Id, model)
// }))

// func PersonalAdminExploration() {
// 	const MaxThreads = 14
// 	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)
// 	keyPersonalAdmin.HSet("default", &PersonalAdmin{Id: "default", Questions: []string{"今天我最期待完成的三件事是什么？ （优先级& 积极关注）", "为了精力充沛地完成这些事，我今天需要具备怎样的状态？ （状态准备& 内在驱动）",
// 		"今天可能遇到的最大挑战是什么？我该如何提前应对？ （风险预判& 应对准备）", "我现在进展如何？是否需要调整我的计划或优先级？ （进度监控& 计划弹性）",

// 		"我的能量水平如何？我需要休息一下，或者做些什么来恢复精力吗？ （能量觉察& 自我关怀）",

// 		"今天我最大的成就是什么？ （成就肯定& 自信提升）",

// 		"今天我学到了什么？有哪些经验或教训可以应用到明天？ （经验总结& 持续改进）",

// 		"为了明天更高效地工作，我今天可以做哪些准备？ （提前准备& 效率提升）", "今天我为自己做了哪些积极的事情，让自己感觉更好？ （自我关怀& 积极心理）"}})

// 	for i, TotalTasks := 0, 1000*1000; i < TotalTasks; i++ {
// 		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
// 		go func() {
// 			defer func() { <-MaxThreadsSemaphore }()
// 			best, _ := keyPersonalAdmin.HGetAll()
// 			//put new nodes to the top
// 			NodesWithPriority := lo.Filter(lo.Values(best), func(v *PersonalAdmin, i int) bool {
// 				return v.ScoreAccessor() > 996 && v.ScoreAccessor() < 1004
// 			})
// 			selected := lo.Union(NodesWithPriority, lo.Values(best))
// 			lo.Shuffle(selected[len(NodesWithPriority):])
// 			param := map[string]any{
// 				"ItemList": selected[:min(6, len(selected))],
// 				//[]*models.Model{models.FuseO1, models.DeepSeekR132B}[i%2],
// 				"Model":       models.LoadbalancedPick(models.FuseO1),
// 				"TotoalNodes": len(best),
// 			}
// 			err := AgentPersonalAdmin.Call(context.Background(), param)
// 			if err != nil {
// 				fmt.Printf("Agent call failed: %v\n", err)
// 			}
// 		}()
// 	}
// 	// Wait for all the goroutines to finish)
// 	for i := 0; i < MaxThreads; i++ {
// 		MaxThreadsSemaphore <- struct{}{}
// 	}

// }

package projects

// import (
// 	"context"
// 	"fmt"
// 	"slices"
// 	"strings"
// 	"text/template"

// 	"github.com/doptime/eloevo/agent"
// 	"github.com/doptime/eloevo/models"
// 	"github.com/doptime/eloevo/prototype"
// 	"github.com/doptime/eloevo/tool"
// 	"github.com/doptime/eloevo/utils"
// 	"github.com/doptime/redisdb"
// 	"github.com/samber/lo"
// )

// type BusinessPlans struct {
// 	Id       string         `description:"Id, string, unique"`
// 	Votes    int64          `description:"-"`
// 	Item     string         `description:"item of the solution"`
// 	ParentId string         `description:"Parent Id this item belongs to, string, empth as root"`
// 	Extra    map[string]any `msgpack:"-" json:"-" description:"-"`
// }

// func (u *BusinessPlans) GetId() string {
// 	return u.Id
// }
// func (u *BusinessPlans) ScoreAccessor(delta ...int) int {
// 	eloDelta := append(delta, 0)[0]
// 	if eloDelta >= -5 && eloDelta <= 5 && eloDelta != 0 {
// 		u.Votes += int64(eloDelta)
// 		if key, ok := u.Extra["Key"].(*redisdb.HashKey[string, *BusinessPlans]); ok {
// 			if u.Votes < 0 {
// 				key.ConcatKey("Expired").HSet(u.Id, u)
// 				key.HDel(u.Id)
// 			} else {
// 				key.HSet(u.Id, u)
// 			}
// 		}
// 	}
// 	return int(u.Votes)
// }
// func (u *BusinessPlans) String(layer ...int) string {
// 	numLayer := append(layer, 0)[0]
// 	indence := strings.Repeat("\t", numLayer)
// 	childrenStr := strings.Builder{}
// 	itemList, _ := u.Extra["ItemList"].([]*BusinessPlans)
// 	for _, v := range lo.Filter(itemList, func(v *BusinessPlans, i int) bool { return v.ParentId == u.Id }) {
// 		childrenStr.WriteString("\n" + v.String(numLayer+1))
// 	}
// 	return fmt.Sprint(indence, "Id:", u.Id, " Votes:", u.Votes, " Item:", u.Item, childrenStr.String())
// }

// var keyBusinessPlans = redisdb.NewHashKey[string, *BusinessPlans](redisdb.Opt.Rds("Catalogs"))

// var AgentBusinessPlans = agent.NewAgent(template.Must(template.New("utilifyFunction").Parse(`
// 你是集 “创业生态架构师”、“技术趋势预言家”、“商业模式创新专家” 三位一体的连续创业家。
// 目标是通过 “寻找未被满足的市场需求”、“发现技术创新带来的机会”、“预测未来趋势”和其它的动态认知框架，深入分析商业领域 “{{.Topic}}”，找出商业领域(Industry) "{{.Topic}}"下创业项目。
// 最终目标是获得在该领域创业需要的高价值、前瞻性的，既有战略深度又具备创新活力的创业项目矩阵。这些创业项目以层次化的方式构建。
// 预期这些创业项目/BusinessPlans在接下来的世界中，在商业领域"{{.Topic}}"能够产生最大化的联合的商业效用，以捕获该行业中的主要机会。

// 这是现有的方案 ：
// {{range  $index, $item := .RootList}}
// {{$item.String}}
// {{end}}

// ToDo:
// 现在，假定你采用reddit用户的投票方式，对上面的方案的选项进行思考和评估后进行投票。投票将提升或降低选项的优先序，投票值为[-5,5]之间的整数;不必对所有选项投票，而是需要对需要调整排序的选项进行投票；票数低于0的项目将被自动删除：
// 先对现有的选项组成的方案进行思考和评估：
// 	1、对回溯或在检测到错误进行显式修改；该需求是否是基于错误的幻想或者错误的假设；格式，内容是否异常；
// 	2、验证或系统地检查中间结果；看看从第一性原理出发，这个需要是否可以被绕过或者替代；是否属于死愚蠢的需求；是否在更多票数的条目中已经包含，属于冗余条目；
// 	3、子目标设定，即将复杂问题分解为可管理的步骤；需求是否需要进一步细化，以便更好地建构；
// 	4、逆向思考，即在目标导向的推理问题中，从期望的结果出发，逐步向后推导，找到解决问题的路径:
// 	基于BusinessUtilityFunction=exp(WeightMarketSizeln(MarketSize) + 0.18ln(MarketGrowthRate) + 0.22ln(ExpectedReturn) + 0.10ln(TechnicalFeasibility) + 0.15ln(InnovationPotential) + 0.080ln(ResourceAllocation) - 0.12ln(ProjectRisk + 1) - 0.080ln(CompetitionIntensity) - 0.10ln(ImplementationDifficulty) + 0.060ln(TimeToMarket) + 0.040ln(TeamExperience) + 0.050ln(PolicySupport))  方案是否有重大缺陷。

// 	- 在讨论的基础上，投票以修改解决方案选项的权重（排序），请优先考虑删除劣质条目以优化方案，形成ProConsToItems。
// 	- 按照讨论。如果存在改进解决方案的可能，请提出新的Items. 请直接补充描述0条或者多条Items，形成NewProposedItems。
// 最后调用FunctionCall:SolutionRefine 保存排序结果。
// `))).WithToolCallLocked().WithTools(tool.NewTool("SolutionRefine", "Vote on items to refine solution; Propose new solution item to parent Item if needed.", func(model *prototype.SolutionRefine) {
// 	hashKey, ok := model.Extra["Key"].(*redisdb.HashKey[string, *BusinessPlans])
// 	if !ok || model == nil {
// 		return
// 	}
// 	all, _ := hashKey.HGetAll()
// 	for k, v := range model.ProConsToItems {
// 		if item, ok := all[k]; ok {
// 			item.Extra = model.Extra
// 			item.ScoreAccessor(v)
// 		}
// 	}
// 	for ParentId, v := range model.NewProposedItems {
// 		if _, ok := all[ParentId]; !ok || ParentId == "root" {
// 			ParentId = ""
// 		}
// 		if len(v) > 8 {
// 			item := &BusinessPlans{Votes: 1, Item: v, Id: utils.ID(v), ParentId: ParentId}
// 			hashKey.HSet(item.Id, item)
// 		}
// 	}
// }))

// func BusinessPlansExploration() {
// 	// Create a new weighted chooser
// 	const MaxThreads = 1
// 	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

// 	for i, TotalTasks := 0, 1000*1000; i < TotalTasks; i++ {
// 		MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
// 		go func() {
// 			//industry := IndustryList[rand.Intn(len(IndustryList))]
// 			industry := "Synthetic Biology Platforms"
// 			defer func() { <-MaxThreadsSemaphore }()
// 			key := keyBusinessPlans.ConcatKey(industry)
// 			best, _ := key.HGetAll()
// 			// for k, v := range best {
// 			// 	var k1 string
// 			// 	if err := msgpack.Unmarshal([]byte(k), &k1); err == nil {
// 			// 		key.HDel(k)
// 			// 		key.HSet(k1, v)
// 			// 	}
// 			// }
// 			listSorted := lo.Values(best)
// 			slices.SortFunc(listSorted, func(a, b *BusinessPlans) int {
// 				return -(a.ScoreAccessor() - b.ScoreAccessor())
// 			})
// 			RootList := lo.Filter(listSorted, func(v *BusinessPlans, i int) bool {
// 				return v.ParentId == ""
// 			})
// 			param := map[string]any{
// 				"RootList": RootList,
// 				"Key":      key,
// 				"Topic":    industry,
// 				"ItemList": listSorted,
// 				"TotoalNodes": len(best),
// 			}
// 			for _, v := range listSorted {
// 				v.Extra = param
// 			}
// 			//print the lefts
// 			for i, v := range RootList {
// 				fmt.Println("Rank", i+1, v.String())
// 			}
//models.Qwq32B, models.Gemma3, models.DeepSeekV3
// 			err := AgentBusinessPlans.WithModel(models.Qwq32B).Call(context.Background(), param)
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

// // Autonomous Systems & Mobility
// var IndustryList = []string{
// 	"Healthcare & Longevity (合并医疗创新、药物发现、个性化医疗等)", "AI Core Technologies", "Circular Economy & Resource Management (整合回收、材料科学、可持续制造)", "Hydrogen & Clean Energy Solutions", "Cybersecurity Infrastructure & Quantum Security", "Advanced Battery Technology & Energy Storage", "Bioeconomy (合成生物、生物制造、可持续材料等)", "Next-Generation Semiconductor Technologies", "Smart Energy & Grid Management", "Smart Infrastructure (城市、交通、水资源、能源网格的智能化整合)", "Synthetic Biology & Bioengineering", "Carbon Capture, Utilization, and Storage (CCUS)", "Quantum Computing & Sensing", "Advanced Materials & Nanotechnology", "Sustainable Agriculture & Regenerative Farming", "Sustainable Water Management & Purification", "Carbon Removal Technologies", "Digital Infrastructure & Connectivity", "Climate Change Adaptation Technologies", "Smart Agriculture & Farming Tech", "Smart Manufacturing & Industry 4.0", "Decentralized Renewable Energy Grids", "AI Infrastructure & Compute", "Climate Resilience Infrastructure", "Supply Chain Resilience & Optimization", "Food Tech & Alternative Proteins", "AI-Driven Climate & Environmental Modeling", "Climate Tech & Environmental Remediation", "Personalized Medicine & Genetic Therapies", "Renewable Energy Integration & Storage", "Regenerative Medicine", "AI-Powered Cybersecurity", "Green Chemistry & Industrial Ecology", "Sustainable Materials Science", "Sustainable Chemistry & Green Materials", "Autonomous Vehicle Technologies", "Climate-Positive Agriculture", "Fundamental Materials Science & Engineering", "Robotics and Automation", "Global Logistics & Transportation Infrastructure", "Fusion Power", "Climate Modeling and Analytics", "Data Privacy and Security Technologies", "Financial Technology (FinTech) Infrastructure", "Traditional Industry Digital Transformation (制造、建筑、农业等基础行业数字化转型)", "Computational Biology and Bioinformatics", "Sustainable Urban Mobility", "Digital Identity & Privacy Technologies", "Biomanufacturing & Synthetic Biology", "Digital Economy & Platform Business Models", "Ethical AI & Governance", "Global Health Security & Pandemic Preparedness", "Advanced Energy Storage Solutions", "AI-Driven Materials Discovery", "Advanced Energy Storage", "Carbon Neutrality Integration Technologies", "AI-Augmented Scientific Discovery", "Global Water Security Technologies", "Advanced Robotics & Embodied AI", "Next-Generation Power Transmission", "Industrial IoT & Connected Operations", "Sustainable Food Production Systems", "Decentralized Energy Grids", "Smart Cities & Urban Tech", "Industrial IoT & Digital Twins", "FinTech & Digital Banking Infrastructure", "Precision Fermentation & Cellular Agriculture", "Personalized Healthcare & Diagnostics", "Construction & Smart Buildings", "Advanced Manufacturing & Robotics", "Supply Chain & Logistics Technologies", "Financial Services & Digital Banking", "Global Water Stress Solutions", "EdTech & Future of Learning", "Industrial Automation & Cobotics", "Sustainable Infrastructure Development", "Post-Lithium Battery Technologies", "Resilient Infrastructure & Disaster Recovery", "Resilient Food Systems & Supply Chains", "Sustainable Food Production Systems (Beyond AgTech)", "Global Essential Goods Production (食品、药品等基础民生商品生产)", "Resource Security & Geopolitical Risk Management", "Advanced Nuclear Fission & Generation IV Reactors", "Precision Nutrition and Microbiome Engineering", "AI-Driven Supply Chain Optimization & Resilience", "Climate Adaptation Infrastructure", "Quantum-Safe Cryptography", "AI-Driven Automation Solutions", "Extreme Climate Adaptation Systems", "Carbon-Neutral Energy Transition Technologies", "Global Supply Chain Reconstruction Technologies", "Sustainable Water Technologies", "Ethical AI & Responsible Innovation", "AI-Powered Precision Medicine", "AI Safety and Alignment Research", "Mental Health Technologies", "Biomanufacturing Scale-up and Commercialization", "pL3: AgriTech for Climate Adaptation", "Biomanufacturing at Scale", "Digital Biology", "Climate Finance & Investment", "Climate-Resilient Infrastructure Materials", "Quantum Information Science", "Sustainable Packaging Solutions", "Essential Consumer Goods Manufacturing", "Basic Education & Workforce Development", "Critical Mineral Extraction and Processing", "Education Infrastructure & Public Learning Systems", "Blue Economy & Marine Tech", "Advanced Nuclear Energy (SMRs, Fusion)", "Biomanufacturing & Scale-Up", "Robotics as a Service (RaaS)", "Silver Economy & Aging Tech", "AI-Driven Optimization & Efficiency", "Carbon Accounting and Verification Technologies", "Traditional Energy Transition Technologies", "Universal Basic Infrastructure (整合水、电、通信等基础民生基础设施)", "AI-Driven Robotics Development", "Climate Risk Modeling & Insurance", "Brain-Computer Interfaces & Neurotechnology", "AI Applications in Materials Science", "Green Hydrogen Production & Infrastructure", "Ethical AI Deployment & Governance", "Sustainable Chemical Engineering", "Regenerative Technologies", "Ocean Carbon Sequestration Technologies", "Ethical Supply Chain & Fair Trade", "Bio-Integrated Technology",
// }

package agents

import (
	"context"
	"fmt"
	"sync"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/elo"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/module"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/mroth/weightedrand"
	"github.com/samber/lo"
)

var AgentGenTestModule = agent.NewAgent(template.Must(template.New("question").Parse(`
你是一个商业方案迭代器，用测试和开发一体化的方式迭代改进目标系统。
目标是为AGI时代的商业方案找出真实的需求场景作为测试场景。然后1. 围绕测试场景，分析真实需求; 2. 引用相关模块; 3. 如果需要则开发或完善相关的模块; 4.编译测试所需要的模块; 5.调试运行整个系统。
开发主题：
	AGI时代使用物流无人机/作为机器人载具的无人机会有丰富的应用场景。
这个无人机平台目标起飞重量在25kg. 
现在你需要分析这个需求:
{{.BusinessSenarioCatalogue}}

- 步骤一:生成一个互补的，确切的测试场景。
这些是已有的测试场景名称：
{{range .TestScenarios}}
{{.}}
{{end}}
	你需要按照MECE原则生成一个互补的，确切的测试场景。

- 步骤二: 提出真实的测试需求是什么。
然后你需要分析场景背后的真实测试需求
比如，如果一个测试场景中涉及25kg的固定翼无人机从快递店起飞进行物流配送。那么这个真实测试需求就是验证垂直起飞能力。
这些是现有的真实测试需求名称：
{{range .RequirementNames}}
{{.}}
{{end}}


- 步骤三: 把真实测试需求转变成为系统的模块组件需求。
现在，为了使得你生成的需求能够被建构成为一个系统，能够在软硬件和流程上得以正常运行，并且测试能够有效集成到开发过程中。让我们看看还需要如何调整现有的模块，或者如何才能使得现有的模块更简单通用可（免）维护。

	这些是现有的模块名称：
	{{range .Modules}}
	{{.}}
	{{end}}
让我们深入分析场景背后的真实测试需求，看看为实现理想的系统，在模块化需求层面我们还能够做什么。
如果我们期望的模块已经被良好构建，那么我们就把对应的模块选择到目标系统测试模块中。
如果我们期望的模块并未并未被良好构建，，或是你有一个了不起的想法，可以简化或改进系统。那么原则上，我们就在本次迭代当中提出或改进有且只有一个模块化。

模块的描述包括：名称，解决的问题，实现思路，技术方案，以及模块的功能和特性。
你需要小心的评估现有的模块。如果现有的模块不能满足这些需求，你需要提出新的模块化需求。
如果需要对现有的模块做调整和改进，你需要提出这些调整和改进的需求，生成新的竞争性的需求。
请注意你需要像Elon Musk一样，提出一些真正的创新性的需求，努力杀死愚蠢的需求。90%的需求都是愚蠢的。
比如垂直起飞的测试需求。提出的模块就包括：
	模块名称:垂直起飞模块
	- 解决的问题: 无人机在狭小空间起飞
	- 实现思路: 无人机机头向上，螺旋桨向下，提供升力
	- 技术方案: 1.无人机收到起飞指令后，2.规划升空路径。3.沿着路径飞行并自动调整姿态
	模块名称: 外接电源模块
	- 解决的问题: 无人机起飞时候电池功率不足
	- 实现思路: 无人机在起飞前连接外接电源，提供4kw的电力。
	- 技术方案: 1.连接线采用16AWG软硅胶线，2.电压72V，3.要求人工连接外接电源，4.无人机主板自动切换电源模块。...
	

你需要围绕这个
市场需求: 确保项目有明确的市场需求和增长潜力。
这个场景需求将被用来分析以指导无人机模块的设计和开发工作。




约束条件:
	- 生成的商业场景需求需要细粒度的。这些需求的细节将用于开发无人机的软硬件方案。所以细节需要有助于思考和分析无人机的各种特性，以确保无人机能够适应各种环境和任务。
	- 尽可能创建更丰富的，符合真是商业场景的需求。因为开发的方案需要面向这些需求进行测试。
	- 生成的需求要符合MECE原则，即互相独立，完全穷尽，不重复。
	- 特性的需求是多样化的
	- 由于任务将被上百万次地抽样运行，生成的每个机会必须彼此独特，避免重复
	- 每次仅生成一个具体的测试场景, 尽可能引入设计无人机模块、物理特性、以及无人机和环境相互作用的描述。
	- 生成的商业场景需求名称 和 商业场景需求描述 要求采用单行文本描述，不要使用多行文本描述。
生成的需求特性描述举例:
	- 比如 从福州飞厦门.起飞信息:电池电量1kwh, 续航300km，电池状态正常。起飞地点:湖里区五缘西一里快递驿站。 天气预报: 风力5m/s,晴朗.螺旋桨数据:正常。路线规划: 福州->泉州->厦门。说明，在泉州利用地形提高飞行高度200m. 货物: 25kg 物流件。
	
	-
	测试需求类别: 产品/服务: 产品的独特性、技术可行性和用户体验至关重要
	商业场景需求名称: 医疗紧急物资空投配送
	商业场景需求描述: 无人机需在灾区或偏远地区进行快速医疗物资配送，配备精确定位系统、稳定的气候适应性飞行模块，以及安全的载物机制，确保急需物资能够及时、安全地送达指定地点。

	- 
	商业场景需求名称: 无人机自主避障与路径优化
	商业场景需求描述: 无人机在复杂城市环境中执行物流任务时，需具备实时障碍物识别与避让能力，结合AI路径规划算法，动态优化飞行路线，确保高效且安全的配送过程。

	-
	商业场景需求名称: 实时无人机状态监测与诊断
	商业场景需求描述: 无人机需配备全面的传感器系统，实现对飞行状态、电池寿命、发动机性能等关键参数的实时监测与自动诊断，确保在任务执行过程中能够及时发现并应对潜在故障，提升飞行安全性和可靠性。
	
	-
	商业场景需求名称: 无人机法律合规飞行设计
	商业场景需求描述: 无人机设计需遵循各地区空域管理法规、隐私保护法和数据传输规范，集成地理围栏功能、自动许可验证模块，并确保飞行数据加密传输，以保障无人机运营的合法性和用户隐私安全。
	
返回格式说明 
	商业场景需求描述以两个换行符结束
商业场景需求名称: 为该商业场景需求命名,名称需要能够描述需求的内容
商业场景需求描述: 由模型自行决定需要包含哪些需求描述



请面向目标意图，为列表中的内容，依照实现目标意图所需要的基于MECE原则的方案拆解。为这些方案排名。请为这些方案中实质雷同的方案中的劣质版本

   `))).WithCallback(func(ctx context.Context, inputs string) error {
	keyAircraftRequirementName := redisdb.HashKey[string, string](redisdb.WithKey("AircraftRequirements"))
	name := utils.ExtractTagValue(inputs, "需求名称")
	keyAircraftRequirementName.HSet(name, inputs)
	return nil
})

var keyAircraftTests = redisdb.HashKey[string, *Tests](redisdb.WithKey("AircraftTests"))
var keyAircraftModuels = redisdb.HashKey[string, *module.Module](redisdb.WithKey("AircraftModules"))

type Tests struct {
	elo.Elo
	Catalogue    string
	Name         string
	Descriptions string
	Modules      []*module.Module
}

func (t *Tests) String() string {
	return fmt.Sprintf("Catalogue: %s, Name: %s, Descriptions: %s", t.Catalogue, t.Name, t.Descriptions)
}

var AircraftTests = map[string]*Tests{}

// GenNicheMarketOpportunityParallel calls GenNicheMarketOpportunity 1000 times in 16 parallel threads.
func GenModuleParallel() (err error) {
	const numThreads = 16
	const numCallsPerThread = 1000 * 1000 / numThreads
	AircraftTests, _ = keyAircraftTests.HGetAll()
	AircraftModules, _ := keyAircraftModuels.HGetAll()

	choices := []weightedrand.Choice{
		{Item: "市场需求: 确保项目有明确的市场需求和增长潜力", Weight: 15},
		{Item: "产品/服务: 产品的独特性、技术可行性和用户体验至关重要", Weight: 15},
		{Item: "商业模式: 盈利模式和成本结构的合理性直接影响项目的可持续性", Weight: 15},
		{Item: "团队能力: 强大的团队是执行和应对挑战的基础", Weight: 15},
		{Item: "竞争与壁垒: 了解竞争环境并建立竞争优势，确保项目能够在市场中立足", Weight: 10},
		{Item: "财务与资源: 充足的资金和资源是项目启动和扩展的保障", Weight: 10},
		{Item: "法律与合规: 确保项目在法律框架内运作，避免潜在的法律风险", Weight: 5},
		{Item: "市场进入策略: 有效的市场进入策略能够加速项目的市场渗透", Weight: 5},
		{Item: "技术与创新: 技术优势和创新能力能够提升项目的竞争力", Weight: 5},
		{Item: "用户获取与营销: 高效的用户获取和营销策略有助于快速扩大用户基础", Weight: 5},
		{Item: "可持续性与社会影响: 确保项目具备长期可持续发展，并产生积极的社会影响", Weight: 3},
		{Item: "风险管理: 有效的风险管理能够降低项目失败的可能性", Weight: 2},
	}
	chooser, _ := weightedrand.NewChooser(choices...)

	var wg sync.WaitGroup
	wg.Add(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				var inputsParams = map[string]any{
					"BusinessSenarioCatalogue": chooser.Pick().(string),
					"TestScenarios":            lo.Keys(AircraftTests),
					"Modules":                  lo.Keys(AircraftModules), // 假设有一个获取模块列表的函数
				}
				fmt.Println("GenNicheMarketOpportunity...")
				//return AgentGenNicheMarketOpportunity.WithModel(models.EloModels.SelectOne("roundrobin")).		Call(context.Background(), inputsParams)
				AgentGenTestModule.WithModel(models.EloModels.SelectOne("roundrobin")).Call(context.Background(), inputsParams)
			}
		}()
	}
	wg.Wait()
}

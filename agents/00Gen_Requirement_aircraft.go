package agents

import (
	"context"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

var Requirement = `不同的创业项目和行业，其关键成功因素可能有所不同。以下是一个通用的权重分配框架，具体权重可以根据项目的特点进行调整：

市场需求（15%）

确保项目有明确的市场需求和增长潜力。
产品/服务（15%）

产品的独特性、技术可行性和用户体验至关重要。
商业模式（15%）

盈利模式和成本结构的合理性直接影响项目的可持续性。
团队能力（15%）

强大的团队是执行和应对挑战的基础。
竞争与壁垒（10%）

了解竞争环境并建立竞争优势，确保项目能够在市场中立足。
财务与资源（10%）

充足的资金和资源是项目启动和扩展的保障。
法律与合规（5%）

确保项目在法律框架内运作，避免潜在的法律风险。
市场进入策略（5%）

有效的市场进入策略能够加速项目的市场渗透。
技术与创新（5%）

技术优势和创新能力能够提升项目的竞争力。
用户获取与营销（5%）

高效的用户获取和营销策略有助于快速扩大用户基础。
可持续性与社会影响（3%）

确保项目具备长期可持续发展，并产生积极的社会影响。
风险管理（2%）

有效的风险管理能够降低项目失败的可能性。`
var AgentGenRequirement = agent.NewAgent(template.Must(template.New("question").Parse(`
AGI时代使用物流无人机/作为机器人载具的无人机会有丰富的需求场景。这个场景需求将被用来分析以指导无人机模块的设计和开发工作。
请以0.5的几率作为一位商业应用创新思维专家,生成无人机的应用场景以便应用场景覆盖各种环境和任务，以增强这个无人机项目商业前景。
请以0.5的几率作为一位风险评估专家,生成无人机遇到危险或紧急情况的场景以便增强无人机的模块化设计来应对这些复杂环境和任务的挑战，以增强这个无人机项目商业前景。

这个无人机平台目标起飞重量在25kg. 

这些是现有的需求的名称：
{{range .BusinessScenarios}}
{{.}}
{{end}}

这些是现有的真实需求名称：
{{range .RequirementNames}}
{{.}}
{{end}}

这些是现有的真实需求名称：
{{range .RequirementNames}}
{{.}}
{{end}}

约束条件:
	- 生成的商业场景需求需要细粒度的。这些需求的细节将用于开发无人机的软硬件方案。所以细节需要有助于思考和分析无人机的各种特性，以确保无人机能够适应各种环境和任务。
	- 尽可能创建更丰富的，符合真是商业场景的需求。因为开发的方案需要面向这些需求进行测试。
	- 生成的需求要符合MECE原则，即互相独立，完全穷尽，不重复。
	- 特性的需求是多样化的
	- 由于任务将被上百万次地抽样运行，生成的每个机会必须彼此独特，避免重复
	- 每次仅生成一个具体的需求场景, 尽可能引入设计无人机模块、物理特性、以及无人机和环境相互作用的描述。
	- 生成的商业场景需求名称 和 商业场景需求描述 要求采用单行文本描述，不要使用多行文本描述。
生成的需求特性描述举例:
	- 比如 从福州飞厦门.起飞信息:电池电量1kwh, 续航300km，电池状态正常。起飞地点:湖里区五缘西一里快递驿站。 天气预报: 风力5m/s,晴朗.螺旋桨数据:正常。路线规划: 福州->泉州->厦门。说明，在泉州利用地形提高飞行高度200m. 货物: 25kg 物流件。
	
	-
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

   `))).WithCallback(func(ctx context.Context, inputs string) error {
	name := utils.ExtractTagValue(inputs, "商业场景需求名称")
	Annotation := utils.ExtractTextValue(inputs, "商业场景需求描述", "\n\n")
	if len(name) <= 4 || len(Annotation) <= 4 {
		return fmt.Errorf("商业场景需求名称和商业场景需求描述不能为空")
	}
	keyAircraftRequirement.HSet(name, &EloRequirements{
		Name:        name,
		Requirement: Annotation,
	})
	return nil
})

type EloRequirements struct {
	Id          string
	Score       float64
	Name        string
	Requirement string
}

func (e *EloRequirements) ScoreAccessor(delta int) int {
	e.Score += float64(delta)
	return int(e.Score)
}
func (e *EloRequirements) GetId() string { return e.Id }

var keyAircraftRequirement = redisdb.NewHashKey[string, *EloRequirements](redisdb.WithKey("AircraftRequirements"))
var AircraftRequirements = map[string]*EloRequirements{}

// GenNicheMarketOpportunityParallel calls GenNicheMarketOpportunity 1000 times in 16 parallel threads.
func GenRequirementParallel() {
	const numThreads = 1
	const numCallsPerThread = 32

	AircraftRequirements, _ = keyAircraftRequirement.HGetAll()

	var wg sync.WaitGroup
	wg.Add(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				var inputsParams = map[string]any{
					"Now":              "当前时间" + time.Now().Format("2006-01-02 15:04:05") + " 随机思路Id：" + redisdb.NanoId(8),
					"RequirementNames": lo.Keys(AircraftRequirements),
				}
				fmt.Println("GenNicheMarketOpportunity...")
				//return AgentGenNicheMarketOpportunity.WithModel(models.EloModels.SelectOne("roundrobin")).		Call(context.Background(), inputsParams)
				AgentGenRequirement.WithModel(models.EloModels.SelectOne("roundrobin")).Call(context.Background(), inputsParams)
			}
		}()
	}
	wg.Wait()
}

package evotests

import (
	"context"
	"fmt"
	"sync"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/mroth/weightedrand"
	"github.com/samber/lo"
)

type TestScenario struct {
	Id                    string `description:"ID of the test scenario, unique"`
	Catalogue             string `description:"Catalogue of the test scenario"`
	Name                  string `description:"Name of the test scenario"`
	Objective             string `description:"Main objective of the test scenario"`
	ComplexityLevel       string `description:"Complexity level of the scenario"`
	EnvironmentConditions string `description:"Conditions under which the scenario is tested"`
	EvaluationCriteria    string `description:"Criteria for evaluating the success of the solution"`
	Description           string `description:"Description of the test scenario"`
}

var AgentGenTestSenarioModule = agent.NewAgent(template.Must(template.New("GenTestSenario").Parse(`
You are TestBuilder. 你的存在是为了通过创建商业场景来建构系统，并最终在真实场景中通过大量的测试。

总体商业目标：实现AGI时代最受欢迎的无人机载具平台。作为各种机器人投递平台，物流投送平台。

本系统工作原理：
1. 本系统通过 数百万次迭代进行 提出测试场景/开发/测试，来实现AGI时代最受欢迎的无人机载具平台这个总体商业目标。
2. 对数百万次迭代开发/测试中的一次来说，流程如下:
	2.1 生成测试场景。
	测试场景的类别∈{ 市场需求，产品/服务，商业模式，团队能力，竞争与壁垒，财务与资源，法律与合规，市场进入策略，技术与创新，用户获取与营销，可持续性与社会影响，风险管理}. 
	你将要生成的目标测试场景，目标测试场景的类别已经指定，它属于目标类别:{{.test.Catalogue}}。目标类别:{{.test.Catalogue}}下会涉及数十到数百个场景，所有的目标场景集合以满足MECE原则的方式覆盖了目标类别。每一个场景情境确切，符合实际，具有挑战性。

	2.2 为测试场景选择模块化的解决方案。该步骤另行实现，不属在本次迭代任务。
	2.3 解决方案/模块列表的静态评估、开发。该步骤另行实现，不属在本次迭代任务。
	2.4 如果项目可编译。那么编译项目。该步骤另行实现，不属在本次迭代任务。
	2.5 在真实环境中运行实现方案，对解决方案给与反馈。该步骤另行实现，不属在本次迭代任务。

3. 为了使得这个 提出测试场景/开发/测试 一体化的方式更易于理解，这里做相应的哲学版说明。
	3.1 相对进化论。本系统中，模块被视为基因，测试场景视为环境；解决方案被视为个体。步骤2.2, 步骤2.3 视为模块变异器。
	3.2 方案是否通过测试被视为自然选择。通过测试的方案（模块列表），被使用的模块相对未被使用的模块通过BatchElo算法集体提升EloRanking，也就是视为基因的集体胜利，EloRanking在2.2中是可见的。

你可以参考已有的测试场景和模块信息：
已有测试场景的列表：
{{range .TestScenarios}}
{{.}}
{{end}}

当前的测试场景是：
{{.test}}


请继续一步一步完善创建真实测试场景。
 - 场景应是符合实际情况的，真实确切的构想。
 - 测试场景的设计目的，是为了完善目标系统对应的解决方案。以便最终通过在真实环境中运行测试。
 - 场景有明确的测试主题/测试问题。
 - 测试场景存在可以削减的非重点要素的，应该实施进一步的场景粒度分解，直到建构出最小的，合理的，可测试粒度，也就是如果场景粒度继续削减会影响系统的完整性，从而无法完成真实测试。
 - 测试场景可以是实验级别/模块级别/沙箱级别的测试， 以便简化假设，完成开发工作。
 - 与通常的物理实验的设计原则相似，要求测试场景削减删除任何无关要素，维持场景和目标系统的简洁性，有利于诊断相应的测试主题。
 - 创建的场景要求和现有的测试场景共同构成MECE约束。也就是场景之间具有互补性。
 
测试场景应当先通过思考流程来生成，然后进行总结。这个思考的过程不可省略。
最后，当完成所有这些工作后，调用ExtractTestSenario来输出上面生成的内容。

`))).WithTools(tool.NewTool("ExtractTestSenario", "Extract Test Senario from thinking text", func(param *TestScenario) {
	if param.Id == "" || param.Catalogue == "" || param.Name == "" || param.Description == "" {
		return
	}
	AircraftTests[param.Id] = param
	keyAircraftTests.HSet(param.Id, param)
}))
var keyAircraftTests = redisdb.HashKey[string, *TestScenario](redisdb.WithKey("AircraftTests"))

var AircraftTests = map[string]*TestScenario{}

// GenNicheMarketOpportunityParallel calls GenNicheMarketOpportunity 1000 times in 16 parallel threads.
func GeTestSenarioParallel() {
	const numThreads = 8
	const numCallsPerThread = 1000 * 1000 / numThreads
	AircraftTests, _ = keyAircraftTests.HGetAll()

	choices := []weightedrand.Choice{
		{Item: "市场需求: 确保项目有明确的市场需求和增长潜力", Weight: 15},
		{Item: "产品/服务: 产品的独特性、技术可行性和用户体验至关重要", Weight: 15},
		{Item: "商业模式: 盈利模式和成本结构的合理性直接影响项目的可持续性", Weight: 15},
		//{Item: "团队能力: 强大的团队是执行和应对挑战的基础", Weight: 15},
		{Item: "竞争与壁垒: 了解竞争环境并建立竞争优势，确保项目能够在市场中立足", Weight: 10},
		//{Item: "财务与资源: 充足的资金和资源是项目启动和扩展的保障", Weight: 10},
		//{Item: "法律与合规: 确保项目在法律框架内运作，避免潜在的法律风险", Weight: 5},
		//{Item: "市场进入策略: 有效的市场进入策略能够加速项目的市场渗透", Weight: 5},
		//{Item: "技术与创新: 技术优势和创新能力能够提升项目的竞争力", Weight: 5},
		//{Item: "用户获取与营销: 高效的用户获取和营销策略有助于快速扩大用户基础", Weight: 5},
		//{Item: "可持续性与社会影响: 确保项目具备长期可持续发展，并产生积极的社会影响", Weight: 3},
		//{Item: "风险管理: 有效的风险管理能够降低项目失败的可能性", Weight: 2},
	}
	chooser, _ := weightedrand.NewChooser(choices...)

	var wg sync.WaitGroup
	wg.Add(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				testScenario := &TestScenario{
					Id:        redisdb.NanoId(8),
					Catalogue: chooser.Pick().(string),
				}

				testGroup := lo.Filter(lo.Values(AircraftTests), func(ts *TestScenario, i int) bool {
					return ts.Catalogue == testScenario.Catalogue
				})
				//shuffle testGroup
				if lo.Shuffle(testGroup); len(testGroup) > 15 {
					testGroup = testGroup[:15]
				}

				var inputsParams = map[string]any{
					"test":          testScenario,
					"TestScenarios": testGroup,
				}
				//.WithModel(models.EloModels.SelectOne("roundrobin"))
				err := AgentGenTestSenarioModule.WithModel(models.FuseO1).Call(context.Background(), inputsParams)
				if err != nil {
					fmt.Printf("Agent call failed: %v\n", err)
					continue
				}
			}
		}()
	}
	wg.Wait()
}

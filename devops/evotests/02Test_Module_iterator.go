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

type ModuleFile struct {
	ModuleId    string
	Name        string   `description:"Name of the File"`
	Dependance  []string `description:"Dependance of the File, refer to Name of the Files"`
	DesignIdeas []string
	Content     string `description:"Content of the module file"`
}

var keyAircraftModuleFiles = redisdb.NewHashKey[string, *ModuleFile](redisdb.Opt.Key("AircraftModules"))

var AgentModuleEvalBuild = agent.NewAgent(template.Must(template.New("GenTestSolution").Parse(`
You are TestBuilder. 你的存在是为了通过创建商业场景来建构系统，并最终在真实场景中通过大量的测试。

总体商业目标：实现AGI时代最受欢迎的无人机载具平台。作为各种机器人投递平台，物流投送平台。

本系统工作原理：
1. 本系统通过 数百万次迭代进行 提出测试场景/开发/测试，来实现AGI时代最受欢迎的无人机载具平台这个总体商业目标。
2. 对数百万次迭代开发/测试中的一次来说，流程如下:
	2.1 生成测试场景。(已完成此步骤)
	2.2 为测试场景选择一系列的模块化以构造完备的解决方案如果无法形成完备的解决方案，则为解决方案补充一个Missing Module。(已完成此步骤)
	2.3 解决方案中的 模块的静态 开发。（这是本次迭代任务！）
	2.4 编译模块。编译项目。（后续任务）
	2.5 在真实环境中运行实现方案，对解决方案给与反馈。（后续任务）

3. 为了使得 提出测试场景/开发/测试 一体化的方式更易于理解，这里做相应的哲学版说明。
	3.1 相对进化论。本系统中，模块被视为基因，测试场景视为环境；解决方案被视为个体。步骤2.2, 步骤2.3 视为模块变异器。
	3.2 方案是否通过测试被视为自然选择。通过测试的方案（模块列表），被使用的模块相对未被使用的模块通过BatchElo算法集体提升EloRanking，也就是视为基因的集体胜利，EloRanking在2.2中是可见的。

当前的测试场景是：
{{.test}}

现有的模块信息：
{{.Module}}

现有的模块源码 / 文件信息：
{{.ModuleFiles}}

现有的解决方案信息：
{{.SolutionModules}}

现有的模块依赖性：
{{.ModuleDependencies}}


对当前/本次 迭代任务: "2.3 解决方案/模块列表的静态评估、开发。" 的进一步说明:
	- 你的目标是改进模块的实现文档。
	- 涉及软件的开发语言采用golang，js, react, daisyui, tailwindcss等。
	- 如果文件缺失。请大量提出文件创建请求。
	- 创建的文件内容，要确保模块逻辑通往正确的方向。因为本系统采用进化的方式开发，模块作为解决方案拥有正确的演化适应性方向至关重要。
	- 创建的文件内容不必苛求一步到位，因为本系统采用进化的方式开发，在方向优先被照顾的前提下，尽可能做到每次都有改进。
	- 请接收模块围绕测试意图的实现应该采用的大胆的创新，想法，不必参考任何传统的案例，不必借鉴经典的视线思路。但应该符合设计极简约束，符合可行约束。
	- 在完成所有这些工作后，调用 ExtractModuleFiles 来输出上面生成的缺失模块内容。
`))).WithTools(tool.NewTool("ExtractModuleDataStructure", "Extract Module Data Structure from Text", func(module *ModuleFile) {
	if module.Name == "" || module.ModuleId == "" || len(module.DesignIdeas) == 0 {
		return
	}
	keyAircraftModuleFiles.ConcatKey(module.ModuleId).HSet(module.ModuleId, module)
}))

func GeSolutionModule() {
	const numThreads = 1
	const numCallsPerThread = 1000 * 1000 / numThreads
	AircraftTests, _ = keyAircraftTests.HGetAll()

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
				testScenario := &TestScenario{
					Id:        redisdb.NanoId(8),
					Catalogue: chooser.Pick().(string),
				}

				// testGroup := lo.Filter(lo.Values(AircraftTests), func(ts *TestScenario, i int) bool {
				// 	return ts.Catalogue == testScenario.Catalogue
				// })
				// //shuffle testGroup
				// if lo.Shuffle(testGroup); len(testGroup) > 15 {
				// 	testGroup = testGroup[:15]
				// }

				var inputsParams = map[string]any{
					"test": testScenario,
					"Solutions": &Solutions{
						SenarioId:  testScenario.Id,
						SolutionId: redisdb.NanoId(8),
						ModelIds:   []string{},
					},
					"Modules": lo.MapToSlice(ModelMap, func(key string, m *Model) *Model {
						return m
					}),
				}
				//.WithModel(models.EloModels.SelectOne("roundrobin"))
				err := AgentModuleEvalBuild.WithModels(models.FuseO1).Call(context.Background(), inputsParams)
				if err != nil {
					fmt.Printf("Agent call failed: %v\n", err)
					continue
				}
			}
		}()
	}
	wg.Wait()
}

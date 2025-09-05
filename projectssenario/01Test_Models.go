package projectssenario

import (
	"context"
	"fmt"
	"sync"
	"text/template"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
	"golang.org/x/exp/rand"
)

type Solutions struct {
	SenarioId       string
	SolutionId      string `json:"-"`
	InnovativeIdeas []string
	ProblemToSolve  []string
	ModelIds        []string `description:"Solution == Models, use Ids to reference Models"`
	TestFeedbacks   []string `json:"-"`
}

var keyAircraftSolution = redisdb.NewHashKey[string, *Solutions](redisdb.Opt.Key("AircraftSolution"))

// 模块结构体（简化版）
type Model struct {
	Name    string `description:"model name, non-empty"`
	ModelId string `description:"model id, non-empty"`

	Milestone float64 // 1: file/code constructed, 2:file/code tested, 3:hardware constructed, 4:hardware tested, 5:Income generated

	ProblemToSolve []string `description:"non-empty"`
	DesignIdeas    []string `description:"non-empty"`
	Dependencies   []string

	DevelopFeedbacks []string `description:"Feedbacks from coding, documentation"`
	CompileFeedbacks []string `description:"Feedbacks from compiling"`
}

var keyAircraftModels = redisdb.NewHashKey[string, *Model](redisdb.Opt.Key("AircraftModels"))
var ModelMap = make(map[string]*Model)

func init() {
	ModelMap, _ = keyAircraftModels.HGetAll()
}

var AgentGenSolutionModel = agent.NewAgent().WithTemplate(template.Must(template.New("GenTestSolution").Parse(`
总体商业目标：实现AGI时代最受欢迎的无人机载具平台。作为各种机器人投递平台，物流投送平台。

# 本系统(TestBuilder)实现总体目标的工作原理：
	本系统通过数百万次迭代以下流程来实现总体商业目标。
	1 生成测试场景。
	2 为测试场景选择若干模型(也就是模块，但模型强调创造性思考和设计)的组合，来以构造完备的解决方案；如果无法形成完备的解决方案，则为解决方案补充一个Missing Model。
	3 解决方案/模型的静态评估、开发。
	4 如果项目可编译。那么编译项目。
	5 在真实环境中运行实现方案，对解决方案给与反馈。

# 当前迭代的任务说明:
	本次迭代任务是 "2 为测试场景选择一系列的模块化以构造完备的解决方案；如果无法形成完备的解决方案，则为解决方案补充一个Missing Model。"
	当前迭代的任务进一步说明:
	任务1: 为测试场景，从现有的模块信息当中选择一系列的模块以构造完备的解决方案. 具体步骤：
		1.1 讨论并以模块Id列表的形式给出解决方案
			模块列表中的模块在功能上是构成对 测试场景的解决方案需要的功能 的MECE分解。挑选出的模块数量需要尽可能少，以便符合高内聚，低耦合原则，降低需要考虑的系统耦合要素。
		1.2 讨论提出的解决方案是否足以完成对 测试场景的解决方案需要的功能 的MECE分解; 
		1.3 如果方案无法完成测试场景的需求。那么判定提出解决方案的前置条件，功能完备的模块集合还不存在，此时请转而继续执行任务2
		1.4 如果解决方案足以完成对 测试场景的解决方案需要的功能 的MECE分解，那么生成模块列表.此时使用ToolCall: ExtractSolutionFromThinkingText 来输出上面生成的解决方案。并且结束本次迭代任务。	
	任务2: 完成本次迭代目标 - 生成，并且只生成一个解决方案中缺失的模型/模块. 具体步骤：		
		2.1 识别并确认出任务1中缺失的第一性原理层面的真实需求。将此时的需求定义为输出模型/模块. 接下来，我们将开始反复分析输出模型/模块的真实需求。
		2.2 反复尝试绕过提出的真实需求，也就是尝试证伪输出模型/模块的真实需求，如果需求被证伪，那么进一步提出真正的需求，如此反复证伪，直到需求无法证伪。
		2.3 对2.2中的真实需求进行MECE分解。真实需求的分解采用大胆的创新，同时符合设计极简，容易实施约束。并选择分解后的主要模块成为新的输出模块
		2.4 重复动作2.2，除非当前输出模型/模块 不适宜被过度再分解。
		2.5 在完成2.2后，使用ToolCall: ExtractModelFromThinkingText 来保存输出完整的模型/模块内容, Model的 ModelId 值为{{.ModelId}}。并且结束本次迭代任务。
		任务2举例: 比方说，你需要引用一个固定机翼模块。但又不存在固定机翼模块。那你需要引用机翼模块，并用MECE分解为机翼支架，机翼马达，机翼表面材料... 等模块。对还不存在的机翼表面材料模块，假设我们继续MECE分解为, 固定翼表面，固定翼变长... 等模块。 最后讨论发现固定翼表面模块并不适合MECE分解，所以我们继续提出一个固定翼表面模块，它像抽烟机管道一样的可伸缩结构，通过注入空气来控制长度，通过褶皱来避免展向气流。注意，你只需要提出一个最细粒度的模块，但这个模块应当完成一个意图实现。

当前的测试场景是：
{{.test}}

现有的模块信息：
{{range .Models}}
{{.}}
{{end}}

现有的解决方案信息：
{{.Solutions}}

这是Model的格式说明：
{{.Model}}

现在请以这个标签开始你的回复：<think>

`))).WithTools(tool.NewTool("ExtractModelFromThinkingText", "Extract Model from thinking Text", func(model *Model) {
	if model.Name == "" || model.ModelId == "" || len(model.DesignIdeas) == 0 {
		return
	}
	ModelMap[model.ModelId] = model
	keyAircraftModels.HSet(model.ModelId, model)

})).WithTools(tool.NewTool("ExtractSolutionFromThinkingText", "Extract Solution from thinking Text", func(param *Solutions) {
	if param.SenarioId == "" || len(param.ModelIds) == 0 {
		return
	}
	keyAircraftSolution.HSet(param.SenarioId, param)
}))

func GenModelParallel() {
	const numThreads = 1
	const numCallsPerThread = 1000 * 1000 / numThreads
	AircraftTests, _ = keyAircraftTests.HGetAll()

	var wg sync.WaitGroup
	wg.Add(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				//testSenario : random one value from AircraftTests
				testSenario := lo.Values(AircraftTests)[rand.Intn(len(AircraftTests))]
				var inputsParams = map[string]any{
					"test": testSenario,
					"Solutions": &Solutions{
						SenarioId:  testSenario.Id,
						SolutionId: redisdb.NanoId(8),
						ModelIds:   []string{},
					},
					"Models": lo.MapToSlice(ModelMap, func(key string, m *Model) *Model {
						return m
					}),
					"ModelId": redisdb.NanoId(8),
				}
				err := AgentGenSolutionModel.WithModels(models.FuseO1).Call(context.Background(), inputsParams)
				if err == nil {
					continue
				}
				fmt.Printf("Agent call failed: %v\n", err)
			}
		}()
	}
	wg.Wait()
}

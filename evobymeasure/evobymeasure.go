package evobymeasure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

type FileRefine struct {
	Filename            string `description:"string, Ascii filename of current node。using bullet name to denodes node's modualized intention. extension name such as .md ... is needed"`
	BulletDescription   string `description:"string, Required when create. BulletDescription 是文件内容的摘要。描述和文件的模块化的意图。规定实现的细节."`
	Delete              bool   `msgpack:"-" description:"bool, Whether this node is deleted. If true, the node will be removed from the system."`
	FileContent         string `description:"string, The contents of the file stored on disk"`
	ModificationGroupID string `description:"string, The ID of the modification. FileRefines shares the given ModificationGroupID."`

	ProductGoal string       `msgpack:"-" description:"-"`
	ThisAgent   *agent.Agent `msgpack:"-" description:"-"`
}
type FileRefineMeasurements struct {
	ModificationGroupID string `description:"string, The ID of the modification. FileRefines shares the given ModificationGroupID."`
	// === 上下文：目标与权重 (Context: Goal & Weights) ===
	// 权重值由更高层的策略控制器在创建此任务时填充，决定了本次评估的侧重点
	Weight_StructuralQuality float64 `description:"number, [0-1], 结构质量指标簇的总体权重"`
	Weight_Maintainability   float64 `description:"number, [0-1], 可维护性指标簇的总体权重"`
	Weight_Performance       float64 `description:"number, [0-1], 性能可靠性指标簇的总体权重"`
	Weight_UserValue         float64 `description:"number, [0-1], 用户价值指标簇的总体权重"`

	// === 护栏：绝对值门禁 (Guardrails: Absolute Gates) ===
	// 这些值是评估前的硬性要求，用于一票否决
	Guardrail_Absolute_BuildMustSucceed bool `description:"bool,  预期是否能够成功编译/构建. 这是评分函数的第一道门禁. 如果为false 则修改方案被一票否决."`
	//Guardrail_Absolute_MinTestCoverage       float64 `description:"number, [0-1], 要求的最低单元测试覆盖率. 例如 0.75"`
	//Guardrail_Absolute_ChangeFailureRate_Max float64 `description:"number, [0-1], 预估的变更失败率不能超过此阈值. 例如 0.05"`

	// === 度量：结构质量指标 (Metrics: Structural Quality) ===
	// Delta值表示改进幅度，正为优，负为劣
	Metric_Struct_CognitiveComplexityDelta  float64 `description:"number, [-10,10], 认知复杂度变化. 计算: 10 * (Before - After) / Before. Before为0时特殊处理."`
	Metric_Struct_CyclomaticComplexityDelta float64 `description:"number, [-10,10], 圈复杂度变化. 计算: 10 * (Before - After) / Before. Before为0时特殊处理."`
	Metric_Struct_CodeCouplingDelta         float64 `description:"number, [-10,10], 模块耦合度变化(越低越好). 计算: 10 * (Before - After) / Before."`
	Metric_Struct_CodeCohesionDelta         float64 `description:"number, [-10,10], 模块内聚度变化(越高越好). 计算: 10 * (After - Before) / Before."`
	Metric_Struct_CodeConcisenessDelta      float64 `description:"number, [-10,10], 代码行数(LOC)变化(功能等价下, 越少越好). 计算: 10 * (LOC_Before - LOC_After) / LOC_Before."`

	// === 度量：可维护性指标 (Metrics: Maintainability) ===
	Metric_Maint_CoreLogicDocumentationDelta float64 `description:"number, [-10,10], 核心逻辑文档/注释覆盖率变化(越高越好). 计算: 10 * (Rate_After - Rate_Before) / Rate_Before."`
	Metric_Maint_TestabilityDelta            float64 `description:"number, [-10,10], 可测试性变化(通常与单元测试覆盖率正相关). 计算: 10 * (Coverage_After - Coverage_Before) / Coverage_Before."`
	Metric_Maint_DuplicationRateDelta        float64 `description:"number, [-10,10], 重复代码率变化(越低越好). 计算: 10 * (Rate_Before - Rate_After) / Rate_Before."`

	// === 度量：性能与可靠性指标 (Metrics: Performance & Reliability) ===
	// 这些通常需要通过模拟、沙箱或灰度环境预估
	Metric_Perf_CorrectnessDelta         float64 `description:"number, [-10,10], 正确性分数变化(如静态分析错误数减少). 计算: 10 * (Errors_Before - Errors_After) / Errors_Before."`
	Metric_Perf_ChangeFailureRateDelta   float64 `description:"number, [-10,10], 预估变更失败率变化(越低越好). 计算: 10 * (Rate_Before - Rate_After) / Rate_Before."`
	Metric_Perf_EstimatedLatencyDelta    float64 `description:"number, [-10,10], 预估核心路径P99延迟变化(越低越好). 计算: 10 * (Latency_Before - Latency_After) / Latency_Before."`
	Metric_Perf_EstimatedThroughputDelta float64 `description:"number, [-10,10], 预估系统吞吐量变化(越高越好). 计算: 10 * (TPS_After - TPS_Before) / TPS_Before."`
	Metric_Perf_ProductionErrorRateDelta float64 `description:"number, [-10,10], 预估线上错误率变化(越低越好). 计算: 10 * (Rate_Before - Rate_After) / Rate_Before."`

	// === 度量：用户与业务价值指标 (Metrics: User & Business Value) ===
	// 这些是最难量化的，早期可以依赖启发式规则或人工打分
	Metric_User_SenarioCoverageDelta   float64 `description:"number, [-50,50], 业务场景覆盖度变化. 可能基于需求/测试用例关联分析. 计算: 10 * (Coverage_After - Coverage_Before) / Coverage_Before."`
	Metric_User_SatisfactionProxyDelta float64 `description:"number, [-50,50], 用户满意度代理指标变化. 例如，简化了一个已知复杂流程，可给一个正值. 计算: 人工或启发式规则赋值."`
}

func (a *FileRefine) FileSize() string {
	if a == nil || a.FileContent == "" {
		return " (size: 0 B)"
	}
	size := len(a.FileContent)
	if size > 1024*1024 {
		return fmt.Sprintf(" (size: %.2f MB)", float64(size)/1024/1024)
	} else if size > 1024 {
		return fmt.Sprintf(" (size: %.2f KB)", float64(size/1024))
	} else {
		return fmt.Sprintf(" (size: %d B)", size)
	}
}

type FileRefineList []*FileRefine

func (a FileRefineList) Uniq() FileRefineList { return lo.Uniq(a) }
func (a FileRefineList) FullView() string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Filename, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		//description := "\nBulletDescription: " + v.BulletDescription

		s := fmt.Sprint(indence, "\n Pathname", v.Filename, v.FileSize(), "\nFileContent: ", v.FileContent, "\n\n\n\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a FileRefineList) View(FullViewPaths ...string) string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Filename, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		fileContent := lo.Ternary(slices.Contains(FullViewPaths, v.Filename), "\nFileContent: "+v.FileContent, "")

		s := fmt.Sprint(indence, "\n Pathname", v.Filename, v.FileSize(), "\nBulletDescription: ", v.BulletDescription, fileContent, "\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a FileRefineList) PathnameSorted() FileRefineList {
	slices.SortFunc(a, func(a, b *FileRefine) int {
		return strings.Compare(a.Filename, b.Filename)
	})
	return a
}

func LoadExtraPathToMapFileRefineMap(RootPath, ExtraPath string, solution map[string]*FileRefine) {
	extraPath := filepath.Join(RootPath, ExtraPath)
	ExtraPathFiles, _ := os.ReadDir(extraPath)
	for _, file := range ExtraPathFiles {
		fn := filepath.Join(extraPath, file.Name())
		filename := ExtraPath + "/" + file.Name()
		//hidden file skip
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		FileContent := utils.TextFromFile(fn)
		if strings.Contains(FileContent, "\x00") {
			continue // skip file with null character
		}
		filename = strings.TrimPrefix(filename, "./")
		solution[filename] = &FileRefine{
			Filename:    filename,
			FileContent: FileContent,
		}
	}
}

func (node *FileRefine) SaveContentToPath(RootPath string) {
	//save to root path
	filename := filepath.Join(RootPath, node.Filename)
	err := os.WriteFile(filename, []byte(node.FileContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filename, err)
	}
}

// VetoPenalty 是一个巨大的负分，用于惩罚任何违反护栏的改进
const VetoPenalty = -10000.0

// ScoreOfFileRefine calculates a weighted score for the refinement.
// It assumes the FileRefine struct is fully populated with context (weights) and metrics.
func (f *FileRefineMeasurements) ScoreOfFileRefine() float64 {
	// --- Step 1: Guardrail Checks (护栏检查) ---
	// 任何违反基本质量门禁的尝试都将立即被否决.
	if !f.Guardrail_Absolute_BuildMustSucceed {
		return VetoPenalty // 编译/构建失败，直接否决
	}
	// 获取修改后的绝对测试覆盖率 (这里假设可以从某个地方获得，例如一个临时的Metric字段)
	// 为简化，我们这里假设TestabilityDelta的计算基准包含了绝对值信息。
	// 一个更鲁棒的设计可能需要一个`Metric_Absolute_TestCoverage`字段。
	// if After_Test_Coverage < f.Guardrail_Absolute_MinTestCoverage {
	//     return VetoPenalty
	// }
	// 其他护栏检查...

	// --- Step 2: Calculate Weighted Sub-Scores (计算各指标簇的加权子分数) ---

	// 结构质量分
	structuralScore := f.Metric_Struct_CognitiveComplexityDelta +
		f.Metric_Struct_CyclomaticComplexityDelta +
		f.Metric_Struct_CodeCouplingDelta +
		f.Metric_Struct_CodeCohesionDelta +
		f.Metric_Struct_CodeConcisenessDelta

	// 可维护性分
	maintainabilityScore := f.Metric_Maint_CoreLogicDocumentationDelta +
		f.Metric_Maint_TestabilityDelta +
		f.Metric_Maint_DuplicationRateDelta

	// 性能与可靠性分
	performanceScore := f.Metric_Perf_CorrectnessDelta +
		f.Metric_Perf_ChangeFailureRateDelta +
		f.Metric_Perf_EstimatedLatencyDelta +
		f.Metric_Perf_EstimatedThroughputDelta +
		f.Metric_Perf_ProductionErrorRateDelta

	// 用户价值分
	userValueScore := f.Metric_User_SenarioCoverageDelta +
		f.Metric_User_SatisfactionProxyDelta

	// --- Step 3: Final Weighted Sum (最终加权求和) ---
	// 将各簇的分数与上下文中的权重相乘，得到最终总分。
	finalScore := (f.Weight_StructuralQuality * structuralScore) +
		(f.Weight_Maintainability * maintainabilityScore) +
		(f.Weight_Performance * performanceScore) +
		(f.Weight_UserValue * userValueScore)

	// 确保权重总和大致归一化，否则分数会无限膨胀。
	// 调用者有责任提供合理的权重。
	return finalScore
}

// var ThisTopic = "直观感受1-20数字的大小"
var SolutionBaseLearnByChoose = redisdb.NewHashKey[string, *FileRefine](redisdb.Opt.HttpVisit(), redisdb.Opt.Key("ProjectRefine:LearnByChoose"))

// var RootPathLearnByChoose = "/Users/yang/doptime/evolab/web/app/perceptual-understanding-numbers-1-to-20"
// /Users/yang/doptime/evolab/web/app/learn-by-choose/data-mock.ts
var RootPathLearnByChoose = "/Users/yang/doptime/evolab/web/app/learn-by-choose"
var ExtraPathLearnByChoose = "../../components/guesture"

func LoadExtraPathToMap1(solution map[string]*FileRefine) {
	extraPath := filepath.Join(RootPathLearnByChoose, ExtraPathLearnByChoose)
	ExtraPathFiles, _ := os.ReadDir(extraPath)
	for _, file := range ExtraPathFiles {
		fn := filepath.Join(extraPath, file.Name())
		filename := ExtraPathLearnByChoose + "/" + file.Name()
		solution[filename] = &FileRefine{
			Filename:    filename,
			FileContent: utils.TextFromFile(fn),
		}
	}
}

var AgentEvoLearningSolutionLearnByChoose = agent.NewAgent(template.Must(template.New("AgentEvoLearningSolutionLearnByChoose").Parse(`
## 本系统采用迭代方式完成开发工作,以实现或改进方案的实现。
- 简洁可靠: 优先删除无效的引用、消除错误。确保实现方案和代码简洁可靠。
- 以终为始: 努力实现系统的目标意图，并使得项目处于可编译、可运行的状态。
- 粒度适中:，对粒度过大，难以实施的文件，优先对粒度进行拆分，也就是创建多个更细粒度的文件实现，并且删除原有的文件; 反之需要用高内聚低耦合的方式，重构文件内容。

## 当前的系统状态:
这是系统的product goal:
{{.ProductGoal}}


这是当前的解决方案:
{{.Solution}}

这是函数调用需要填写的ModificationGroupID:
{{.ModificationGroupID}}

TODO:总的任务分成两部分，一部分是对方案的改进，另一部分是对改进的评价。两个任务分别通过调用不同的函数来完成。

TODO Step 1: SolutionFileRefine
请在现有的实现。思考并提出一个清晰而明确的小改进，并提交完整的修改代码到本地文件:
- 请先讨论，提出需要修改的主题。注意改进的主题粒度宜小不宜大。如果你有十成的能力，那只要用20%的能力来确保改进的可靠性。
- 主题的品味关键是确保演化方向的正确性。绝不要追求大幅的改进。要追求小幅的调整。对大幅的改进目标，应该转而实现目标中的一个小Milestone。
- 由于当前正在大并发生成修改方案，因此请确保生成的内容在探索方向上具有多样性。
- 确保当前的改进副作用很少。现有的好的思路和实现方式可以被保留。
- 修改后的文件内容必须完整而没有遗漏和错误，不能只提交部分内容。如果是增量修改，请将其转化为全量的文件内容。以避免编译失败。
- 最后通过（多次）调用 FunctionCall:SolutionFileRefine 来提交不同文件的改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Filename,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)删除无效节点

TODO Step 2:  SolutionRefineMeasure：在完成了（一个或多个）调文件内容的修改之后，还需要对本次修改进行评价。
- 请确保评价仅当发生在全部修改完成之后。
- 当前正在大并发生成修改和评价修改，最后保留得分最高的。你必须在看不到其它修改方案的前提下，正确而适当地评价你的修改，以便能正确筛选出包括最可靠的修改方案，确保最终可以成功演化系统。
- 方案评价的得分的计算公式是确定的。你同样需要填写相关的评价参数：
	structuralScore := f.Metric_Struct_CognitiveComplexityDelta +
		f.Metric_Struct_CyclomaticComplexityDelta +
		f.Metric_Struct_CodeCouplingDelta +
		f.Metric_Struct_CodeCohesionDelta +
		f.Metric_Struct_CodeConcisenessDelta

	// 可维护性分
	maintainabilityScore := f.Metric_Maint_CoreLogicDocumentationDelta +
		f.Metric_Maint_TestabilityDelta +
		f.Metric_Maint_DuplicationRateDelta

	// 性能与可靠性分
	performanceScore := f.Metric_Perf_CorrectnessDelta +
		f.Metric_Perf_ChangeFailureRateDelta +
		f.Metric_Perf_EstimatedLatencyDelta +
		f.Metric_Perf_EstimatedThroughputDelta +
		f.Metric_Perf_ProductionErrorRateDelta

	// 用户价值分
	userValueScore := f.Metric_User_SenarioCoverageDelta +
		f.Metric_User_SatisfactionProxyDelta

	// --- Step 3: Final Weighted Sum (最终加权求和) ---
	// 将各簇的分数与上下文中的权重相乘，得到最终总分。
	finalScore := (f.Weight_StructuralQuality * structuralScore) +
		(f.Weight_Maintainability * maintainabilityScore) +
		(f.Weight_Performance * performanceScore) +
		(f.Weight_UserValue * userValueScore)
- 最后，请调用一次SolutionRefineMeasure 来完成对本次修改进行评价。


`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionFileRefine", "create/modify/remove solution file", func(newItem *FileRefine) {
	newItem.BulletDescription = strings.TrimSpace(newItem.BulletDescription)
	if newItem.Filename == "" {
		return
	}
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "Pathname")
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "Path")
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "./")
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "src/")
	newItem.Filename = strings.TrimPrefix(newItem.Filename, "app/")
	JobReceived = append(JobReceived, newItem)
	if _, ok := ModificationGroups[newItem.ModificationGroupID]; !ok {
		ModificationGroups[newItem.ModificationGroupID] = []*FileRefine{}
	}
	ModificationGroups[newItem.ModificationGroupID] = append(ModificationGroups[newItem.ModificationGroupID], newItem)

}), tool.NewTool("SolutionRefineMeasure", "回顾并评价本批次修改. 约束说明：Weight_StructuralQuality + Weight_Maintainability + Weight_Performance + Weight_UserValue == 1.0", func(measurement *FileRefineMeasurements) {
	if measurement.ModificationGroupID == "" {
		return
	}
	ModificationMeasurements[measurement.ModificationGroupID] = measurement
}))
var ModificationGroups = map[string][]*FileRefine{} // ModificationGroupID -> []*FileRefine
var ModificationMeasurements = map[string]*FileRefineMeasurements{}
var JobReceived = []*FileRefine{}

func SaveTheBestModification() {
	maxScor := -10000.0
	var FileRefines []*FileRefine = nil
	for _, modification := range ModificationMeasurements {
		_score := modification.ScoreOfFileRefine()
		if _score > maxScor {
			maxScor = _score
			FileRefines = ModificationGroups[modification.ModificationGroupID]
		}
	}
	for _, fileRefines := range FileRefines {
		if fileRefines.Delete {
			pathname := filepath.Join(RootPathLearnByChoose, fileRefines.Filename)
			SolutionBaseLearnByChoose.HDel(fileRefines.Filename)
			os.Remove(pathname)
			return
		}
		fileRefines.SaveContentToPath(RootPathLearnByChoose)
		SolutionBaseLearnByChoose.HSet(fileRefines.Filename, fileRefines)
	}

}

func MakeAEvo() {
	var keySolution = SolutionBaseLearnByChoose

	const MaxThreads = 8
	MaxThreadsSemaphore := make(chan struct{}, MaxThreads)

	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		//backlogs, _ := KeyBacklog.LRange(0, -1)
		allNodes, _ := keySolution.HGetAll()

		//load nodes from file
		for _, node := range allNodes {
			utils.TextFromFile(filepath.Join(RootPathLearnByChoose, node.Filename), &node.FileContent)
		}
		LoadExtraPathToMapFileRefineMap(RootPathLearnByChoose, ".", allNodes)
		LoadExtraPathToMapFileRefineMap(RootPathLearnByChoose, ExtraPathLearnByChoose, allNodes)

		SolutionSummary := FileRefineList(lo.Values(allNodes))
		time.Sleep(300 * time.Millisecond)

		for i := 0; i < MaxThreads; i++ {
			MaxThreadsSemaphore <- struct{}{} // Acquire a spot in the semaphore
			go func(SolutionSummary string, AllItems map[string]*FileRefine) {
				defer func() { <-MaxThreadsSemaphore }()

				ProductGoalUniLearning := utils.TextFromFile("/Users/yang/learn-by-choose-goserver/learninggame.md")
				//Gemini25Flashlight Gemini25ProAigpt
				err := AgentEvoLearningSolutionLearnByChoose.WithModels(models.Qwen3BThinking30A3b2507). //CopyPromptOnly(). //Qwen3B32Thinking
																Call(context.Background(), map[string]any{
						"ProductGoal":         string(ProductGoalUniLearning) + "\n\n",
						"Solution":            SolutionSummary,
						"ModificationGroupID": utils.ID(time.Now().String(), 5),
					})
				if err != nil {
					fmt.Printf("Agent call failed: %v\n", err)
				}
			}(SolutionSummary.FullView(), allNodes)
		}
		// Wait for all the goroutines to finish)
		for i := 0; i < MaxThreads; i++ {
			MaxThreadsSemaphore <- struct{}{}
		}

		SaveTheBestModification()
	}

}

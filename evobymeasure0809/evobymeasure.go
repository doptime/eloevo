package evobymeasure

import (
	"context"
	"fmt"
	"io/fs"
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
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
)

// Core idea:
// 测度  全局梯度 与 工具链中的短板
// 如果我们同时接受有限理性（归因）和 贝叶斯统计的作为核心的工作方法论。那么我们应该 在结构空间和性能空间内分别进行的结构和性能的双重测度优化。如果两个都是改进的，那就是更优的。特别是，我们知道高维空间的梯度实际上不会有驻点。因而如果为性能建立维度广泛的性能梯度，同时为结构建立维度广泛的性能梯度，那么实际上进化的连续性和终极性能连续可导性是有保证的。
// 这种梯度同样包括了逻辑归因。
// 在系统改进的时候，一个关键的问题还是工具的性能的受限性。但表面上工具链的短板影响极大，其实问题也不大。我们应该改变的增量在系统短板之下工作。这根本说来重点不在于每次改善的全局有效性，重点在于提高对产出的全局性能度量的准确性； 确保好的改良沿着正确的测度梯度，在后续的测度分析中k可靠地被保留。
// 校验审核相比推理落实是容易的，但最强大的模型反而应当配置在这个环节，这可以使得系统的全局定向改善的杠杆最大化。

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
	//Metric_Maint_TestabilityDelta            float64 `description:"number, [-10,10], 可测试性变化(通常与单元测试覆盖率正相关). 计算: 10 * (Coverage_After - Coverage_Before) / Coverage_Before."`
	//Metric_Maint_DuplicationRateDelta        float64 `description:"number, [-10,10], 重复代码率变化(越低越好). 计算: 10 * (Rate_Before - Rate_After) / Rate_Before."`

	// === 度量：性能与可靠性指标 (Metrics: Performance & Reliability) ===
	// 这些通常需要通过模拟、沙箱或灰度环境预估
	Metric_Perf_CorrectnessDelta       float64 `description:"number, [-10,10], 正确性分数变化(如静态分析错误数减少). 计算: 10 * (Errors_Before - Errors_After) / Errors_Before."`
	Metric_Perf_ChangeFailureRateDelta float64 `description:"number, [-10,10], 预估变更失败率变化(越低越好). 计算: 10 * (Rate_Before - Rate_After) / Rate_Before."`
	// Metric_Perf_EstimatedLatencyDelta    float64 `description:"number, [-10,10], 预估核心路径P99延迟变化(越低越好). 计算: 10 * (Latency_Before - Latency_After) / Latency_Before."`
	//Metric_Perf_EstimatedThroughputDelta float64 `description:"number, [-10,10], 预估系统吞吐量变化(越高越好). 计算: 10 * (TPS_After - TPS_Before) / TPS_Before."`
	//Metric_Perf_ProductionErrorRateDelta float64 `description:"number, [-10,10], 预估线上错误率变化(越低越好). 计算: 10 * (Rate_Before - Rate_After) / Rate_Before."`

	// === 度量：用户与业务价值指标 (Metrics: User & Business Value) ===
	// 这些是最难量化的，早期可以依赖启发式规则或人工打分
	Metric_User_SenarioCoverageDelta   float64 `description:"number, [-20,20], 业务场景覆盖度变化. 可能基于需求/测试用例关联分析. 计算: 10 * (Coverage_After - Coverage_Before) / Coverage_Before."`
	Metric_User_SatisfactionProxyDelta float64 `description:"number, [-50,50], 用户满意度代理指标变化. 需要特别重视，强化权重配分。 例如，简化了一个已知复杂流程，可给一个正值. 计算: 人工或启发式规则赋值."`
	Metric_User_LongTermValueDelta     float64 `description:"number, [-50,50], 用户长期价值代理指标变化. 需要特别重视，强化权重配分。  例如，有效的复习策略; 可以给用于带来长期的价. 计算: 人工或启发式规则赋值."`
}

func (a *FileRefine) FileSize() string {
	return lo.Ternary(a == nil || a.FileContent == "", " (size: 0 B)", fmt.Sprintf(" (size: %s)", humanize.Bytes(uint64(len(a.FileContent)))))
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
	filepath.WalkDir(filepath.Join(RootPath, ExtraPath), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return filepath.SkipDir
		}
		content := lo.Ternary(err != nil || strings.HasPrefix(d.Name(), "."), "\x00", utils.TextFromFile(path))
		if strings.Contains(content, "\x00") {
			return nil
		}
		filename := strings.TrimPrefix(strings.ReplaceAll(path, RootPath+string(os.PathSeparator), ""), "./")
		solution[filename] = &FileRefine{Filename: filename, FileContent: content}
		return nil
	})
	// ExtraPathFiles, _ := os.ReadDir(extraPath)
	// for _, file := range ExtraPathFiles {
	// 	fn := filepath.Join(extraPath, file.Name())
	// 	filename := ExtraPath + "/" + file.Name()
	// 	//hidden file skip
	// 	content := lo.Ternary(strings.HasPrefix(file.Name(), ".") || d.IsDir())
	// 	if strings.HasPrefix(file.Name(), ".") {
	// 		continue
	// 	}
	// 	FileContent := utils.TextFromFile(fn)
	// 	if strings.Contains(FileContent, "\x00") {
	// 		continue // skip file with null character
	// 	}
	// 	filename = strings.TrimPrefix(filename, "./")
	// 	solution[filename] = &FileRefine{
	// 		Filename:    filename,
	// 		FileContent: FileContent,
	// 	}
	// }
}

func (node *FileRefine) SaveContentToPath(RootPath string) {
	//save to root path
	filename := filepath.Join(RootPath, node.Filename)
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return
	}
	tmp := filename + ".tmp-" + time.Now().String()
	if err := os.WriteFile(tmp, []byte(node.FileContent), 0o644); err != nil {
		return
	}
	if err := os.Rename(tmp, filename); err != nil {
		return
	}
}

func (m *FileRefineMeasurements) ScoreOfFileRefine() float64 {
	// 软性护栏检查
	var guardrailPenalty float64 = lo.Ternary(m.Guardrail_Absolute_BuildMustSucceed, 0.0, -50.0)

	var ws = [4]float64{m.Weight_StructuralQuality, m.Weight_Maintainability, m.Weight_Performance, m.Weight_UserValue}
	sum := ws[0] + ws[1] + ws[2] + ws[3]
	if sum <= 0 {
		ws = [4]float64{0.25, 0.25, 0.25, 0.25}
	}
	structuralWeight, maintainabilityWeight, performanceWeight, userValueWeight := ws[0]/sum, ws[1]/sum, ws[2]/sum, ws[3]/sum

	structuralScore := m.Metric_Struct_CognitiveComplexityDelta + m.Metric_Struct_CyclomaticComplexityDelta + m.Metric_Struct_CodeCouplingDelta + m.Metric_Struct_CodeCohesionDelta + m.Metric_Struct_CodeConcisenessDelta

	maintainabilityScore := m.Metric_Maint_CoreLogicDocumentationDelta

	performanceScore := m.Metric_Perf_CorrectnessDelta + m.Metric_Perf_ChangeFailureRateDelta

	userValueScore := m.Metric_User_SenarioCoverageDelta + m.Metric_User_SatisfactionProxyDelta + m.Metric_User_LongTermValueDelta

	finalScore := (structuralWeight * structuralScore) + (maintainabilityWeight * maintainabilityScore) + (performanceWeight * performanceScore) + (userValueWeight * userValueScore) + guardrailPenalty
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

func normalizeFilename(name string) string {
	for found := true; found; {
		found = false
		for _, r := range []string{"Pathname", "pathname", "Path", "./", "src/", "app/"} {
			name = strings.TrimPrefix(name, r)
			found = true
		}
	}
	return name
}
func normalizeFileContent(s string) string {
	s = strings.Replace(s, "use client;\n", "'use client';", 1)
	return s
}

var AgentEvoLearningSolutionLearnByChoose = agent.NewAgent(template.Must(template.New("AgentEvoLearningSolutionLearnByChoose").Parse(`
Note: 请函数调用需要遵守<tool_call>...</tool_call>的约定。每一个函数调用都必须包含完整的<tool_call>...</tool_call>标记。
# 系统演化任务描述:

## 系统的目标:
{{.ProductGoal}}


## 当前解决方案:
{{.Solution}}

## 修改标识ModificationGroupID:
{{.ModificationGroupID}}


TODO 1: SolutionFileRefine, 对方案的改进
请在现有的实现。思考并提出一个在能力圈安全边际之内的改进，并提交完整的修改代码到本地文件:
- 请先讨论，提出需要修改的主题。注意改进的主题粒度不宜大。你需要确保改进的可靠性，在主题粒度之内，应该追求实现的高质量。对于大的目标，应该转而实现目标中的一个小粒度的Milestone以确保实现的可靠性。
- 由于当前正在并发生成修改方案，因此请确保生成的内容在探索方向上具有多样性。
- 确保当前的改进副作用很少。现有的好的思路和实现方式可以被保留。
- 修改后的文件内容必须完整而没有遗漏和错误，不能只提交部分内容。如果是增量修改，请将其转化为全量的文件内容。以避免编译失败。
	最后通过（至少一次/多次）调用toolcall: SolutionFileRefine 来提交不同文件的改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Filename,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)删除无效节点。请直接在回复的正文，在思考内容之后按完整的<tool_call>...</tool_call>格式调用函数。

TODO 2:  SolutionRefineMeasure,对改进的评价：对本次修改进行评价。
- 现在，开始对TODO1 中的产出进行思考分析并且评价。你需要深度思考填写其中相关的评价参数。
- 当前正在大并发生成修改和评价修改，最后保留得分最高的。你必须在看不到其它修改方案的前提下，正确而适当地评价你的修改，以便能，确保最终可以成功演化系统。
- 评价意图是确定方案样本的品质特性，以便正确筛选出最可靠的修改方案，来可靠的进化当前系统。
	最后，请调用一次toolcall: SolutionRefineMeasure 来完成对本次修改进行评价。请直接在回复的正文，在思考内容之后用另外一个完整的<tool_call>...</tool_call>格式调用函数。



`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionFileRefine", "create/modify/remove solution file", func(newItem *FileRefine) {
	newItem.BulletDescription = strings.TrimSpace(newItem.BulletDescription)
	if newItem.Filename == "" {
		return
	}
	newItem.Filename = normalizeFilename(newItem.Filename)
	newItem.FileContent = normalizeFileContent(newItem.FileContent)
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

	const MaxThreads = 1
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
				err := AgentEvoLearningSolutionLearnByChoose.WithModels(models.Glm45AirLocal). //CopyPromptOnly(). //Qwen3B32Thinking
														Call(context.Background(), map[string]any{
						"ProductGoal":         string(ProductGoalUniLearning) + "\n\n",
						"Solution":            SolutionSummary,
						"ModificationGroupID": utils.ID("", 5),
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

package evobymeasured

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
	"github.com/doptime/eloevo/config"
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
	Filename          string `description:"string, Ascii filename of current node。using bullet name to denodes node's modualized intention. extension name such as .md ... is needed"`
	BulletDescription string `description:"string, Required when create. BulletDescription 是文件内容的摘要。描述和文件的模块化的意图。规定实现的细节."`
	Delete            bool   `msgpack:"-" description:"bool, Whether this node is deleted. If true, the node will be removed from the system."`
	Disposed          bool   `msgpack:"-" description:"bool, Whether this Refine is disposed. If true, this refine will be ignored."`
	FileContent       string `description:"string, The contents of the file stored on disk"`

	ProductGoal string       `msgpack:"-" description:"-"`
	ThisAgent   *agent.Agent `msgpack:"-" description:"-"`
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
		fileContent := lo.Ternary(slices.Contains(FullViewPaths, v.Filename), "\n <file-content>: "+v.FileContent, "</file-content>\n")

		s := fmt.Sprint(indence, "<file>\n <file-name>", v.Filename, "</file-name>\n<file-size>", v.FileSize(), "</file-size>\n<file-BulletDescription>", v.BulletDescription, "</file-BulletDescription>", fileContent, "\n</file>")
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

func LoadAllEvoProjects(solution map[string]*FileRefine) {
	enabledRealms := func(realm *config.EvoRealm, _ int) bool { return realm.Enable }
	for _, realm := range lo.Filter(config.EvoRealms, enabledRealms) {
		RootPath := realm.RootPath
		if !realm.Enable {
			continue
		}
		filepath.WalkDir(RootPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			name := d.Name()
			if strings.HasPrefix(name, ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				// skip this file only
				return nil
			}
			content := lo.Ternary(d.IsDir(), "", utils.TextFromFile(path))
			if binaryFile := strings.Contains(content, "\x00") || len(content) == 0; binaryFile {
				return nil
			}
			filename := strings.TrimPrefix(strings.ReplaceAll(path, RootPath, "$"+realm.Name), "./")
			solution[filename] = &FileRefine{Filename: filename, FileContent: content}
			return nil
		})
	}
}
func ToLocalFile(path string) string {
	realm, found := lo.Find(config.EvoRealms, func(r *config.EvoRealm) bool {
		return strings.HasPrefix(path, "$"+r.Name) && r.Enable
	})
	return lo.Ternary(found && realm != nil, filepath.Join(realm.RootPath, strings.TrimPrefix(path, "$"+realm.Name)), path)
}

func (node *FileRefine) SaveContentToPath(filename string) {
	//save to root path
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return
	}
	realm, found := lo.Find(config.EvoRealms, func(r *config.EvoRealm) bool {
		return strings.HasPrefix(filename, r.RootPath) && r.Enable
	})
	//backup current version to db
	if found && realm != nil {
		content, err := os.ReadFile(filename)
		if err == nil && len(content) > 0 {
			SolutionBaseLearnByChoose.ConcatKey(realm.Name).HSet(node.Filename+"."+time.Now().String(), node)
		}
	}
	if err := os.WriteFile(filename, []byte(node.FileContent), 0o644); err != nil {
		return
	}
}

var AgentEvoLearningSolutionLearnByChoose = agent.NewAgent(template.Must(template.New("AgentEvoLearningSolutionLearnByChoose").Parse(`
# 系统演化任务描述:

## 系统的目标:
<product goal>
{{.ProductGoal}}
</product goal>


<current solution>
{{.Solution}}
</current solution>

<system evolution Goals>
## 潜在的改进目标：
1. 实现复习的调度算法
2. 现在的游戏是基于选择题目的方式来进行学习的。未来还需要探索3-4种不同的学习形态和相应的学习哲学。以便让不同的学习内容和不同的学习者都能找到合适的学习方式。
3. 需要构建合适的内容生成系统。
4. 客户端需要支持从服务端获取学习内容的能力。

## 改进目标对应的改进点：集成FSRS库到后端服务
目标1: 实现复习的调度算法（基于FSRS或其他间隔重复算法，共9个改进点）

实施方案：在Go服务器中使用github.com/open-spaced-repetition/go-fsrs包，通过import引入库，并在初始化函数中加载默认参数（如稳定性S和难度D）。
改进点：定义复习卡片数据结构

实施方案：创建一个Card结构体，包括字段如ID、知识点、最后复习时间、稳定性S、难度D，并使用Redis Hash存储每个用户的卡片数据。
改进点：实现卡片稳定性更新函数

实施方案：基于FSRS的UpdateStability函数，编写一个方法，输入用户评分（Grade）和当前S，输出新S；处理链完整情况下的5%提升。
改进点：计算下次复习时间

实施方案：使用FSRS的NextReviewTime函数，结合Retrievability Tolerance (R_t = 0.7 + 0.2 / sqrt(N_user + k))，调度复习间隔；k默认为10。
改进点：每日复习内容筛选

实施方案：编写一个查询函数，从Redis中筛选今日到期卡片（基于上次复习时间和间隔），并合并到当前学习主题包中。
改进点：处理用户评分转换

实施方案：实现Grade计算公式（g = 4 * sqrt(α * correctness^2 + β * (1 - timePercentile)^2)），动态调整α和β基于N_user。
改进点：集成时间百分位计算

实施方案：使用几何平均值计算timePercentile，权重基于sqrt(N_problem + k)和sqrt(N_user + k)；存储历史响应时间在Redis List中。
改进点：测试复习调度逻辑

实施方案：编写单元测试，模拟用户回答场景，验证S、D更新和下次复习时间计算的正确性。
改进点：添加日志监控复习过程

实施方案：使用logger记录每次复习的Grade、S变化和调度时间，便于调试和优化算法参数。

目标2: 探索3-4种不同的学习形态和相应的学习哲学（共9个改进点）

改进点：定义学习形态分类框架

实施方案：创建文档或枚举类型，分类4种形态（如选择题、构建式、讨论式、应用式），每个关联一种哲学（如行为主义、建构主义、人本主义、连接主义）。
改进点：探索选择题形态的哲学扩展

实施方案：基于当前游戏，添加行为主义哲学元素，如即时反馈奖励系统；测试用户在不同内容下的适应性。
改进点：设计构建式学习形态

实施方案：引入生成任务，如用户构建知识地图；哲学基于建构主义，生成AI辅助用户迭代内容。
改进点：实现讨论式学习形态原型

实施方案：添加聊天模拟界面，用户与AI角色讨论主题；哲学基于人本主义，强调情感支持和个性化对话。
改进点：开发应用式学习形态

实施方案：创建项目模拟环境，如数学应用到游戏中；哲学基于连接主义，链接外部API（如天气数据）生成真实场景。
改进点：评估不同形态的用户匹配

实施方案：设计问卷或A/B测试，收集用户偏好数据，匹配学习内容类型（如抽象 vs. 实践）到形态。
改进点：集成多形态切换机制

实施方案：在客户端添加模式选择UI，根据用户进度动态推荐形态；后端存储用户偏好。
改进点：文档化每种形态的哲学基础

实施方案：编写Markdown指南，解释每种形态的核心原则和预期学习效果，便于未来扩展。
改进点：试点测试3-4种形态

实施方案：选择小样本用户群，进行为期一周的测试，收集反馈并迭代一种形态的生成逻辑。

目标3: 构建合适的内容生成系统（共8个改进点）

改进点：选择内容生成API

实施方案：集成OpenAI GPT-4 API作为生成后端，设置prompt模板以确保知识点准确性。
改进点：设计生成prompt模板

实施方案：创建模板，包括主题、年龄、知识点权重和干扰项要求；确保每次生成4个知识点。
改进点：实现邻域主题混淆生成

实施方案：基于语义相似度，从邻域主题随机选k=1-3个，生成干扰知识点；使用LLM计算相似度。
改进点：添加生成质量校验

实施方案：后生成校验函数，检查正确性、一致性和多样性；如果失败，重生成。
改进点：支持多语言生成

实施方案：添加语言参数到prompt，初始支持中文和英文；测试生成内容的准确性。
改进点：优化生成成本

实施方案：缓存常见主题的生成结果在Redis中，设置TTL为1天；减少重复调用API。
改进点：集成用户反馈循环

实施方案：允许用户报告生成错误，后端记录并微调prompt模板。
改进点：测试生成多样性

实施方案：生成100个样本，分析知识点分布和干扰项质量，确保覆盖率>90%。

目标4: 客户端需要支持从服务端获取学习内容的能力（共6个改进点）

改进点：定义API端点

实施方案：在Go服务器添加GET /topics/{userId}端点，返回今日学习主题包JSON。
改进点：客户端集成API调用

实施方案：在React中使用fetch或axios调用服务端API，替换mock数据。
改进点：处理认证和用户ID

实施方案：添加JWT认证，客户端存储token；服务端验证用户并返回个性化内容。
改进点：实现内容缓存机制

实施方案：在客户端使用localStorage缓存最近主题，减少网络请求；设置过期时间1小时。
改进点：添加错误处理和重试

实施方案：客户端在API失败时显示提示，并自动重试3次；日志错误到控制台。
改进点：测试端到端数据流

实施方案：模拟服务端响应，验证客户端从加载到显示内容的完整流程。
</system evolution Goals>

<Implementation Steps>
实施步骤:
1. **提出修改方案**：请围绕改进领域提出面向一个改进点的内容修改。你需要确保改进的可靠性。对于大的目标，应该转而实现目标中的一个小粒度的Milestone以确保实现的可靠性。副作用很少。
2. **强化修改方案**：对改进方案进行进一步改进。
	1. 在解决方案中优化目标，基于全局测度对改进方案进行评估，并给出一个综合评分（基于上述领域的得分）。重点评估领域包括：目标都明确、用户价值、结构质量、可维护性、性能与可靠性。
	- **目标都明确**：围绕特定的目标提升现有方案。高质量得实现给定目标，最小化副作用。
	- **用户价值**：强化业务场景覆盖度、用户满意度和长期价值。
	- **结构质量**：应着重在认知复杂度、圈复杂度、模块耦合度、内聚度和代码简洁度方面进行优化。
	- **可维护性**：应关注核心逻辑文档覆盖率的提升，确保代码可理解和可测试。
	- **性能与可靠性**：优化代码的正确性、变更失败率、预估延迟和吞吐量。
	
	2. 针对评估中发现的问题进一步优化，对每个存在重要缺陷的领域给出具体的优化方案，确保在当前能力圈的安全边际内进行调整。

	3. 重新评估每个改进的实施后效果。在评分时，请计算改进的每个方面的相对重要性权重，并综合衡量所有领域的改进幅度：
	- **目标都明确**：围绕特定的目标提升现有方案。高质量得实现给定目标，最小化副作用。
	- **用户价值**：包括业务场景覆盖度、用户满意度、长期价值
	- **结构质量**：包括认知复杂度、圈复杂度、耦合度、内聚度、代码简洁度
	- **可维护性**：包括文档覆盖率
	- **性能与可靠性**：包括正确性、变更失败率、预估延迟、吞吐量

	4. 合并生成新的优化方案。

	5. 对每个改进进行必要的验证。确保改进后的文件内容没有遗漏或错误，且必须是完整的。如果当前的改进方向存在明显问题，或不具备显著价值，则直接结束，不要调用SolutionFileRefine。
</Implementation Steps>

<Commit Changes>
## 提交改进方案
最后通过 N∈[0,N] 次独立的调用toolcall: SolutionFileRefine 来提交对现有方案的改进。改进形式包括: 1)创建新节点; 2)修改条目:指定Filename,并修改字段(可忽略不修改字段若，若修改的字段需确保完整性); 3)删除无效节点。请确保修改后的文件内容必须完整而没有遗漏和错误，不能只提交部分内容。如果是增量修改，请将其转化为全量的文件内容。以避免编译失败。
</Commit Changes>
`))).WithToolCallMutextRun().WithTools(tool.NewTool("SolutionFileRefine", "create/modify/remove solution file", func(newItem *FileRefine) {
	newItem.BulletDescription = strings.TrimSpace(newItem.BulletDescription)
	if newItem.Filename == "" {
		return
	}
	newItem.Filename = utils.NormalizeFilename(newItem.Filename)
	newItem.FileContent = utils.NormalizeFileContent(newItem.FileContent)
	Modifications[newItem.Filename] = newItem

}))
var Modifications = map[string]*FileRefine{}

var SolutionBaseLearnByChoose = redisdb.NewHashKey[string, *FileRefine](redisdb.Opt.HttpVisit(), redisdb.Opt.Key("ProjectRefine"))

func SaveTheBestModification() {
	for _, fileRefines := range Modifications {
		if fileRefines.Disposed {
			continue
		}
		localFile := ToLocalFile(fileRefines.Filename)
		if fileRefines.Delete {
			os.Remove(localFile)
			continue
		}
		fileRefines.SaveContentToPath(localFile)
	}

}

func MakeAEvo() {
	var keySolution = SolutionBaseLearnByChoose

	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		//backlogs, _ := KeyBacklog.LRange(0, -1)
		allNodes, _ := keySolution.HGetAll()
		Modifications = map[string]*FileRefine{}
		LoadAllEvoProjects(allNodes)

		SolutionSummary := FileRefineList(lo.Values(allNodes))
		time.Sleep(300 * time.Millisecond)
		func(SolutionSummary string, AllItems map[string]*FileRefine) error {
			ProductGoalUniLearning := utils.TextFromFile("/Users/yang/learn-by-choose-goserver/learninggame.md")
			//Gemini25Flashlight Gemini25ProAigpt
			err := AgentEvoLearningSolutionLearnByChoose.WithModels(models.Glm45AirLocal).WithMsgDeClipboard(). //CopyPromptOnly(). //Qwen3B32Thinking
																Call(context.Background(), map[string]any{
					"ProductGoal": string(ProductGoalUniLearning) + "\n\n",
					"Solution":    SolutionSummary,
				})
			if err != nil {
				fmt.Printf("Agent call failed: %v\n", err)
			}
			return err
		}(SolutionSummary.FullView(), allNodes)

		SaveTheBestModification()
	}

}

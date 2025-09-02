package evobymeasured

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/config"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

var KeyGitCommits = redisdb.NewHashKey[string, gitdiff.File](redisdb.Opt.HttpVisit(), redisdb.Opt.Key("GitCommits"))

func LoadAllEvoProjects(KeepFileNames ...[]string) string {
	var allFileInfo strings.Builder
	fileKeepMap := map[string]bool{}

	for _, fn := range KeepFileNames {
		for _, f := range fn {
			fileKeepMap[f] = true
		}
	}

	for _, realm := range lo.Filter(config.EvoRealms, func(realm *config.EvoRealm, _ int) bool { return realm.Enable }) {
		realm.WalkDir(func(path, relativePath string, info os.FileInfo) (e error) {
			fmt.Printf("Processing file: %s\n", path)
			if len(KeepFileNames) > 0 {
				if _, ok := fileKeepMap[relativePath]; !ok {
					return nil
				}
			}

			// Read the file content
			content, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Error reading file %q: %v\n", path, err)
				return err
			}
			if binaryFile := strings.Contains(string(content), "\x00") || len(content) == 0; binaryFile {
				return nil
			}
			fileSz := "\n<file-size>" + humanize.Bytes(uint64(len(content))) + "</file-size>"
			fileContent := "\n<file-content>\n" + utils.LineNumberedFileContent(string(content)) + "\n</file-content>"

			gitDiffFileToShow, _ := KeyGitCommits.ConcatKey(realm.Enable).HGet(relativePath)
			if len(gitDiffFileToShow.TextFragments) > 5 {
				gitDiffFileToShow.TextFragments = gitDiffFileToShow.TextFragments[len(gitDiffFileToShow.TextFragments)-5:]
			}

			commitStr := ""
			if len(gitDiffFileToShow.TextFragments) > 0 {
				commitStr = "\n<git-commits>\n" + gitDiffFileToShow.String() + "\n</git-commits>"
			}

			fileinfo := fmt.Sprint("\n<file>\n<file-name>", relativePath, "</file-name>"+fileSz+commitStr, fileContent, "\n</file>\n")

			allFileInfo.WriteString(fileinfo)
			return nil
		})
	}
	return allFileInfo.String()
}

// ## 潜在的改进目标：
// 1. 实现复习的调度算法
// 2. 现在的游戏是基于选择题目的方式来进行学习的。未来还需要探索3-4种不同的学习形态和相应的学习哲学。以便让不同的学习内容和不同的学习者都能找到合适的学习方式。
// 3. 需要构建合适的内容生成系统。
// 4. 客户端需要支持从服务端获取学习内容的能力。

// ## 改进目标对应的改进点：集成FSRS库到后端服务
// 目标1: 实现复习的调度算法（基于FSRS或其他间隔重复算法，共9个改进点）

// 实施方案：在Go服务器中使用github.com/open-spaced-repetition/go-fsrs包，通过import引入库，并在初始化函数中加载默认参数（如稳定性S和难度D）。
// 改进点：定义复习卡片数据结构

// 实施方案：创建一个Card结构体，包括字段如ID、知识点、最后复习时间、稳定性S、难度D，并使用Redis Hash存储每个用户的卡片数据。
// 改进点：实现卡片稳定性更新函数

// 实施方案：基于FSRS的UpdateStability函数，编写一个方法，输入用户评分（Grade）和当前S，输出新S；处理链完整情况下的5%提升。
// 改进点：计算下次复习时间

// 实施方案：使用FSRS的NextReviewTime函数，结合Retrievability Tolerance (R_t = 0.7 + 0.2 / sqrt(N_user + k))，调度复习间隔；k默认为10。
// 改进点：每日复习内容筛选

// 实施方案：编写一个查询函数，从Redis中筛选今日到期卡片（基于上次复习时间和间隔），并合并到当前学习主题包中。
// 改进点：处理用户评分转换

// 实施方案：实现Grade计算公式（g = 4 * sqrt(α * correctness^2 + β * (1 - timePercentile)^2)），动态调整α和β基于N_user。
// 改进点：集成时间百分位计算

// 实施方案：使用几何平均值计算timePercentile，权重基于sqrt(N_problem + k)和sqrt(N_user + k)；存储历史响应时间在Redis List中。
// 改进点：测试复习调度逻辑

// 实施方案：编写单元测试，模拟用户回答场景，验证S、D更新和下次复习时间计算的正确性。
// 改进点：添加日志监控复习过程

// 实施方案：使用logger记录每次复习的Grade、S变化和调度时间，便于调试和优化算法参数。

// 目标2: 探索3-4种不同的学习形态和相应的学习哲学（共9个改进点）

// 改进点：定义学习形态分类框架

// 实施方案：创建文档或枚举类型，分类4种形态（如选择题、构建式、讨论式、应用式），每个关联一种哲学（如行为主义、建构主义、人本主义、连接主义）。
// 改进点：探索选择题形态的哲学扩展

// 实施方案：基于当前游戏，添加行为主义哲学元素，如即时反馈奖励系统；测试用户在不同内容下的适应性。
// 改进点：设计构建式学习形态

// 实施方案：引入生成任务，如用户构建知识地图；哲学基于建构主义，生成AI辅助用户迭代内容。
// 改进点：实现讨论式学习形态原型

// 实施方案：添加聊天模拟界面，用户与AI角色讨论主题；哲学基于人本主义，强调情感支持和个性化对话。
// 改进点：开发应用式学习形态

// 实施方案：创建项目模拟环境，如数学应用到游戏中；哲学基于连接主义，链接外部API（如天气数据）生成真实场景。
// 改进点：评估不同形态的用户匹配

// 实施方案：设计问卷或A/B测试，收集用户偏好数据，匹配学习内容类型（如抽象 vs. 实践）到形态。
// 改进点：集成多形态切换机制

// 实施方案：在客户端添加模式选择UI，根据用户进度动态推荐形态；后端存储用户偏好。
// 改进点：文档化每种形态的哲学基础

// 实施方案：编写Markdown指南，解释每种形态的核心原则和预期学习效果，便于未来扩展。
// 改进点：试点测试3-4种形态

// 实施方案：选择小样本用户群，进行为期一周的测试，收集反馈并迭代一种形态的生成逻辑。

// 目标3: 构建合适的内容生成系统（共8个改进点）

// 改进点：选择内容生成API

// 实施方案：集成OpenAI GPT-4 API作为生成后端，设置prompt模板以确保知识点准确性。
// 改进点：设计生成prompt模板

// 实施方案：创建模板，包括主题、年龄、知识点权重和干扰项要求；确保每次生成4个知识点。
// 改进点：实现邻域主题混淆生成

// 实施方案：基于语义相似度，从邻域主题随机选k=1-3个，生成干扰知识点；使用LLM计算相似度。
// 改进点：添加生成质量校验

// 实施方案：后生成校验函数，检查正确性、一致性和多样性；如果失败，重生成。
// 改进点：支持多语言生成

// 实施方案：添加语言参数到prompt，初始支持中文和英文；测试生成内容的准确性。
// 改进点：优化生成成本

// 实施方案：缓存常见主题的生成结果在Redis中，设置TTL为1天；减少重复调用API。
// 改进点：集成用户反馈循环

// 实施方案：允许用户报告生成错误，后端记录并微调prompt模板。
// 改进点：测试生成多样性

// 实施方案：生成100个样本，分析知识点分布和干扰项质量，确保覆盖率>90%。

// 目标4: 客户端需要支持从服务端获取学习内容的能力（共6个改进点）

// 改进点：定义API端点

// 实施方案：在Go服务器添加GET /topics/{userId}端点，返回今日学习主题包JSON。
// 改进点：客户端集成API调用

// 实施方案：在React中使用fetch或axios调用服务端API，替换mock数据。
// 改进点：处理认证和用户ID

// 实施方案：添加JWT认证，客户端存储token；服务端验证用户并返回个性化内容。
// 改进点：实现内容缓存机制

// 实施方案：在客户端使用localStorage缓存最近主题，减少网络请求；设置过期时间1小时。
// 改进点：添加错误处理和重试

// 实施方案：客户端在API失败时显示提示，并自动重试3次；日志错误到控制台。
// 改进点：测试端到端数据流

// 实施方案：模拟服务端响应，验证客户端从加载到显示内容的完整流程。

type TextFragment struct {
	Comment     string   `description:"string, The git commit comment associated with the text fragment"`
	OldPosition int64    `description:"integer, The old position of the text fragment"`
	NewPosition int64    `description:"integer, The new position of the text fragment"`
	Lines       []string `description:"array, The lines in the text fragment. Each line starts with a space (context), '+' (addition context), or '-' (deletion context)"`
}
type GitCommitUsingUnifiedDiffFormat struct {
	OldName string `description:"string, The old file name"`
	NewName string `description:"string, The new file name"`

	IsNew    bool `description:"boolean, Whether the file is new"`
	IsDelete bool `description:"boolean, Whether the file is deleted"`
	IsCopy   bool `description:"boolean, Whether the file is copied"`
	IsRename bool `description:"boolean, Whether the file is renamed"`

	TextFragments []*TextFragment
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
目标1.1 改进目标对应的改进点：集成FSRS库到后端服务
</system evolution Goals>

<Implementation Steps>
实施中间步骤:
1. **初步改进方案**：基于现有方案，提出一个准确、可靠的改进方案,以便实施evolution Goals。
2. **基于行为准则的方案强化**：对改进方案进行进一步评估优化，以确保新的改进方案能够克服对本系统的行为准则的破坏。重点评估并优化的领域包括：目标明确、用户价值、结构质量、可维护性、性能与可靠性。
	- **目标明确**：明确围绕特定的目标来提升现有方案。高质量实施给定目标，最小化副作用。
	- **用户价值**：强化业务场景覆盖度、用户满意度和长期价值。
	- **结构质量**：应着重在认知复杂度、圈复杂度、模块耦合度、内聚度和代码简洁度方面进行优化。
	- **可维护性**：应关注核心逻辑文档覆盖率的提升，确保代码可理解和可测试。
	- **性能与可靠性**：优化代码的正确性、变更失败率、预估延迟和吞吐量。
	针对评估中发现的问题进一步优化，对每个重要缺陷给出具体的优化方案。确保在当前能力圈的安全边际内进行调整。
3. 尝试生成最终的优化方案。
请注意。调用GitCommitUsingUnifiedDiffFormat生成的Git Unified Diff文件普遍存在行数量和真实的行变动不匹配得问题。为了解决这个问题。这里务必先写下完整的修改内容。等正式提交改进方案的时候，需要重新校验，并确保与实际变更一致。

## 提交最终改进方案
最后通过 N次独立的toolcall调用: GitCommitUsingUnifiedDiffFormat 来提交一个使用 Git Unified Diff 格式的代码变更。
</Implementation Steps>

`))).WithToolCallMutextRun().WithTools(tool.NewTool("GitCommitUsingUnifiedDiffFormat", "提交一个使用 Git Unified Diff 格式的代码变更", func(file *GitCommitUsingUnifiedDiffFormat) {

	OldName, _ := utils.ToLocalEvoFile(file.OldName)
	newName, realmNew := utils.ToLocalEvoFile(file.NewName)

	if file.IsDelete {
		os.Remove(OldName)
		return
	} else if file.IsNew {
		os.MkdirAll(filepath.Dir(newName), 0o755)
	} else if file.IsCopy {
		os.MkdirAll(filepath.Dir(newName), 0o755)
		os.Link(OldName, newName)
		return
	} else if file.IsRename {
		os.MkdirAll(filepath.Dir(newName), 0o755)
		os.Link(OldName, newName)
		os.Remove(OldName)
	}
	gitDiffFile, err := KeyGitCommits.ConcatKey(realmNew.Name).HGet(file.NewName)
	if err != nil {
		gitDiffFile = gitdiff.File{OldName: file.OldName, NewName: file.NewName, IsNew: file.IsNew, IsDelete: file.IsDelete, IsCopy: file.IsCopy, IsRename: file.IsRename}
	}
	for _, fragment := range file.TextFragments {
		// count the types of lines in the fragment content
		frag := gitdiff.TextFragment{Comment: fragment.Comment, OldPosition: fragment.OldPosition, NewPosition: fragment.NewPosition}
		for _, line := range fragment.Lines {
			if len(line) == 0 {
				continue
			}
			gline := gitdiff.Line{Line: line[1:]}
			var addedLines, deletedLines int64 = 0, 0
			switch line[0] {
			case ' ':
				gline.Op = gitdiff.OpContext
				frag.OldLines++
				frag.NewLines++
				if addedLines == 0 && deletedLines == 0 {
					frag.LeadingContext++
				} else {
					frag.TrailingContext++
				}
			case '+':
				gline.Op = gitdiff.OpAdd
				frag.NewLines++
				addedLines++
				frag.TrailingContext = 0
			case '-':
				gline.Op = gitdiff.OpDelete
				frag.OldLines++
				deletedLines++
				frag.TrailingContext = 0
			}
			frag.Lines = append(frag.Lines, gline)
		}
	}
	content, err := os.Open(OldName)
	if err != nil {
		return
	}
	var output bytes.Buffer
	if err = gitdiff.Apply(&output, content, &gitDiffFile); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(newName, output.Bytes(), 0o644); err != nil {
		return
	}

}))

func MakeAEvo() {

	for i, TotalTasks := 0, 2000; i < TotalTasks; i++ {
		time.Sleep(300 * time.Millisecond)
		ProductGoalUniLearning := utils.TextFromFile("/Users/yang/learn-by-choose-goserver/learninggame.md")

		// SolutionSummary := LoadAllEvoProjects()
		// messege := AgentEvoLearningSolutionLearnByChoose.Messege(map[string]any{
		// 	"ProductGoal": string(ProductGoalUniLearning) + "\n\n",
		// 	"Solution":    SolutionSummary,
		// })
		// param := map[string]any{"Context": messege, "Result": []string{}}
		// agent.AgentSelectContextFiles.Call(context.Background(), param)
		// ResultRelatedFileNames, _ := param["Result"].([]string)

		ResultRelatedFileNames := []string{
			"/learniversebackend/fsrs_example.go",
			"/learniversebackend/fsrs_integrate.go",
			"/learniversebackend/fsrs_interfaces.go",
			"/Learniversebackend/fsrs_refine.go",
			"/Learniversebackend/fsrs_stats.go",
			"/learniversebackend/Learninggame.md",
		}

		SolutionSummaryTrimed := LoadAllEvoProjects(ResultRelatedFileNames)
		errorGroup := errgroup.Group{}
		errorGroup.Go(func() error {
			//Gemini25Flashlight Gemini25ProAigpt Glm45AirLocal
			return AgentEvoLearningSolutionLearnByChoose.WithModels(models.Qwen3B235Thinking2507).WithMsgDeClipboard(). //CopyPromptOnly(). //Qwen3B32Thinking
																	Call(context.Background(), map[string]any{
					"ProductGoal": string(ProductGoalUniLearning) + "\n\n",
					"Solution":    SolutionSummaryTrimed,
				})
		})
		err := errorGroup.Wait()
		if err != nil {
			fmt.Printf("Agent call failed: %v\n", err)
		}
	}

}

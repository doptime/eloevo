package evobymeasured

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/config"
	"github.com/doptime/eloevo/elo"
	"github.com/doptime/eloevo/evo"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type Solution struct {
	GoalToAchieve string                    `description:"The name of the goal this solution is intended to achieve."`
	Edits         []*TextFragmentsEdited    `description:"The list of text fragment edits that make up this solution."`
	EloScore      int64                     `description:"-"` //该方案的Elo评分
	EvolutionID   string                    `description:"-"` //该方案的唯一ID,一般是时间戳+随机数
	FileBefore    map[string]string         `description:"-"` // before modification, key is filename, value is file content
	ModifiedLines map[string]map[int]string `description:"-"` // after modification, key is filename, value is file content
	Diffs         map[string]string         `description:"-"` // after modification, key is filename, value is git diff content
}

func (s *Solution) GetId() string {
	return s.EvolutionID
}
func (s *Solution) ScoreAccessor(delta ...int) int {
	if len(delta) > 0 {
		s.EloScore += int64(delta[0])
	}
	return int(s.EloScore)
}
func (s *Solution) RemoveDueToFileChanged(key *redisdb.HashKey[string, *Solution]) bool {
	for _, file := range lo.Keys(s.FileBefore) {
		OriginalContent := s.FileBefore[file]
		FileName, _ := utils.ToLocalEvoFile(file)
		RawFileStr := utils.TextFromFile(FileName)
		if RawFileStr != OriginalContent {
			delete(s.FileBefore, file)
			delete(s.ModifiedLines, file)
			delete(s.Diffs, file)
			//also remove the edits related to this file
			s.Edits = lo.Filter(s.Edits, func(edit *TextFragmentsEdited, _ int) bool {
				return edit.FileName != file
			})
		}
	}
	if len(s.FileBefore) == 0 {
		key.HDel(s.EvolutionID)
		return true
	} else {
		key.HSet(s.EvolutionID, s)
		return false
	}
}

type TextFragmentsEdited struct {
	GoalName string `description:"The name of the goal this change is associated with, used as an identifier."`

	FileName string `description:"required. The name of the file before the change."`
	NewName  string `description:"optional when renaming or copying. The new name of the file changed to."`

	IsNew    bool `description:"Indicates if the file is newly added in this commit."`
	IsDelete bool `description:"Indicates if the file was deleted in this commit."`
	IsCopy   bool `description:"Indicates if the file was copied from another file in this commit."`
	IsRename bool `description:"Indicates if the file was renamed in this commit."`

	EvolutionParentIDs []string `description:"当前该方案应该从一个最有价值或者是潜力的方案的派生，这是父派生方案的ID。如果是新方案,则为空"`

	Comment string `description:"required. The git commit message or comment associated with the hunk. This provides what was done in this changes."`

	CommitValueDeclaration string `description:"A brief declaration of the value or purpose of this commit. This should be a concise summary that highlights the main intent behind the changes."`

	OldFragmentStartLine                int64  `description:"The starting line number in the original file from which this fragment begins.(1-minimal)"`
	OldFragmentEndLine                  int64  `description:"The end line number in the original file from which this fragment begins.(1-minimal)"`
	NewFragmentText_NoLeadingLineNumber string `description:"A strings of mutiple lines, representing the new lines in this fragment. The Old TextFragment will be replaced by this TextFragment"`

	Realm *config.EvoRealm `description:"-" msgpack:"-"`
	Goal  *evo.Goal        `description:"-" msgpack:"-"` //目标

	EvolutionID string                              `description:"-"` //该方案的唯一ID,一般是时间戳+随机数
	SolutionKey *redisdb.HashKey[string, *Solution] `description:"-" msgpack:"-"`
}

func (solution *Solution) String() string {
	var solutionStrBuilder strings.Builder
	solutionStrBuilder.WriteString(fmt.Sprintf("<Evolution ID=%s>\n", solution.EvolutionID))
	solutionStrBuilder.WriteString(fmt.Sprintf("\t<GoalToAchieve>%s</GoalToAchieve>\n", solution.GoalToAchieve))
	for _, edits := range solution.Edits {
		NewFragmentTextLines := "\n\t<NewFragmentTextLines>\n" + edits.NewFragmentText_NoLeadingLineNumber + "\n\t</NewFragmentTextLines>"
		ChangesOfUnifiedDiffFormat := "\n\t<ChangesInUnifiedDiffFormat>\n" + solution.Diffs[edits.FileName] + "\n\t</ChangesInUnifiedDiffFormat>"
		Comment := "\n\t<Comment>\n" + edits.Comment + "\n\t</Comment>\n"
		solutionStrBuilder.WriteString(fmt.Sprintf("\t<TextFragmentsEdited  FileName=%s NewName=%s IsNew=%v IsDelete=%v IsCopy=%v IsRename=%v EvolutionParentIDs=%v OldFragmentStartLine=%d, OldFragmentEndLine=%d>%s %s %s \n\t</TextFragmentsEdited>", edits.FileName, edits.NewName, edits.IsNew, edits.IsDelete, edits.IsCopy, edits.IsRename, edits.EvolutionParentIDs, edits.Comment, edits.OldFragmentStartLine, edits.OldFragmentEndLine, Comment, NewFragmentTextLines, ChangesOfUnifiedDiffFormat))
	}
	solutionStrBuilder.WriteString(fmt.Sprintf("</Evolution ID=%s>\n", solution.EvolutionID))
	return solutionStrBuilder.String()
}
func (solution *Solution) CommitResultToFile() {
	for _, edits := range solution.Edits {

		edits.FileName = lo.CoalesceOrEmpty(edits.FileName, edits.NewName)
		edits.NewName = lo.CoalesceOrEmpty(edits.NewName, edits.FileName)
		FileName, _ := utils.ToLocalEvoFile(edits.NewName)
		newName, _ := utils.ToLocalEvoFile(edits.NewName)
		if edits.IsDelete {
			os.Remove(edits.FileName)
			return
		} else if edits.IsNew {
			os.MkdirAll(filepath.Dir(newName), 0o755)
		} else if edits.IsCopy {
			os.MkdirAll(filepath.Dir(newName), 0o755)
			os.Link(FileName, newName)
			return
		} else if edits.IsRename {
			os.MkdirAll(filepath.Dir(newName), 0o755)
			os.Link(FileName, newName)
			os.Remove(FileName)
		}
	}
	for file, lines := range solution.ModifiedLines {
		FileName, _ := utils.ToLocalEvoFile(file)

		if FileName == "" {
			fmt.Println("OldName is empty, bad case")
			return
		}
		var contentStrBuilder strings.Builder
		keys := lo.Keys(lines)
		slices.Sort(keys)
		for _, key := range keys {
			contentStrBuilder.WriteString(lines[key] + "\n")
		}
		if err := os.WriteFile(FileName, []byte(contentStrBuilder.String()), 0o644); err != nil {
			return
		}

	}
	//save file diffs
	for file, diff := range solution.Diffs {
		_, Realm := utils.ToLocalEvoFile(file)
		gitdiffHistoryFile := Realm.EvoFile(config.EvoUnifiedDiffFormat, file)
		utils.AppendGitDiff(gitdiffHistoryFile, diff)
	}

}

var keySolution = redisdb.NewHashKey[string, *Solution]()

var AgentEvoAGoal = agent.Create(template.Must(template.New("AgentEvoLearningSolutionLearnByChoose").Parse(`
# 系统演化任务描述:

<Current System>
{{.CurrentSystem}}
</Current System>

<SystemEvolutionGoal>
{{.SystemEvolutionGoal}}
</SystemEvolutionGoal>

<ParentSolutions>
{{range .ParentSolutions}}
{{.}}
{{end}}
</ParentSolutions>

<Implementation Steps>
实施中间步骤:
step1. **找出需要借鉴的最佳父分支方案**：对现有的改进方案进行评估，以便新的方案可以在此基础上进行改进。
	现有改进方案位于ParentSolutions当中。 这些改进方案重点评估并优化的领域包括：目标明确、第一性原理、用户价值、结构质量、可维护性、性能与可靠性。
	- **目标明确**：明确围绕特定的目标来提升现有方案。高质量实施给定目标，最小化副作用。
	- **第一性原理**：准确平衡全面的内容或事实，逻辑自洽。改进方案应基于第一性原理，避免认知偏差。
	- **用户价值**：强化业务场景覆盖度、用户满意度和长期价值。
	- **结构质量**：应着重在认知复杂度、模块耦合度、内聚度和代码简洁度方面进行优化。
	- **可维护性**：应关注核心逻辑文档覆盖率的提升，确保代码可理解和可测试。
	- **性能与可靠性**：优化代码的正确性、变更失败率、预估延迟和吞吐量。
	显性且仔细思考父分支方案的优劣。以便选定并借鉴优秀的父分支，并进行改进。

step2. **制定改进方案**：
	- 1. 讨论，并且提出一个基于有限理性、基于第一性原理的关键改进思路，然后按照行动优先，探索优先的思路，用代码或文本给出具体的实质的改进。
	- 2. 融合父方案进行基于缺陷最小化的融合。给出融合思路+改进思路下的具体的改进方案

step3. **方案提交之前的审核修正**： 给出符合 TextFragmentsEdited 约定的 增量编辑的优化方案。并显式检查其中参数是否准确无误。特别是确定需要替换的旧文本的在原始文件的行范围，也就是OldFragmentStartLine（delete included）, OldFragmentEndLine（delete included）。确认或者是修改参数，对TextFragmentsEdited调用符合意图，没有异常。

## 提交最终改进方案
最后通过 N次独立的toolcall: TextFragmentsEdited,以分段、增量修改的方式，每个调用仅完成一个文件操作或者是文件中的单个代码块的内容变更。
</Implementation Steps>
<think>
`))).WithToolCallMutextRun().WithTools(tool.NewTool("TextFragmentsEdited", "提交代码文本变更，针对1)文件变动 2)内容变动", func(edits *TextFragmentsEdited) {

	edits.NewName = lo.CoalesceOrEmpty(edits.NewName, edits.FileName)
	edits.FileName = lo.CoalesceOrEmpty(edits.FileName, edits.NewName)

	FileName, _ := utils.ToLocalEvoFile(edits.NewName)
	if edits.NewName == "" || edits.GoalName == "" {
		fmt.Println("OldName is empty, bad case")
		return
	}
	//lock the file
	var Solutions = map[string]*Solution{}
	Solutions, _ = edits.SolutionKey.HGetAll()
	for _, solutionKey := range lo.Keys(Solutions) {
		if Solutions[solutionKey].RemoveDueToFileChanged(edits.SolutionKey) {
			delete(Solutions, solutionKey)
		}
	}

	edits.SolutionKey.HGet(edits.EvolutionID)
	if _, found := Solutions[edits.EvolutionID]; !found {
		Solutions[edits.EvolutionID] = &Solution{
			EvolutionID:   edits.EvolutionID,
			GoalToAchieve: edits.GoalName,
			Edits:         []*TextFragmentsEdited{},
			EloScore:      1200,
			FileBefore:    map[string]string{},
			ModifiedLines: map[string]map[int]string{},
			Diffs:         map[string]string{},
		}
	}
	currentSolution := Solutions[edits.EvolutionID]
	currentSolution.Edits = append(Solutions[edits.EvolutionID].Edits, edits)

	//ioFileToSave, _ = os.OpenFile(SolutionFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)

	_, NonFirstTrail := currentSolution.FileBefore[edits.NewName]
	if !NonFirstTrail {
		RawFileStr := utils.TextFromFile(FileName)
		currentSolution.FileBefore[edits.NewName] = RawFileStr
		currentSolution.ModifiedLines[edits.NewName] = utils.LineNumberedMap(RawFileStr, 1)
	}

	for i := edits.OldFragmentStartLine; i <= edits.OldFragmentEndLine; i++ {
		delete(currentSolution.ModifiedLines[edits.NewName], int(i))
	}
	currentSolution.ModifiedLines[edits.NewName][int(edits.OldFragmentStartLine)] = utils.RemoveLineNumber(edits.NewFragmentText_NoLeadingLineNumber)
	parentNodes := lo.Filter(lo.Values(Solutions), func(s *Solution, i int) bool {
		return lo.Contains(edits.EvolutionParentIDs, s.EvolutionID)
	})

	//update elo score
	parentElo := lo.Map(parentNodes, func(s *Solution, _ int) elo.Elo { return s })
	allElo := lo.Map(lo.Values(Solutions), func(s *Solution, _ int) elo.Elo { return s })
	elo.BatchUpdateWinnings(parentElo, allElo)

	var contentStrBuilder strings.Builder
	keys := lo.Keys(currentSolution.ModifiedLines[edits.NewName])
	slices.Sort(keys)
	for _, key := range keys {
		contentStrBuilder.WriteString(currentSolution.ModifiedLines[edits.NewName][key] + "\n")
	}
	//save the git diff

	_edits := myers.ComputeEdits(span.URIFromPath(edits.FileName), currentSolution.FileBefore[edits.NewName], contentStrBuilder.String())
	diff := gotextdiff.ToUnified(edits.FileName, edits.FileName, currentSolution.FileBefore[edits.NewName], _edits)

	currentSolution.Diffs[edits.FileName] = time.Now().Format("2006-01-02 15:04:05") + " " + edits.Comment + "\n" + fmt.Sprintln(diff)
	edits.SolutionKey.HSet(edits.EvolutionID, currentSolution)
	//also update parent nodes to save elo score
	for _, solution := range parentNodes {
		edits.SolutionKey.HSet(solution.EvolutionID, solution)
	}
	//unlock the file

}))

func EvoAGoal(RealmStr, GoalNameAsID string) {
	realm := config.WithSelectedRealms(RealmStr)[0]
	goal := evo.LoadGoal(realm, "CreateUnifiedCookbook")
	if goal == nil {
		fmt.Println("No goal found")
		return
	}
	solutionKey := keySolution.ConcatKey(realm.Name).ConcatKey(GoalNameAsID)
	for i, TurnNum := 0, 6; i < TurnNum; i++ {
		time.Sleep(300 * time.Millisecond)
		//key: evolutionID, value: []*TextFragmentsEdited
		var Solutions = map[string]*Solution{}
		Solutions, _ = solutionKey.HGetAll()
		for _, key := range lo.Keys(Solutions) {
			if Solutions[key].RemoveDueToFileChanged(solutionKey) {
				delete(Solutions, key)
			}
		}
		errorGroup := errgroup.Group{}
		for j := 0; j < 2; j++ {
			errorGroup.Go(func() error {
				//Gemini25Flashlight Gemini25ProAigpt Glm45AirLocal
				model := []*models.Model{models.Qwendeepresearch, models.Qwen3Next80B}[j%2]
				return AgentEvoAGoal. //CopyPromptOnly(). //Qwen3B32Thinking
							Call(map[string]any{
						agent.UseModel:        model,
						"SystemEvolutionGoal": string(goal.String()) + "\n\n",
						"CurrentSystem":       config.WithSelectedRealms("RedisDB").LoadAllEvoProjects(goal.RelatedFiles...),
						"ParentSolutions":     lo.Values(Solutions),
						"EvolutionID":         utils.ID(nil, 8),
						"SolutionKey":         solutionKey,
					})
			})
		}
		err := errorGroup.Wait()
		if err != nil {
			fmt.Printf("Agent call failed: %v\n", err)
		}
	}
	//TODO: now TurnNum has been done, it's time to select the best solution	 and commit it to the goal file
	var Solutions = map[string]*Solution{}
	Solutions, _ = solutionKey.HGetAll()
	solutionsList := lo.Values(Solutions)
	slices.SortFunc(solutionsList, func(a, b *Solution) int {
		return int(b.EloScore - a.EloScore)
	})
	if len(solutionsList) > 0 {
		BestSolution := solutionsList[0]
		fmt.Println("Best solution found:", BestSolution.String())
		BestSolution.CommitResultToFile()
	}

}

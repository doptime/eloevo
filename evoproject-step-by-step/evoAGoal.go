package evoprojectstepbystep

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
	"github.com/doptime/eloevo/evo"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/samber/lo"
)

type Solution struct {
	GoalToAchieve string                    `description:"The name of the goal this solution is intended to achieve."`
	Edits         []*TextFragmentsEdited    `description:"The list of text fragment edits that make up this solution."`
	EvolutionID   string                    `description:"-"` //该方案的唯一ID,一般是时间戳+随机数
	FileBefore    map[string]string         `description:"-"` // before modification, key is filename, value is file content
	ModifiedLines map[string]map[int]string `description:"-"` // after modification, key is filename, value is file content
	Diffs         map[string]string         `description:"-"` // after modification, key is filename, value is git diff content
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

	KeyConsiderations      string `description:"Key considerations or important aspects that were taken into account while making this change. This could include facts, design principles, constraints, or specific requirements that influenced the change."`
	FocusedSettlements     string `description:"The specific areas or aspects that were the primary focus of this change. This could include performance improvements, bug fixes, feature additions, code refactoring, or any other targeted objectives."`
	CommitValueDeclaration string `description:"A brief declaration of the value or purpose of this commit. This should be a concise summary that highlights the main intent behind the changes."`

	Comment string `description:"required. The git commit message or comment associated with the hunk. This provides what was done in this changes."`

	OldFragmentStartLine                int64  `description:"The starting line number in the original file from which this fragment begins.(1-minimal)"`
	OldFragmentEndLine                  int64  `description:"The end line number in the original file from which this fragment begins.(1-minimal)"`
	NewFragmentText_NoLeadingLineNumber string `description:"A strings of mutiple lines, representing the new lines in this fragment. The Old TextFragment will be replaced by this TextFragment.不含行号, 系统会自动处理行号"`

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
		solutionStrBuilder.WriteString(fmt.Sprintf("\t<TextFragmentsEdited  FileName=%s NewName=%s IsNew=%v IsDelete=%v IsCopy=%v IsRename=%v KeyConsiderations=\"%s\" FocusedSettlements=\"%s\" CommitValueDeclaration=\"%s\" Comment=\"%s\" OldFragmentStartLine=%d, OldFragmentEndLine=%d>%s %s %s \n\t</TextFragmentsEdited>", edits.FileName, edits.NewName, edits.IsNew, edits.IsDelete, edits.IsCopy, edits.IsRename, edits.KeyConsiderations, edits.FocusedSettlements, edits.CommitValueDeclaration, edits.Comment, edits.OldFragmentStartLine, edits.OldFragmentEndLine, Comment, NewFragmentTextLines, ChangesOfUnifiedDiffFormat))
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
			continue
		} else if edits.IsNew {
			os.MkdirAll(filepath.Dir(newName), 0o755)
		} else if edits.IsCopy {
			os.MkdirAll(filepath.Dir(newName), 0o755)
			os.Link(FileName, newName)
			continue
		} else if edits.IsRename {
			os.MkdirAll(filepath.Dir(newName), 0o755)
			os.Link(FileName, newName)
			os.Remove(FileName)
			continue
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
			continue
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
You are a deep system evolute assistant . You should analyze the given system ,and make some changes based on the tools provided. The TODO Goal was defined in SystemEvolutionGoal, your core function is to conduct thorough, multi-source investigations into any topic. You must handle both broad, open-domain inquiries and queries within specialized academic fields. For every request, synthesize information from credible, diverse sources to deliver a comprehensive, accurate, and objective response. When you have gathered sufficient information and are ready to provide the definitive response.

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
step1. **评估父分支方案**：对已有的改进方案的实施框架进行评估，以便新的方案可以在此基础上进行改进。
	现有改进方案位于ParentSolutions当中。 这些改进方案重点评估并优化的领域包括但不受限于：目标明确、第一性原理、用户价值、结构质量、可维护性、性能与可靠性。
	- **目标明确**：明确围绕特定的目标来提升现有方案。高质量实施给定目标，最小化副作用。
	- **第一性原理**：准确平衡全面的内容或事实，逻辑自洽。改进方案应基于第一性原理，避免认知偏差。
	- **用户价值**：强化业务场景覆盖度、用户满意度和长期价值。
	- **结构质量**：应着重在认知复杂度、模块耦合度、内聚度和代码简洁度方面进行优化。
	- **可维护性**：应关注核心逻辑文档覆盖率的提升，确保代码可理解和可测试。
	- **性能与可靠性**：优化代码的正确性、变更失败率、预估延迟和吞吐量。
	仔细思考父分支方案的优劣。以便选定需要借鉴的优秀父特性。同时，也需要及时删除那些已经变得冗余、过时、不再适用的父分支方案（通过填写EvolutionShouldBeRemovedIDs）。
	

step2. **制定Step-By-Step改进方案实施框架**：
	- 1. 讨论，并且提出一个多步骤的改进方案框架。这些方案主要包括改进现有解决方案的缺陷，或者是增强解决方案的特性。
	- 2. 仔细实施Step-By-Step中的每一步。每一步都需要有明确的结果。如果结果并不符合预期，需要继续改进结构。对所有的步骤都以这种方式实施改进。

step3. **方案提交之前的审核修正**： 给出符合 TextFragmentsEdited 约定的 增量编辑的解决方案。
并显式检查其中参数是否准确无误。特别是确定需要替换的旧文本的在原始文件的行范围，也就是OldFragmentStartLine（delete included）, OldFragmentEndLine（delete included）。确认或者是修改参数，对TextFragmentsEdited调用符合意图，没有异常。

step4. **提交最终改进方案**：
	通过 N次独立的toolcall: TextFragmentsEdited,以分段、增量修改的方式，每个调用仅完成一个文件操作或者是文件中的单个代码块的内容变更。
</Implementation Steps>
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

	edits.SolutionKey.HGet(edits.EvolutionID)
	if _, found := Solutions[edits.EvolutionID]; !found {
		Solutions[edits.EvolutionID] = &Solution{
			EvolutionID:   edits.EvolutionID,
			GoalToAchieve: edits.GoalName,
			Edits:         []*TextFragmentsEdited{},
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
	//unlock the file

}))

func TakeAGoal(realm *config.EvoRealm) (GoalFile string) {
	goalPath := realm.RootPath + "/!system_evolution/todo/"
	for {
		FilesNamesInGoalDir := utils.FilesNamesInDir(goalPath)
		if len(FilesNamesInGoalDir) == 0 {
			fmt.Println("No available goal found, wait for 3 seconds")
			time.Sleep(3 * time.Second)
			continue
		}
		DoingFileName := strings.Replace(FilesNamesInGoalDir[0], "/todo/", "/doing/", 1)
		os.Rename(FilesNamesInGoalDir[0], DoingFileName)
		return DoingFileName
	}
}
func StartAGoal(GoalFile string, realm *config.EvoRealm, Realms ...string) {

	const TurnNum = 6
	modelList := models.NewModelList("Qwen3Next80b", models.Qwen3Next80B, models.Qwen3vl30b)
	relativePath := realm.RelativePath(GoalFile)
	GoalContent := utils.TextFromFile(GoalFile)

	//save the best solution to the goal file
	const falseSave = false
	if falseSave {
		GoalName := strings.Split(filepath.Base(GoalFile), ".")[0]
		solutionKey := keySolution.ConcatKey(realm.Name).ConcatKey(GoalName)
		var Solutions = map[string]*Solution{}
		Solutions, _ = solutionKey.HGetAll()
		solutionsList := lo.Values(Solutions)

		slices.SortFunc(solutionsList, func(a, b *Solution) int {
			return strings.Compare(b.EvolutionID, a.EvolutionID)
		})

		if len(solutionsList) > 0 {
			BestSolution := solutionsList[0]
			fmt.Println("Best solution found:", BestSolution.String())
			BestSolution.CommitResultToFile()
		}

		//label the goal file as done
		os.Rename(GoalFile, strings.Replace(GoalFile, "/doing/", "/done/", 1))

	}

	//each turn, call the agent to generate new solutions
	for turni := 0; turni < TurnNum; turni++ {
		time.Sleep(300 * time.Millisecond)

		for p, parallelN := 0, 2; p < parallelN; p++ {
			GoalName := strings.Split(filepath.Base(GoalFile), ".")[0]
			solutionKey := keySolution.ConcatKey(realm.Name).ConcatKey(GoalName)

			//key: evolutionID, value: []*TextFragmentsEdited
			var Solutions = map[string]*Solution{}
			Solutions, _ = solutionKey.HGetAll()
			//for _, key := range lo.Keys(Solutions) { if Solutions[key].RemoveDueToFileChanged(solutionKey) { delete(Solutions, key) } }

			AgentEvoAGoal.Call(map[string]any{
				//agent.UseCopyPromptOnly: true,
				agent.UseModel: modelList.SequentialPick(), //.WithToolsInUserPrompt(), //CopyPromptOnly().
				//  //Qwen3B32Thinking  model := []*models.Model{models.Qwendeepresearch, models.Qwen3Next80B,models.Qwen3B235Thinking2507}[j%2]
				"SystemEvolutionGoal": "\n\n<GoalFile path=\"" + relativePath + "\">\n" + GoalContent + "\n</GoalFile>\n",
				"CurrentSystem":       config.WithSelectedRealms(Realms...).LoadAllEvoProjects(),
				"ParentSolutions":     lo.Values(Solutions),
				"EvolutionID":         fmt.Sprintf("%d-%d-%d", time.Now().UnixNano(), turni, p),
				"SolutionKey":         solutionKey,
			})

			//now TurnNum has been done, it's time to select the best solution	 and commit it to the goal file
			if turni == TurnNum-1 && p == parallelN-1 {
				GoalName := strings.Split(filepath.Base(GoalFile), ".")[0]
				solutionKey := keySolution.ConcatKey(realm.Name).ConcatKey(GoalName)
				var Solutions = map[string]*Solution{}
				Solutions, _ = solutionKey.HGetAll()
				solutionsList := lo.Values(Solutions)

				slices.SortFunc(solutionsList, func(a, b *Solution) int {
					return strings.Compare(b.EvolutionID, a.EvolutionID)
				})

				if len(solutionsList) > 0 {
					BestSolution := solutionsList[0]
					fmt.Println("Best solution found:", BestSolution.String())
					BestSolution.CommitResultToFile()
				}

				//label the goal file as done
				os.Rename(GoalFile, strings.Replace(GoalFile, ".doing", ".done", 1))

			}
		}
	}
}

func EvoAGoal(Realms ...string) {
	AllRealms := config.WithSelectedRealms(Realms...)
	if len(AllRealms) == 0 {
		fmt.Println("No realm found")
		return
	}

	for {
		realm := AllRealms[0]
		GoalFile := TakeAGoal(realm)
		go StartAGoal(GoalFile, realm, Realms...)
	}
}

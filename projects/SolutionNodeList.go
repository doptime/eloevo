package projects

import (
	"fmt"
	"slices"
	"strings"

	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

//不使用超边节点

type SolutionNode struct {
	Filename          string `description:"string, Ascii filename of current node。using bullet name to denodes node's modualized intention. extension name such as .md ... is needed"`
	Delete            bool   `msgpack:"-" description:"bool, Whether this node is deleted. If true, the node will be removed from the system."`
	BulletDescription string `description:"string, Required when create. BulletDescription 是文件内容的摘要。描述和文件的模块化的意图。规定实现的细节."`
	FileContent       string `description:"string, The contents of the file stored on disk"`

	// 新增字段：用于将迭代反馈传递给 EndGoalDrivenPlanner
	NextIterFeedback string `description:"string, 当前开发迭代的反馈和可操作性洞察，例如已识别的缺失依赖项、约束冲突或潜在的冗余实现。此信息将传递给 EndGoalDrivenPlanner，以用于后续迭代中的战略考量和任务分配。"`

	//被人类专家标记为锁定的条目。Locked = true. 不能被删除和修改
	Locked bool `description:"-"`

	HashKey redisdb.HashKey[string, *SolutionNode] `description:"-" msgpack:"-"`
}

type SolutionNodeList []*SolutionNode

func (a SolutionNodeList) Uniq() SolutionNodeList { return lo.Uniq(a) }
func (a SolutionNodeList) LockedOnly() SolutionNodeList {
	return lo.Filter(a, func(v *SolutionNode, _ int) bool { return v.Locked })
}
func (a SolutionNodeList) FullView() string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Filename, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		NextIterFeedback := lo.Ternary(v.NextIterFeedback != "", " NextIterFeedback: "+v.NextIterFeedback, "")

		s := fmt.Sprint(indence, "\n Pathname", v.Filename, "\nBulletDescription: ", v.BulletDescription, "\nFileContent: ", v.FileContent, NextIterFeedback, "\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a SolutionNodeList) View(FullViewPaths []string) string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Filename, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		NextIterFeedback := lo.Ternary(v.NextIterFeedback != "", " NextIterFeedback: "+v.NextIterFeedback, "")
		fileContent := lo.Ternary(slices.Contains(FullViewPaths, v.Filename), "\nFileContent: "+v.FileContent, "")

		s := fmt.Sprint(indence, "\n Pathname", v.Filename, "\nBulletDescription: ", v.BulletDescription, fileContent, NextIterFeedback, "\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a SolutionNodeList) SummaryView() string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Filename, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		NextIterFeedback := lo.Ternary(v.NextIterFeedback != "", " NextIterFeedback: "+v.NextIterFeedback, "")

		s := fmt.Sprint(indence, "\n Pathname", v.Filename, "\nBulletDescription: ", v.BulletDescription, NextIterFeedback, "\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a SolutionNodeList) PathnameSorted() SolutionNodeList {
	slices.SortFunc(a, func(a, b *SolutionNode) int {
		return strings.Compare(a.Filename, b.Filename)
	})
	return a
}

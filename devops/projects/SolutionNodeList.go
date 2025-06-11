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
	Pathname          string `description:"string, Ascii pathname of current node。pathname is multi-level, using bullet name to denodes node's modualized intention. extension name such as .md ... is needed"`
	Delete            bool   `msgpack:"-" description:"bool, Whether this node is deleted. If true, the node will be removed from the system."`
	BulletDescription string `description:"string, Required when create. BulletDescription 是文件内容的摘要。描述和文件的模块化的意图。规定实现的细节."`
	FileContent       string `description:"string, The contents of the file stored on disk"`

	SelfMemo string `description:"Self memo information. Such as missing elements, constrains collision, or what to do in further iterations."`

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
		numlayter := len(strings.Split(v.Pathname, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		SelfMemo := lo.Ternary(v.SelfMemo != "", " SelfMemo: "+v.SelfMemo, "")

		s := fmt.Sprint(indence, "\n Pathname", v.Pathname, "\nBulletDescription: ", v.BulletDescription, "\nFileContent: ", v.FileContent, SelfMemo, "\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a SolutionNodeList) View(FullViewPaths []string) string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Pathname, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		SelfMemo := lo.Ternary(v.SelfMemo != "", " SelfMemo: "+v.SelfMemo, "")
		fileContent := lo.Ternary(slices.Contains(FullViewPaths, v.Pathname), "\nFileContent: "+v.FileContent, "")

		s := fmt.Sprint(indence, "\n Pathname", v.Pathname, "\nBulletDescription: ", v.BulletDescription, fileContent, SelfMemo, "\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a SolutionNodeList) SummaryView() string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Pathname, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		SelfMemo := lo.Ternary(v.SelfMemo != "", " SelfMemo: "+v.SelfMemo, "")

		s := fmt.Sprint(indence, "\n Pathname", v.Pathname, "\nBulletDescription: ", v.BulletDescription, SelfMemo, "\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a SolutionNodeList) PathnameSorted() SolutionNodeList {
	slices.SortFunc(a, func(a, b *SolutionNode) int {
		return strings.Compare(a.Pathname, b.Pathname)
	})
	return a
}

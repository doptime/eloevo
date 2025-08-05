package projects

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type SolutionGraphNodeList []*SolutionGraphNode

func (a SolutionGraphNodeList) Uniq() SolutionGraphNodeList { return lo.Uniq(a) }
func (a SolutionGraphNodeList) LockedOnly() SolutionGraphNodeList {
	return lo.Filter(a, func(v *SolutionGraphNode, _ int) bool { return v.Locked })
}
func (a SolutionGraphNodeList) FullView() string {
	return strings.Join(lo.Map(a, func(v *SolutionGraphNode, _ int) string { return v.String() }), "\n")
}
func (a SolutionGraphNodeList) SummaryView() string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Pathname, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		NextIterFeedback := lo.Ternary(v.NextIterFeedback != "", " NextIterFeedback: "+v.NextIterFeedback, "")

		s := fmt.Sprint(indence, "\nBulletDescription: ", v.BulletDescription, NextIterFeedback, " [Id:", v.Id, lo.Ternary(v.SuperEdge, " SuperEdge", ""), " importance:", strconv.Itoa(int(v.Importance)), " priority:", strconv.Itoa(int(v.Priority)), "]\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a SolutionGraphNodeList) SuerEdge() SolutionGraphNodeList {
	return lo.Filter(a, func(v *SolutionGraphNode, _ int) bool { return v.SuperEdge })
}
func (a SolutionGraphNodeList) Solution() SolutionGraphNodeList {
	return lo.Filter(a, func(v *SolutionGraphNode, _ int) bool { return !v.SuperEdge })
}
func (a SolutionGraphNodeList) PathnameSorted() SolutionGraphNodeList {
	slices.SortFunc(a, func(a, b *SolutionGraphNode) int {
		return strings.Compare(a.Pathname, b.Pathname)
	})
	return a
}

package batchop

import (
	"strings"

	"github.com/doptime/evolab/query"
	"github.com/doptime/evolab/tools"

	"github.com/atotto/clipboard"
)

func WithModel(Model string, node ...*query.Query) {
	for _, v := range node {
		v.Model = Model
	}
}

func CopyToClipboard(node ...*query.Query) {
	var stringBuilder strings.Builder
	for _, n := range node {
		stringBuilder.WriteString("\n\n# Stage: " + n.Group + " Model: " + n.Model + " Solution: \n\n")
		if n.MsgAssistant != "" {
			stringBuilder.WriteString(n.MsgAssistant)
		}
	}

	if s := stringBuilder.String(); len(s) > 0 {
		clipboard.WriteAll(s)
	}
}

func SelectBestNode(nodes ...*query.Query) (best *query.Query) {
	bestScore := float64(0)
	for _, v := range nodes {
		score, e := tools.ReadFloatAfterTag(v.MsgAssistant, "overall_score")
		if e == nil && score > bestScore {
			best = v
		}
	}
	return best
}

package projects

import (
	"strings"

	"github.com/doptime/eloevo/elo"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/samber/lo"
)

var ToolDroneBotSolutionItemRefine = tool.NewTool("SolutionItemRefine", "Propose/edit/delete solution item to improve solution.", func(newItem *BusinessPlans) {
	newItem.Item = strings.TrimSpace(newItem.Item)
	newItem.Importance = min(10, max(-1, newItem.Importance))
	newItem.Priority = min(10, max(0, newItem.Priority))
	var oItem *BusinessPlans = nil
	if newItem.Id != "" {
		oItem, _ = keyBusinessDronebot.HGet(newItem.Id)
	}
	if newItem.Id = utils.ID(newItem.Item, 4); oItem != nil {
		newItem.Id = oItem.Id
		oItem.Importance = newItem.Importance
		oItem.Priority = newItem.Priority
		if newItem.Item != "" && oItem.Item != newItem.Item {
			oItem.Item = newItem.Item
			if embed, err := utils.GetEmbedding(oItem.Item); err == nil {
				oItem.Embed(embed)
				milvusCollection.Upsert(oItem)
			}
		}
		oItem.SuperEdge = newItem.SuperEdge
		oItem.SuperEdgeNodes = lo.Ternary(len(newItem.SuperEdgeNodes) > 0, newItem.SuperEdgeNodes, oItem.SuperEdgeNodes)

	}
	if isNewModel := oItem == nil; isNewModel {
		if len(newItem.Item) > 0 && !utils.HasForbiddenWords(strings.ToLower(newItem.Item), ForbiddenWords) {
			keyBusinessDronebot.HSet(newItem.Id, newItem)
		}
		return
	}
	keyBusinessDronebot.HSet(oItem.Id, oItem)
})
var ToolDroneBotIterPlan = tool.NewTool("SuperEdgePlannedForNextLoop", "Propose super edge items in the next iter loop", func(edgeIds *SuperEdgePlannedForNextLoop) {
	if len(edgeIds.SuperEdgeIds) > 0 {
		keyIterPlannedDrone.RPush(edgeIds)
	}
})

type BatchEloResults struct {
	SortedItemIds []string                  `description:"Sorted item ids"`
	AllItems      map[string]*BusinessPlans `description:"-"`
}

var ToolDroneBatchEloResults = tool.NewTool("BatchEloResults", "Batch Elo Results. Represented By Item Ids. The better the item, the lower the index ", func(results *BatchEloResults) {

	Items := make([]elo.Elo, 0)
	for _, itemId := range results.SortedItemIds {
		item, ok := results.AllItems[itemId]
		if !ok {
			continue
		}
		if !item.Locked {
			Items = append(Items, item)
		}

	}
	if len(Items) > 0 {
		elo.BatchUpdateRanking(Items...)
	}
})

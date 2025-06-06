package projects

// var ToolDroneBotSolutionItemRefine = tool.NewTool("SolutionItemRefine", "Propose/edit/delete solution item to improve solution.", func(newItem *SolutionGraphNode) {
// 	newItem.BulletDescription= strings.TrimSpace(newItem.BulletName)
// 	newItem.Importance = min(10, max(-1, newItem.Importance))
// 	newItem.Priority = min(10, max(0, newItem.Priority))
// 	var oItem *SolutionGraphNode = nil
// 	if newItem.Id != "" {
// 		oItem, _ = KeyBusinessDronebot.HGet(newItem.Id)
// 	}
// 	if newItem.Id = utils.ID(newItem.BulletName, 4); oItem != nil {
// 		newItem.Id = oItem.Id
// 		oItem.Importance = newItem.Importance
// 		oItem.Priority = newItem.Priority
// 		oItem.Detail = lo.Ternary(newItem.Detail != "", newItem.Detail, oItem.Detail)
// 		if newItem.BulletDescription!= "" && oItem.BulletDescription!= newItem.BulletDescription{
// 			oItem.BulletDescription= newItem.BulletName
// 			if embed, err := utils.GetEmbedding(oItem.BulletName); err == nil {
// 				oItem.Embed(embed)
// 				milvusCollection.Upsert(oItem)
// 			}
// 		}
// 		oItem.SuperEdge = newItem.SuperEdge
// 		oItem.SuperEdgeNodes = lo.Ternary(len(newItem.SuperEdgeNodes) > 0, newItem.SuperEdgeNodes, oItem.SuperEdgeNodes)

// 	}
// 	if isNewModel := oItem == nil; isNewModel {
// 		if len(newItem.BulletName) > 0 && !utils.HasForbiddenWords(strings.ToLower(newItem.BulletName), ForbiddenWords) {
// 			KeyBusinessDronebot.HSet(newItem.Id, newItem)
// 		}
// 		return
// 	}
// 	KeyBusinessDronebot.HSet(oItem.Id, oItem)
// })
// var ToolDroneBotIterPlan = tool.NewTool("SuperEdgePlannedForNextLoop", "Propose super edge items in the next iter loop", func(edgeIds *SuperEdgePlannedForNextLoop) {
// 	if len(edgeIds.SuperEdgeIds) > 0 {
// 		keyIterPlannedDrone.RPush(edgeIds)
// 	}
// })

// type BatchEloResults struct {
// 	SortedItemIds []string                      `description:"Sorted item ids"`
// 	AllItems      map[string]*SolutionGraphNode `description:"-"`
// }

// var ToolDroneBatchEloResults = tool.NewTool("BatchEloResults", "Batch Elo Results. Represented By Item Ids. The better the item, the lower the index ", func(results *BatchEloResults) {

// 	// Items := make([]elo.Elo, 0)
// 	// for _, itemId := range results.SortedItemIds {
// 	// 	item, ok := results.AllItems[itemId]
// 	// 	if !ok {
// 	// 		continue
// 	// 	}
// 	// 	if !item.Locked {
// 	// 		Items = append(Items, item)
// 	// 	}

// 	// }
// 	// if len(Items) > 0 {
// 	// 	elo.BatchUpdateRanking(Items...)
// 	// }
// })

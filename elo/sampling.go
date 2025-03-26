package elo

import (
	"math"
	"math/rand/v2"

	"github.com/samber/lo"
)

// Sampling 改进后的采样函数，采用加权随机选择
func Sampling(players []Elo) (Elo, Elo) {
	if len(players) < 2 {
		return nil, nil // 至少需要两个选手参赛
	}

	// 计算每个玩家的权重
	var totalWeight float64
	weights := make([]float64, len(players))
	for i, player := range players {
		// 权重可以根据评分和比赛次数进行调整
		weight := float64(player.ScoreAccessor())
		weights[i] = weight
		totalWeight += weight
	}

	// 归一化权重
	for i := range weights {
		weights[i] /= totalWeight
	}

	// 加权随机选择第一个玩家
	player1Index := weightedRandomChoice(weights)

	// 选择第二个玩家，确保不与第一个玩家相同
	player2Index := player1Index
	for player2Index == player1Index {
		player2Index = weightedRandomChoice(weights)
	}

	return players[player1Index], players[player2Index]
}

// SamplingMap  从 map 中采样 N 个 Elo 玩家，采用加权随机选择
func SamplingMap(players map[string]Elo, N int) []Elo {
	if len(players) < N {
		return lo.Values(players)
	}
	selectedPlayers := make([]Elo, 0, N)     // 初始化用于存储选中玩家的切片，预分配容量为 N
	availablePlayers := make(map[string]Elo) // 创建一个可用的玩家 map，用于在选择后移除玩家，避免重复选择
	weightsMap := make(map[string]float64)   // 创建一个玩家权重 map，键为玩家的 key (string)

	var totalWeight float64 //  总权重

	//  1.  计算每个玩家的权重
	for key, player := range players {
		// 权重计算方式与 Sampling 函数保持一致，可以根据实际需求调整
		weight := math.Sqrt(math.Abs(float64(player.ScoreAccessor())))
		weightsMap[key] = weight
		totalWeight += weight
		availablePlayers[key] = player //  将所有玩家复制到 availablePlayers map
	}

	// 2.  归一化权重 (可选，但推荐，保证权重总和为 1，方便理解概率)
	for key := range weightsMap {
		weightsMap[key] /= totalWeight
	}

	//  3.  进行 N 次加权随机选择
	for i := 0; i < N; i++ {
		if len(availablePlayers) == 0 {
			break //  如果可用玩家为空，提前退出循环，防止无限循环 (理论上不会发生，除非 N 大于玩家总数，但前面已经做了数量检查)
		}

		currentWeights := make([]float64, 0, len(availablePlayers)) //  临时的权重切片，用于传递给 weightedRandomChoice 函数
		playerKeys := make([]string, 0, len(availablePlayers))      //  临时的玩家 key 切片，与权重切片一一对应

		//  从 availablePlayers map 中提取权重和 key，保持顺序一致
		for key, weight := range weightsMap {
			if _, ok := availablePlayers[key]; ok { //  只考虑仍然可用的玩家
				currentWeights = append(currentWeights, weight)
				playerKeys = append(playerKeys, key)
			}
		}

		//  使用 weightedRandomChoice 函数进行加权随机选择，返回的是权重切片的索引
		selectedIndex := weightedRandomChoice(currentWeights)
		selectedKey := playerKeys[selectedIndex]        //  根据索引获取选中的玩家 key
		selectedPlayer := availablePlayers[selectedKey] //  根据 key 从 availablePlayers map 中获取选中的玩家

		selectedPlayers = append(selectedPlayers, selectedPlayer) //  将选中的玩家添加到结果切片中
		delete(availablePlayers, selectedKey)                     //  从 availablePlayers map 中移除选中的玩家，确保不会重复选择
	}

	return selectedPlayers
}

// weightedRandomChoice 根据权重进行加权随机选择
func weightedRandomChoice(weights []float64) int {
	r := rand.Float64()
	var cumulativeWeight float64
	for i, weight := range weights {
		cumulativeWeight += weight
		if r <= cumulativeWeight {
			return i
		}
	}
	return len(weights) - 1
}

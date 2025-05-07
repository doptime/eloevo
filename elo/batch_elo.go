package elo

import (
	"math"

	"github.com/samber/lo"
)

// BatchUpdateWinnings updates Elo ratings for a list of winners and players
func BatchUpdateWinnings(winners []Elo, players []Elo, matchId string) {
	// Create a map for quick lookup of winners
	winnerMap := lo.SliceToMap(winners, func(m Elo) (string, struct{}) { return m.GetId(), struct{}{} })

	// Identify the losers
	losers := lo.Filter(players, func(m Elo, _ int) bool {
		_, exists := winnerMap[m.GetId()]
		return !exists
	})

	// No updates needed if there are no losers
	if len(losers) == 0 {
		return
	}

	// Initialize maps to store expected and actual scores
	expectedScores := make(map[string]float64, len(players))
	actualScores := make(map[string]float64, len(players))
	comparisonsNum := make(map[string]float64, len(players))

	// Calculate expected and actual scores for each player
	for _, winner := range winners {
		for _, loser := range losers {
			// Aggregate expected scores for both winner and loser
			expectedScores[winner.GetId()] += ExpectedScoreA(winner.ScoreAccessor(), loser.ScoreAccessor())
			expectedScores[loser.GetId()] += ExpectedScoreA(loser.ScoreAccessor(), winner.ScoreAccessor())
			comparisonsNum[winner.GetId()]++
			comparisonsNum[loser.GetId()]++
			// Set actual scores based on win/loss
			actualScores[winner.GetId()] += 1.0
		}
	}

	// Update ratings for each player
	for _, player := range players {
		expected, actual := expectedScores[player.GetId()], actualScores[player.GetId()]
		var numComparisons float64 = comparisonsNum[player.GetId()]
		// Avoid division by zero
		if numComparisons == 0 {
			continue
		}

		// Calculate rating change
		k := 20
		delta := ((actual - expected) / numComparisons) * float64(k)
		deltaInt := int(math.Round(delta))

		// Update the player's rating
		player.ScoreAccessor(deltaInt)
	}
}

const DefaultEloScore = 1000

// BatchUpdateRanking 更新 Elo 评分，playersRanked 按排名顺序排列，前面的玩家胜出
func BatchUpdateRanking(playersRanked ...Elo) {
	// 获取玩家数量
	numPlayers := len(playersRanked)

	// 如果没有玩家或只有一个玩家，无需更新
	if numPlayers < 2 {
		return
	}
	lo.ForEach(playersRanked, func(player Elo, _ int) {
		if player.ScoreAccessor() == 0 {
			player.ScoreAccessor(DefaultEloScore)
		}
	})

	// 计算每个玩家的预期得分和实际得分
	expectedScores := make(map[string]float64, numPlayers)
	actualScores := make(map[string]float64, numPlayers)
	comparisonsNum := make(map[string]float64, numPlayers)

	// 遍历每个玩家并计算其与其他玩家的预期得分
	for i := 0; i < numPlayers; i++ {
		for j := i + 1; j < numPlayers; j++ {
			// 玩家 i 胜出，玩家 j 败北
			winner, loser := playersRanked[i], playersRanked[j]
			if winner.GetId() == loser.GetId() {
				continue
			}

			// 计算预期得分
			expectedScores[winner.GetId()] += ExpectedScoreA(winner.ScoreAccessor(), loser.ScoreAccessor())
			expectedScores[loser.GetId()] += ExpectedScoreA(loser.ScoreAccessor(), winner.ScoreAccessor())
			comparisonsNum[winner.GetId()]++
			comparisonsNum[loser.GetId()]++

			// 记录实际得分
			actualScores[winner.GetId()] += 1.0
			//actualScores[loser.GetId()] += 0.0
		}
	}

	// 更新每个玩家的 Elo 评分
	for _, player := range playersRanked {
		expected := expectedScores[player.GetId()]
		actual := actualScores[player.GetId()]

		// 计算评级变化值
		k := 20
		delta := (actual - expected) / comparisonsNum[player.GetId()] * float64(k)
		deltaInt := int(math.Round(delta))

		// 更新玩家评分
		player.ScoreAccessor(deltaInt)
	}
}

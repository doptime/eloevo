package elo

import (
	"math"
)

// 外部定义的常数 D，偏差常量（通常为 400）
const D = 400
const BatchD = 400

// Elo 表示每个玩家的状态，包括评分、比赛次数和唯一标识符
type Elo interface {
	GetId() string
	ScoreAccessor(delta ...int) int
}

// ExpectedScore 计算玩家 A 相对于玩家 B 的预期胜率
func ExpectedScoreA(RatingA int, RatingB int) float64 {
	return 1 / (1 + math.Pow(10, float64(RatingA-RatingB)/float64(BatchD)))
}

// UpdateRatings 根据比赛结果更新两个玩家的评分
// scoreA 为玩家 A 的比赛结果：胜利为 1.0，平局为 0.5，失败为 0.0
func UpdateRatings(playerA, playerB Elo, matchId string, scoreA float64) {

	ExpectedScoreA := ExpectedScoreA(playerA.ScoreAccessor(), playerB.ScoreAccessor())
	ExpectedScoreB := 1 - ExpectedScoreA
	if scoreA == 0.5 {
		scoreA = 0.5
	}
	deltaA := BatchD * (scoreA - ExpectedScoreA)
	deltaB := BatchD * ((1 - scoreA) - ExpectedScoreB)
	playerA.ScoreAccessor(int(deltaA))
	playerB.ScoreAccessor(int(deltaB))

}

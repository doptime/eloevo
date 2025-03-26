package elo

import (
	"github.com/samber/lo"
)

type VotesSlice []Elo

func (elos VotesSlice) UpdateVotes(votesReceived map[string]int64) {
	_map := lo.SliceToMap(elos, func(v Elo) (string, Elo) {
		return v.GetId(), v
	})
	for k, v := range votesReceived {
		if item, ok := _map[k]; ok {
			item.ScoreAccessor(int(v))
		}
	}
}

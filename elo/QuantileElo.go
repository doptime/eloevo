package elo

import (
	"reflect"
	"slices"
)

type QuantileSlice []Quantile

// Quantile 表示每个玩家的状态，包括评分、比赛次数和唯一标识符
type Quantile interface {
	GetId() string
	Quantile(delta ...float64) float64
}

func UpdateQuantilTo(Score float64, playersRanked ...Quantile) {
	// 获取玩家数量
	numPlayers := len(playersRanked)

	// 如果没有玩家或只有一个玩家，无需更新
	if numPlayers < 2 {
		return
	}
	for i := 0; i < numPlayers; i++ {
		delta := Score - playersRanked[i].Quantile()
		playersRanked[i].Quantile(delta)
	}
}

// UpdateQuantile 更新 Quantile 评分，playersRanked 按排名顺序排列，前面的玩家胜出
func UpdateQuantile(playersRanked ...Quantile) {
	// 获取玩家数量
	numPlayers := len(playersRanked)

	// 如果没有玩家或只有一个玩家，无需更新
	if numPlayers < 2 {
		return
	}
	for i := 0; i < numPlayers; i++ {
		quantile := (float64(i) + 1) / float64(numPlayers)
		delta := quantile - playersRanked[i].Quantile()
		playersRanked[i].Quantile(delta)
	}
}

func (elos QuantileSlice) Sort() (ret QuantileSlice) {
	ret = append([]Quantile{}, elos...)
	slices.SortFunc(ret, func(i, j Quantile) int {
		if i.Quantile() == j.Quantile() {
			return 0
		}
		if i.Quantile() < j.Quantile() {
			return -1
		}
		return 1
	})
	return ret
}
func (elos QuantileSlice) Reverse() (ret QuantileSlice) {
	ret = make(QuantileSlice, len(elos))
	for i, ie := 0, len(elos); i < ie; i++ {
		ret[i] = elos[ie-i-1]
	}
	return ret
}
func (elos QuantileSlice) TakeByIds(ids ...string) QuantileSlice {
	eloMap := make(map[string]Quantile)
	for _, v := range elos {
		eloMap[v.GetId()] = v
	}
	var result QuantileSlice
	for _, id := range ids {
		if elo, ok := eloMap[id]; ok {
			result = append(result, elo)
		}
	}
	return result
}

func ToQuantileSlice(input interface{}) QuantileSlice {
	var result QuantileSlice
	// 获取输入的反射类型
	val := reflect.ValueOf(input)

	// 判断输入类型
	switch val.Kind() {
	case reflect.Slice:
		// 处理切片类型
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			if eloItem, ok := item.(Quantile); ok {
				result = append(result, eloItem)
			}
		}

	case reflect.Map:
		// 处理映射类型
		for _, key := range val.MapKeys() {
			item := val.MapIndex(key).Interface()
			if eloItem, ok := item.(Quantile); ok {
				result = append(result, eloItem)
			}
		}

	default:
		panic("unsupported type in elo.ToSlice")
	}

	return result
}

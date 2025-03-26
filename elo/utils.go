package elo

import (
	"reflect"
	"slices"
)

func SetTopNElosToLowestScore(Num int, elos ...Elo) {
	if len(elos) < Num {
		return
	}
	slices.SortFunc(elos, func(i, j Elo) int {
		return -(i.ScoreAccessor() - j.ScoreAccessor())
	})
	lowestScore := elos[Num-1].ScoreAccessor()
	for _, v := range elos {
		v.ScoreAccessor(lowestScore - v.ScoreAccessor())
	}
}

func Sorted(elos ...Elo) []Elo {
	slices.SortFunc(elos, func(i, j Elo) int {
		return i.ScoreAccessor() - j.ScoreAccessor()
	})
	return elos
}
func SortedReversed(elos ...Elo) []Elo {
	slices.SortFunc(elos, func(i, j Elo) int {
		return -(i.ScoreAccessor() - j.ScoreAccessor())
	})
	return elos
}

type EloSlice []Elo

func (elos EloSlice) Sort() (ret EloSlice) {
	ret = append([]Elo{}, elos...)
	slices.SortFunc(ret, func(i, j Elo) int {
		return i.ScoreAccessor() - j.ScoreAccessor()
	})
	return ret
}
func (elos EloSlice) Reverse() (ret EloSlice) {
	ret = make(EloSlice, len(elos))
	for i, ie := 0, len(elos); i < ie; i++ {
		ret[i] = elos[ie-i-1]
	}
	return ret
}
func (elos EloSlice) TakeByIds(ids ...string) EloSlice {
	eloMap := make(map[string]Elo)
	for _, v := range elos {
		eloMap[v.GetId()] = v
	}
	var result EloSlice
	for _, id := range ids {
		if elo, ok := eloMap[id]; ok {
			result = append(result, elo)
		}
	}
	return result
}

func ToSlice(input interface{}) EloSlice {
	var result EloSlice
	// 获取输入的反射类型
	val := reflect.ValueOf(input)

	// 判断输入类型
	switch val.Kind() {
	case reflect.Slice:
		// 处理切片类型
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			if eloItem, ok := item.(Elo); ok {
				result = append(result, eloItem)
			}
		}

	case reflect.Map:
		// 处理映射类型
		for _, key := range val.MapKeys() {
			item := val.MapIndex(key).Interface()
			if eloItem, ok := item.(Elo); ok {
				result = append(result, eloItem)
			}
		}

	default:
		panic("unsupported type in elo.ToSlice")
	}

	return result
}

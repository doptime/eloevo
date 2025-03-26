package mixincached

import (
	"github.com/doptime/redisdb"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type EloScore struct {
	Key        string
	cache      map[string]int64
	RedisKey   *redisdb.HashKey[string, int64]
	DefaultElo int64
}

var EloScoreMap = cmap.New[*EloScore]()

func WithElo(redisDBName string, key string, DefaultEloScore ...int64) (ret *EloScore) {
	if v, ok := EloScoreMap.Get(key); ok {
		return v
	}
	ret = &EloScore{Key: key + ":eloscore", cache: map[string]int64{}, DefaultElo: append(DefaultEloScore, 1000)[0]}
	ret.RedisKey = redisdb.NewHashKey[string, int64](redisdb.WithRds(redisDBName), redisdb.WithKey(ret.Key))
	ret.cache, _ = ret.RedisKey.HGetAll()
	EloScoreMap.Set(key, ret)
	return ret
}
func (es *EloScore) ScoreAccessor(Id string, delta float64) int {
	currentElo, ok := es.cache[Id]
	if !ok {
		currentElo = es.DefaultElo
		es.cache[Id] = es.DefaultElo
	}
	es.cache[Id] = currentElo + int64(delta)
	if delta != 0 || !ok {
		es.RedisKey.HSet(Id, es.cache[Id])
	}
	return int(es.cache[Id])
}

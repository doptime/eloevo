package mixincached

import (
	"github.com/doptime/redisdb"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type CtxStrings struct {
	Cache        map[string][]string
	RedisKey     *redisdb.HashKey[string, []string]
	ReserveNItem int
}

var CtxStringsMap = cmap.New[*CtxStrings]()

func HashKeyStrs(redisDBName string, key string) (ret *CtxStrings) {
	if v, ok := CtxStringsMap.Get(key); ok {
		return v
	}
	ret = &CtxStrings{Cache: map[string][]string{}}
	ret.RedisKey = redisdb.NewHashKey[string, []string](redisdb.Opt.Rds(redisDBName).Key(key))
	ret.Cache, _ = ret.RedisKey.HGetAll()
	ret.ReserveNItem = -1
	CtxStringsMap.Set(key, ret)
	return ret
}
func (es *CtxStrings) WithNItemOnly(n int) (ret *CtxStrings) {
	es.ReserveNItem = n
	return es
}
func (es *CtxStrings) GetInsert(Id string, NewFeedback ...string) []string {
	if _, ok := es.Cache[Id]; !ok {
		es.Cache[Id] = []string{}
	}
	if len(NewFeedback) > 0 {
		es.Cache[Id] = append(es.Cache[Id], NewFeedback...)
		if es.ReserveNItem > 0 && len(es.Cache[Id]) > es.ReserveNItem {
			es.Cache[Id] = es.Cache[Id][len(es.Cache[Id])-es.ReserveNItem:]
		}
		es.RedisKey.HSet(Id, es.Cache[Id])
	}
	return es.Cache[Id]
}

package tools

import (
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/redisdb"
)

type RedisHashKeyFieldValue struct {
	Key   string `json:"key,omitempty" description:"Key of redis hash"`
	Field string `json:"field,omitempty" description:"Field of redis hash"`
	Value string `json:"value,omitempty" description:"The content string to save"`
}

func SaveToRedisHashKey(param *RedisHashKeyFieldValue) {
	if param.Key == "" || param.Field == "" || param.Value == "" {
		return
	}
	redisdb.NewHashKey[string, string](redisdb.Opt.Key(param.Key)).HSet(param.Field, param.Value)

}

var SaveStringToRedisHashKey = tool.NewTool("RedisHashKeyFieldValue", "Save String to Redis Hash Key", SaveToRedisHashKey)

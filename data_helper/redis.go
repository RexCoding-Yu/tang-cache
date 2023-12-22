package data_helper

import (
	"TangCache/config"
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
)

type RedisPlugin struct {
	client           *redis.Client     // redis客户端
	ttl              uint64            // 过期时间
	keyPrefix        string            // KEY的前缀
	preloadScriptMap map[string]string // 存放预加载脚本的map
}

func (r *RedisPlugin) Init(conf *config.CacheConfig, prefix string) error {
	// 从配置中创建client或获取client
	if conf.RedisConfig.Options != nil {
		r.client = redis.NewClient(conf.RedisConfig.Options)
	} else if r.client = conf.RedisConfig.Client; r.client == nil {
		return errors.New("must provide ether redis config or client")
	}
	r.ttl = conf.CacheTTL
	r.keyPrefix = prefix
	r.preloadScriptMap = make(map[string]string)
	return r.initScripts()
}

// 预加载部分可能用的到的脚本到Redis
func (r *RedisPlugin) initScripts() error {

	if r.preloadScriptMap == nil {
		panic("preloadScriptMap init fail")
	}

	// 批量查询Key
	batchKeyExistScript := `
		for idx, val in pairs(KEYS) do
			local exists = redis.call('EXISTS', val)
			if exists == 0 then
				return 0
			end
		end
		return 1`

	// 批量删除键
	batchKeyCleanScript := `
		local cursor = "0"
		local count = 0
		repeat
		  local result = redis.call("SCAN", cursor, "MATCH", ARGV[1])
		  cursor = result[1]
		  local keys = result[2]
		  for i=1, #keys do
			redis.call("DEL", keys[i])
			count = count + 1
		  end
		until cursor == "0"
		return count`

	result := r.client.ScriptLoad(context.Background(), batchKeyExistScript)
	if result.Err() != nil {
		return result.Err()
	}
	r.preloadScriptMap["batchKeyExistScript"] = result.Val()

	result = r.client.ScriptLoad(context.Background(), batchKeyCleanScript)
	if result.Err() != nil {
		return result.Err()
	}
	r.preloadScriptMap["batchKeyCleanScript"] = result.Val()

	return nil
}

func (r *RedisPlugin) CleanCache(ctx context.Context) (int64, error) {
	result := r.client.EvalSha(ctx, r.preloadScriptMap["batchKeyCleanScript"], []string{}, r.keyPrefix+":*")
	return result.Int64()
}

func (r *RedisPlugin) BatchKeyExist(ctx context.Context, keys []string) (bool, error) {
	result := r.client.EvalSha(ctx, r.preloadScriptMap["batchKeyExistScript"], keys)
	return result.Bool()
}

func (r *RedisPlugin) KeyExist(ctx context.Context, key string) (int64, error) {
	result := r.client.Exists(ctx, key)
	return result.Result()
}

func (r *RedisPlugin) GetValue(ctx context.Context, key string) (string, error) {
	result := r.client.Get(ctx, key)
	return result.Result()
}

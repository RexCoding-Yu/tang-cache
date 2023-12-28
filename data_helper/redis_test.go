package data_helper

import (
	"TangCache/config"
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"testing"
)

func TestRedisPlugin_SetValue(t *testing.T) {
	type fields struct {
		client           *redis.Client
		ttl              uint64
		keyPrefix        string
		preloadScriptMap map[string]string
	}
	testRedisConfig := config.RedisConfig{
		Options: &redis.Options{Addr: "localhost:7890"},
	}
	testRedisConfig.Client = redis.NewClient(testRedisConfig.Options)
	type args struct {
		ctx   context.Context
		key   string
		value interface{}
	}
	config := config.CacheConfig{
		CacheLevel:   config.CacheLevelAll,
		CacheStorage: config.CacheStorageRedis,
		RedisConfig: &config.RedisConfig{
			Options: &redis.Options{Addr: "localhost:6379"},
		},
		CacheTTL: 60,
	}

	plugin := RedisPlugin{}
	err := plugin.Init(&config, "test")
	if err != nil {
		log.Fatal(err)
		return
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestInsert1",
			fields: fields{
				client:           plugin.client,
				ttl:              plugin.ttl,
				keyPrefix:        plugin.keyPrefix,
				preloadScriptMap: plugin.preloadScriptMap,
			},
			args: args{
				ctx:   context.Background(),
				key:   "test_insert",
				value: "hello_tang_cache",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RedisPlugin{
				client:           tt.fields.client,
				ttl:              tt.fields.ttl,
				keyPrefix:        tt.fields.keyPrefix,
				preloadScriptMap: tt.fields.preloadScriptMap,
			}
			if err := r.SetValue(tt.args.ctx, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SetValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package data_helper

import (
	"TangCache/config"
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"testing"
)

type testData struct {
	Id       uint
	UserName string
}

func TestRedisPlugin_SetValue(t *testing.T) {
	type fields struct {
		client           *redis.Client
		ttl              uint64
		keyPrefix        string
		preloadScriptMap map[string]string
	}
	td := testData{
		Id:       1,
		UserName: "rex",
	}
	testRedisConfig := config.RedisConfig{
		Options: &redis.Options{Addr: "localhost:6379", DB: 0},
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
		CacheTTL: 60 * 60,
	}

	plugin := RedisPlugin{}
	err := plugin.Init(&config, "test")
	if err != nil {
		log.Print(err)
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
				value: td,
			},
			wantErr: false,
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

func TestRedisPlugin_GetValue(t *testing.T) {
	type fields struct {
		client           *redis.Client
		ttl              uint64
		keyPrefix        string
		preloadScriptMap map[string]string
	}
	type args struct {
		ctx context.Context
		key string
		ptr interface{}
	}
	td := &testData{}
	testRedisConfig := config.RedisConfig{
		Options: &redis.Options{Addr: "localhost:6379", DB: 0},
	}
	testRedisConfig.Client = redis.NewClient(testRedisConfig.Options)
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
		log.Print(err)
		return
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestGet1",
			fields: fields{
				client:           plugin.client,
				ttl:              plugin.ttl,
				keyPrefix:        plugin.keyPrefix,
				preloadScriptMap: plugin.preloadScriptMap,
			},
			args: args{
				ctx: context.Background(),
				key: "test_insert",
				ptr: td,
			},
			wantErr: false,
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
			if err := r.GetValue(tt.args.ctx, tt.args.key, tt.args.ptr); (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			log.Print(tt.args.ptr)
		})
	}
}

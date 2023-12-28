package data_helper

import (
	"TangCache/config"
	"TangCache/util"
	"context"
	"reflect"
)

type TangCacheInterface interface {
	Init(conf *config.CacheConfig, prefix string) error
	InitScripts() error
	CleanCache(ctx context.Context) (int64, error)
	BatchKeyExist(ctx context.Context, keys []string) (bool, error)
	KeyExist(ctx context.Context, key string) (int64, error)
	GetValue(ctx context.Context, key string, ptr interface{}) error
	SetValue(ctx context.Context, key string, value interface{}) error
	BatchGetValue(ctx context.Context, keys []string, p reflect.Type) (interface{}, error)
	BatchSetValue(ctx context.Context, pairs []util.Pair) error
	DeleteKey(ctx context.Context, key string) (int64, error)
	BatchDeleteKeys(ctx context.Context, keys []string) (int64, error)
	DeleteKeysWithPrefix(ctx context.Context, keyPrefix string) (int64, error)
}

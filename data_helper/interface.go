package data_helper

import (
	"TangCache/util"
	"context"
	"reflect"
)

type TangCacheInterface interface {
	BatchKeyExist(ctx context.Context, keys []string) (bool, error)
	KeyExists(ctx context.Context, key string) (bool, error)
	GetValue(ctx context.Context, key string, ptr interface{}) error
	BatchGetValue(ctx context.Context, keys []string, p reflect.Type) (interface{}, error)

	CleanCache(ctx context.Context) error
	DeleteKeysWithPrefix(ctx context.Context, keyPrefix string) (int64, error)
	DeleteKey(ctx context.Context, key string) (int64, error)
	BatchDeleteKeys(ctx context.Context, keys []string) (int64, error)
	BatchSetValue(ctx context.Context, pairs []util.Pair) error
	SetValue(ctx context.Context, key string, value interface{}) error
}

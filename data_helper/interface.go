package data_helper

import (
	"TangCache/config"
	"TangCache/util"
	"context"
)

type TangCacheInterface interface {
	Init(config *config.CacheConfig, prefix string) error

	BatchKeyExist(ctx context.Context, keys []string) (bool, error)
	KeyExists(ctx context.Context, key string) (bool, error)
	GetValue(ctx context.Context, key string) (string, error)
	BatchGetValues(ctx context.Context, keys []string) ([]string, error)

	CleanCache(ctx context.Context) error
	DeleteKeysWithPrefix(ctx context.Context, keyPrefix string) error
	DeleteKey(ctx context.Context, key string) error
	BatchDeleteKeys(ctx context.Context, keys []string) error
	BatchSetKeys(ctx context.Context, sets []util.StringSet) error
	SetKey(ctx context.Context, set util.StringSet) error
}

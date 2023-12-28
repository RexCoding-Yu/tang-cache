package cache

import (
	"TangCache/config"
	"TangCache/data_helper"
	"TangCache/util"
	"context"
	"gorm.io/gorm"
	"log"
	"reflect"
)

type TangCache struct {
	Config     *config.CacheConfig
	InstanceId string

	db       *gorm.DB
	cache    data_helper.TangCacheInterface
	hitCount int64
}

func (c *TangCache) InvalidateSearchCache(ctx context.Context, tableName string) error {
	_, err := c.cache.DeleteKeysWithPrefix(ctx, util.GenSearchCachePrefix(c.InstanceId, tableName))
	return err
}

func (c *TangCache) SetQueryCache(ctx context.Context, tableName string, value string, sql string, args ...interface{}) {
	key := util.GenSearchCacheKey(c.InstanceId, tableName, sql, args...)
	err := c.cache.SetValue(ctx, key, value)
	if err != nil {
		log.Printf("[GetQueryCache:SetValue] error %v", err)
		return
	}
}

func (c *TangCache) GetQueryCache(ctx context.Context, tableName string, sql string, p reflect.Type, args ...interface{}) (interface{}, error) {
	key := util.GenSearchCacheKey(c.InstanceId, tableName, sql, args...)
	ptr := reflect.New(p)
	err := c.cache.GetValue(ctx, key, ptr)
	return ptr, err
}

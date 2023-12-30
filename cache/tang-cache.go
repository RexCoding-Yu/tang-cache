package cache

import (
	"TangCache/config"
	"TangCache/data_helper"
	"TangCache/util"
	"context"
	"fmt"
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

func NewTangCache(cacheConfig *config.CacheConfig) (*TangCache, error) {
	if cacheConfig == nil {
		return nil, fmt.Errorf("you pass a nil config")
	}
	cache := &TangCache{
		Config: cacheConfig,
	}
	err := cache.Init()
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func (c *TangCache) Name() string {
	return util.TangCachePrefix
}

func (c *TangCache) Initialize(db *gorm.DB) (err error) {
	err = db.Callback().Create().After("*").Register("gorm:cache:after_create", AfterCreate(c))
	if err != nil {
		return err
	}

	err = db.Callback().Delete().After("*").Register("gorm:cache:after_delete", AfterDelete(c))
	if err != nil {
		return err
	}

	err = db.Callback().Update().After("*").Register("gorm:cache:after_update", AfterUpdate(c))
	if err != nil {
		return err
	}

	err = db.Callback().Query().Before("gorm:query").Register("gorm:cache:before_query", BeforeQuery(c))
	if err != nil {
		return err
	}

	err = db.Callback().Query().After("*").Register("gorm:cache:after_query", AfterQuery(c))
	if err != nil {
		return err
	}

	return
}

func (c *TangCache) Init() error {
	if c.Config.CacheStorage == config.CacheStorageRedis {
		if c.Config.RedisConfig == nil {
			panic("please init redis config!")
		}
	}
	c.InstanceId = util.TangCachePrefix + ":" + util.GenInstanceId()

	prefix := c.InstanceId

	if c.Config.CacheStorage == config.CacheStorageRedis {
		c.cache = &data_helper.RedisPlugin{}
	}

	err := c.cache.Init(c.Config, prefix)
	if err != nil {
		log.Printf("[Init] cache init error: %v", err)
		return err
	}
	return nil
}

func (c *TangCache) InvalidateSearchCache(ctx context.Context, tableName string) error {
	_, err := c.cache.DeleteKeysWithPrefix(ctx, util.GenSearchCachePrefix(c.InstanceId, tableName))
	return err
}

func (c *TangCache) SetSearchCache(ctx context.Context, tableName string, value interface{}, sql string, args ...interface{}) error {
	key := util.GenSearchCacheKey(c.InstanceId, tableName, sql, args...)
	err := c.cache.SetValue(ctx, key, value)
	if err != nil {
		log.Printf("[GetQueryCache:SetValue] error %v", err)
		return err
	}
	return nil
}

func (c *TangCache) GetSearchCache(ctx context.Context, tableName string, sql string, p reflect.Type, args ...interface{}) (interface{}, error) {
	key := util.GenSearchCacheKey(c.InstanceId, tableName, sql, args...)
	ptr := reflect.New(p)
	err := c.cache.GetValue(ctx, key, ptr)
	return ptr, err
}

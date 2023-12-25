package cache

import (
	"TangCache/config"
	"TangCache/data_helper"
	"gorm.io/gorm"
)

type TangCache struct {
	Config     *config.CacheConfig
	InstanceId string

	db       *gorm.DB
	cache    data_helper.TangCacheInterface
	hitCount int64
}

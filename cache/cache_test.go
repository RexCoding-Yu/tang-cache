package cache

import (
	"TangCache/config"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestSecondLevelCache(t *testing.T) {
	dsn := "root:rex333153..@tcp(rex.fno.ink)/test_tang_cache?charset=utf8mb4&parseTime=True"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	redisOption := &redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	}
	redisConfig := &config.RedisConfig{
		Options: redisOption,
	}
	cache, _ := NewTangCache(&config.CacheConfig{
		CacheLevel:           config.CacheLevelAll,
		CacheStorage:         config.CacheStorageRedis,
		RedisConfig:          redisConfig,
		InvalidateWhenUpdate: true,
		CacheTTL:             60 * 60, // 60*60 s
		CacheMaxItemCnt:      0,
	})

	err := db.Use(cache)
	if err != nil {
		return
	}

	type User struct {
		gorm.Model
		UserName string `json:"user_name"`
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		return
	}
	var users []User
	// 不会命中缓存
	db.Where("user_name = ?", "tangchongjie").Find(&users)
	fmt.Print(users)
	// 命中缓存
	db.Where("user_name = ?", "tangchongjie").Find(&users)
	fmt.Print(users)
	// 命中缓存
	db.Where("user_name = ?", "tangchongjie").Find(&users)
	fmt.Print(users)
	// 命中缓存
	db.Where("user_name = ?", "tangchongjie").Find(&users)
	fmt.Print(users)
	// 命中缓存
	db.Where("user_name = ?", "tangchongjie").Find(&users)
	fmt.Print(users)
}

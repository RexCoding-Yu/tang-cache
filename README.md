# TangCache
一个Gorm的缓存插件，当前缓存颗粒度较大，后面会维护小颗粒度的版本。  
项目初衷是给[New-Api](https://github.com/Calcium-Ion/new-api)提供一个不错的缓存插件。  
使用方法：
```.go
func TestSecondLevelCache(t *testing.T) {
	dsn := "name:pwd@tcp(url)/test_tang_cache?charset=utf8mb4&parseTime=True"
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
}
```
下载方式
```.sh
go get github.com/RexCoding-Yu/tang-cache
```
项目名是为了纪念我的好友兼恩师-阿里架构师唐崇杰

package cache

import (
	"TangCache/config"
	"TangCache/util"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"log"
	"reflect"
	"sync"
)

func AfterCreate(cache *TangCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		} else {
			tableName = db.Statement.Table
		}
		ctx := db.Statement.Context

		if db.Error == nil && cache.Config.InvalidateWhenUpdate && util.ShouldCache(tableName, cache.Config.AllowTables) {
			if cache.Config.CacheLevel == config.CacheLevelAll || cache.Config.CacheLevel == config.CacheLevelOnlySearch {
				err := cache.InvalidateSearchCache(ctx, tableName)
				if err != nil {
					return
				}
			}
		}
	}
}

func BeforeQuery(cache *TangCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		callbacks.BuildQuerySQL(db)
		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		} else {
			tableName = db.Statement.Table
		}
		ctx := db.Statement.Context

		sql := db.Statement.SQL.String()
		db.InstanceSet("gorm:cache:sql", sql)
		db.InstanceSet("gorm:cache:vars", db.Statement.Vars)

		if util.ShouldCache(tableName, cache.Config.AllowTables) {

			if cache.Config.CacheLevel == config.CacheLevelAll || cache.Config.CacheLevel == config.CacheLevelOnlySearch {

				cacheValue, err := cache.GetSearchCache(ctx, tableName, sql, reflect.TypeOf(db.Statement.Model), db.Statement.Vars)
				if err != nil {
					if !errors.Is(err, redis.Nil) {
						log.Printf("[BeforeQuery] get cache value for sql %s error: %v", sql, err)
					}
					db.Error = nil
					return
				}
				log.Printf("[BeforeQuery] get value: %s", cacheValue)
				db.RowsAffected = 1
				db.Statement.Dest = cacheValue
				db.Error = util.SearchCacheHit
				return
			}
		}
	}
}

func AfterQuery(cache *TangCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		} else {
			tableName = db.Statement.Table
		}
		ctx := db.Statement.Context
		sqlObj, _ := db.InstanceGet("gorm:cache:sql")
		sql := sqlObj.(string)
		varObj, _ := db.InstanceGet("gorm:cache:vars")
		vars := varObj.([]interface{})

		if util.ShouldCache(tableName, cache.Config.AllowTables) {
			if db.Error == nil {
				// 如果没有这个错误，意味着缓存没有命中，需要添加缓存
				_, objects := getObjectsAfterLoad(db)

				var wg sync.WaitGroup
				// 计数器设为1
				wg.Add(1)

				go func() {
					defer wg.Done()

					if cache.Config.CacheLevel == config.CacheLevelAll || cache.Config.CacheLevel == config.CacheLevelOnlySearch {
						// 检验是否超出缓存数量
						if cache.Config.CacheMaxItemCnt > 0 && uint64(len(objects)) > cache.Config.CacheMaxItemCnt {
							return
						}

						log.Printf("[AfterQuery] start to set search cache for sql: %s", sql)
						// 用msgpack压缩
						cacheBytes, err := msgpack.Marshal(db.Statement.Dest)
						if err != nil {
							log.Printf("[AfterQuery] cannot marshal cache "+
								"for sql: %s, not cached", sql)
							return
						}
						log.Printf("[AfterQuery] set cache: %v", string(cacheBytes))
						// 尝试设置缓存
						err = cache.SetSearchCache(ctx, tableName, objects, sql, vars)
						if err != nil {
							log.Printf("[AfterQuery] set search cache for sql: %s error: %v", sql, err)
							return
						}
						log.Printf("[AfterQuery] sql %s cached", sql)
					}
				}()

				// 预留给以后的多种缓存形式使用
				wg.Wait()
				return
			}
		}

		if errors.Is(db.Error, util.SearchCacheHit) {
			// 命中搜索缓存
			db.Error = nil
			return
		}

		if errors.Is(db.Error, util.PrimaryCacheHit) {
			// 命中主键缓存
			db.Error = nil
			return
		}
	}
}

func AfterUpdate(cache *TangCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		} else {
			tableName = db.Statement.Table
		}
		ctx := db.Statement.Context

		if db.Error == nil && cache.Config.InvalidateWhenUpdate && util.ShouldCache(tableName, cache.Config.AllowTables) {
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()

				if cache.Config.CacheLevel == config.CacheLevelAll || cache.Config.CacheLevel == config.CacheLevelOnlySearch {
					log.Printf("[AfterUpdate] now start to invalidate search cache for table: %s", tableName)
					err := cache.InvalidateSearchCache(ctx, tableName)
					if err != nil {
						log.Printf("[AfterUpdate] invalidating search cache for table %s error: %v",
							tableName, err)
						return
					}
					log.Printf("[AfterUpdate] invalidating search cache for table: %s finished.", tableName)
				}
			}()

			wg.Wait()
		}
	}
}

func AfterDelete(cache *TangCache) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		} else {
			tableName = db.Statement.Table
		}
		ctx := db.Statement.Context

		if db.Error == nil && cache.Config.InvalidateWhenUpdate && util.ShouldCache(tableName, cache.Config.AllowTables) {
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()

				if cache.Config.CacheLevel == config.CacheLevelAll || cache.Config.CacheLevel == config.CacheLevelOnlySearch {
					log.Printf("[AfterUpdate] now start to invalidate search cache for table: %s", tableName)
					err := cache.InvalidateSearchCache(ctx, tableName)
					if err != nil {
						log.Printf("[AfterUpdate] invalidating search cache for table %s error: %v",
							tableName, err)
						return
					}
					log.Printf("[AfterUpdate] invalidating search cache for table: %s finished.", tableName)
				}
			}()

			wg.Wait()
		}
	}
}

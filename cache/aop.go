package cache

import (
	"TangCache/config"
	"TangCache/util"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"log"
	"reflect"
	"strings"
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

				cacheValue, err := cache.GetQueryCache(ctx, tableName, sql, reflect.TypeOf(db.Statement.Model), db.Statement.Vars)
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

			if cache.Config.CacheLevel == config.CacheLevelAll || cache.Config.CacheLevel == config.CacheLevelOnlyPrimary {
				primaryKeys := getPrimaryKeysFromWhereClause(db)
				cache.Logger.CtxInfo(ctx, "[BeforeQuery] parse primary keys = %v", primaryKeys)

				if len(primaryKeys) == 0 {
					return
				}

				// if (IN primaryKeys)/(Eq primaryKey) are the only clauses
				hasOtherClauseInWhere := hasOtherClauseExceptPrimaryField(db)
				if hasOtherClauseInWhere {
					// if query has other clauses, it can only query the database
					return
				}

				// primary cache hit
				cacheValues, err := cache.BatchGetPrimaryCache(ctx, tableName, primaryKeys)
				if err != nil {
					cache.Logger.CtxError(ctx, "[BeforeQuery] get primary cache value for key %v error: %v", primaryKeys, err)
					db.Error = nil
					return
				}
				if len(cacheValues) != len(primaryKeys) {
					db.Error = nil
					return
				}
				finalValue := ""

				destKind := reflect.Indirect(reflect.ValueOf(db.Statement.Dest)).Kind()
				if destKind == reflect.Struct && len(cacheValues) == 1 {
					finalValue = cacheValues[0]
				} else if (destKind == reflect.Array || destKind == reflect.Slice) && len(cacheValues) >= 1 {
					finalValue = "[" + strings.Join(cacheValues, ",") + "]"
				}
				if len(finalValue) == 0 {
					cache.Logger.CtxError(ctx, "[BeforeQuery] length of cache values and dest not matched")
					db.Error = util.ErrCacheUnmarshal
					return
				}

				err = json.Unmarshal([]byte(finalValue), db.Statement.Dest)
				if err != nil {
					cache.Logger.CtxError(ctx, "[BeforeQuery] unmarshal final value error: %v", err)
					db.Error = util.ErrCacheUnmarshal
					return
				}
				cache.IncrHitCount()
				db.Error = util.PrimaryCacheHit
				return
			}
		}
	}
}

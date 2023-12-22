package config

type CacheLevel int

const (
	CacheLevelOff         CacheLevel = 0 // 关闭缓存
	CacheLevelOnlyPrimary CacheLevel = 1 // 仅缓存主键
	CacheLevelOnlySearch  CacheLevel = 2 // 仅缓存搜索
	CacheLevelAll         CacheLevel = 3 // 全部缓存
)

type CacheStorage int

const (
	CacheStorageMemory CacheStorage = 0 // 内存缓存
	CacheStorageRedis  CacheStorage = 1 // Redis缓存
)

type CacheConfig struct {
	// CacheLevel 缓存级别
	CacheLevel CacheLevel

	// CacheStorage 缓存介质，内存或者Redis
	CacheStorage CacheStorage

	// RedisConfig 如果使用Redis缓存，需要配置Redis实例
	RedisConfig *RedisConfig

	// AllowTables 如果是空的，全部的表都会使用缓存。否则填入表名，只对对应的表使用缓存。
	AllowTables []string

	// InvalidateWhenUpdate 删、改的时候是否启用延迟双删
	InvalidateWhenUpdate bool

	// CacheTTL 缓存过期时间
	CacheTTL uint64

	// CacheMaxItemCnt 最大对象缓存数量，对于大数量的对象尽量限制。0是不限制。
	CacheMaxItemCnt uint64

	// CacheSize 仅对内存缓存生效，限制最大缓存数量。
	CacheSize int

	DebugMode bool
}

package util

import "errors"

var PrimaryCacheHit = errors.New("primary cache hit")
var SearchCacheHit = errors.New("search cache hit")

var ErrCacheUnmarshal = errors.New("cache hit, but unmarshal error")
var ErrCacheLoadFailed = errors.New("cache hit, but load value error")

type Pair struct {
	Key   string
	Value interface{}
}

const (
	TangCachePrefix = "TangCache"
)

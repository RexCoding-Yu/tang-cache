package data_helper

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/RexCoding-Yu/tang-cache/config"
	"github.com/RexCoding-Yu/tang-cache/util"
	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v5"
	"log"
	"reflect"
	"sync"
	"time"
)

const redisMaxLength int64 = 8 * 512 * 1024 * 1024 //512M

type RedisPlugin struct {
	client           *redis.Client     // redis客户端
	ttl              uint64            // 过期时间
	keyPrefix        string            // KEY的前缀
	preloadScriptMap map[string]string // 存放预加载脚本的map
}

func (r *RedisPlugin) Init(conf *config.CacheConfig, prefix string) error {
	// 从配置中创建client或获取client
	if conf.RedisConfig.Options != nil {
		r.client = redis.NewClient(conf.RedisConfig.Options)
	} else if r.client = conf.RedisConfig.Client; r.client == nil {
		return errors.New("must provide ether redis config or client")
	}
	r.ttl = conf.CacheTTL
	r.keyPrefix = prefix
	r.preloadScriptMap = make(map[string]string)
	return r.InitScripts()
}

// InitScripts 预加载部分可能用的到的脚本到Redis
func (r *RedisPlugin) InitScripts() error {

	if r.preloadScriptMap == nil {
		panic("preloadScriptMap init fail")
	}

	// 批量查询Key
	batchKeyExistScript := `
		for idx, val in pairs(KEYS) do
			local exists = redis.call('EXISTS', val)
			if exists == 0 then
				return 0
			end
		end
		return 1`

	// 批量删除键
	batchKeyCleanScript := `
		local cursor = "0"
		local count = 0
		repeat
		  local result = redis.call("SCAN", cursor, "MATCH", ARGV[1])
		  cursor = result[1]
		  local keys = result[2]
		  for i=1, #keys do
			redis.call("DEL", keys[i])
			count = count + 1
		  end
		until cursor == "0"
		return count`

	result := r.client.ScriptLoad(context.Background(), batchKeyExistScript)
	if result.Err() != nil {
		return result.Err()
	}
	r.preloadScriptMap["batchKeyExistScript"] = result.Val()

	result = r.client.ScriptLoad(context.Background(), batchKeyCleanScript)
	if result.Err() != nil {
		return result.Err()
	}
	r.preloadScriptMap["batchKeyCleanScript"] = result.Val()

	return nil
}

// CleanCache 清空缓存
func (r *RedisPlugin) CleanCache(ctx context.Context) (int64, error) {
	result := r.client.EvalSha(ctx, r.preloadScriptMap["batchKeyCleanScript"], []string{}, r.keyPrefix+":*")
	return result.Int64()
}

// BatchKeyExist 批量判断Key是否存在于缓存中
func (r *RedisPlugin) BatchKeyExist(ctx context.Context, keys []string) (bool, error) {
	result := r.client.EvalSha(ctx, r.preloadScriptMap["batchKeyExistScript"], keys)
	return result.Bool()
}

// KeyExist 判断一个Key是否存在
func (r *RedisPlugin) KeyExist(ctx context.Context, key string) (int64, error) {
	result := r.client.Exists(ctx, key)
	return result.Result()
}

// GetValue 通过Key获取Value
func (r *RedisPlugin) GetValue(ctx context.Context, key string, ptr interface{}) error {
	value := r.client.Get(ctx, key)
	valueBytes, err := value.Bytes()
	if err != nil {
		log.Print(err)
		return err
	}
	err = msgpack.Unmarshal(valueBytes, ptr)
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

// SetValue 通过Key设置Value
func (r *RedisPlugin) SetValue(ctx context.Context, key string, value interface{}) error {
	packValue, _ := msgpack.Marshal(value)
	_, err := r.client.Set(ctx, key, packValue, time.Duration(float64(r.ttl)*float64(time.Second)*util.RandFloat64())).Result()
	return err
}

// BatchGetValue 批量获取缓存值
func (r *RedisPlugin) BatchGetValue(ctx context.Context, keys []string, p reflect.Type) (interface{}, error) {
	results := r.client.MGet(ctx, keys...)
	err := results.Err()
	val := results.Val()
	if err != nil {
		return nil, err
	}

	// 反射传入类型，创建切片
	slice := reflect.MakeSlice(reflect.SliceOf(p), len(val), len(val))

	// 填充切片
	for i, v := range val {
		// 从Redis获取的数据是字符串，需要转换为字节切片；断言value是字符串类型，并且转换
		data := []byte(v.(string))
		// 创建一个buffer用于读取数据
		buffer := bytes.NewBuffer(data)
		// 创建一个gob解码器
		dec := gob.NewDecoder(buffer)
		// 创建一个新的p类型的实例
		value := reflect.New(p).Interface()
		// 使用解码器将数据解码
		if err := dec.Decode(value); err != nil {
			return nil, err
		}
		// 将反序列化后的值存储到切片中
		slice.Index(i).Set(reflect.ValueOf(value).Elem())
	}

	// 将切片作为接口返回
	return slice.Interface(), nil
}

// BatchSetValue 批量插入
func (r *RedisPlugin) BatchSetValue(ctx context.Context, pairs []util.Pair) error {
	for _, obj := range pairs {
		err := r.SetValue(ctx, obj.Key, obj.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteKey 删除Key对应的缓存
func (r *RedisPlugin) DeleteKey(ctx context.Context, key string) (int64, error) {
	results := r.client.Del(ctx, key)
	return results.Result()
}

// BatchDeleteKeys 批量删除Key对应的缓存
func (r *RedisPlugin) BatchDeleteKeys(ctx context.Context, keys []string) (int64, error) {
	results := r.client.Del(ctx, keys...)
	return results.Result()
}

// DeleteKeysWithPrefix 通过前缀删除对应的缓存
func (r *RedisPlugin) DeleteKeysWithPrefix(ctx context.Context, keyPrefix string) (int64, error) {
	results := r.client.EvalSha(ctx, r.preloadScriptMap["batchKeyCleanScript"], []string{"0"}, keyPrefix+":*")
	return results.Int64()
}

// SetBitValue 布隆过滤器-设置Bit
func (r *RedisPlugin) SetBitValue(ctx context.Context, offsets []int64) error {
	var wg sync.WaitGroup
	doneChan := make(chan struct{}, 1)
	errChan := make(chan error, 1)
	for _, off := range offsets {
		wg.Add(1)
		go func(offset int64) {
			defer wg.Done()
			key, thisOffset := r.getKeyOffset(offset)
			_, err := r.client.SetBit(ctx, key, thisOffset, 1).Result()
			if err != nil {
				errChan <- err
				return
			}
		}(off)
	}

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		return nil
	case err := <-errChan:
		return err
	}
}

// Test 布隆过滤器-测试
func (r *RedisPlugin) Test(ctx context.Context, offsets []int64) (bool, error) {
	for _, offset := range offsets {
		key, thisOffset := r.getKeyOffset(offset)
		bitValue, err := r.client.GetBit(ctx, key, thisOffset).Result()
		if err != nil {
			return false, err
		}
		if bitValue == 0 {
			return false, nil
		}
	}
	return true, nil
}

func (r *RedisPlugin) getKeyOffset(offset int64) (string, int64) {
	index := int64(offset / redisMaxLength)
	thisOffset := offset - index*redisMaxLength
	key := fmt.Sprintf("%s:%d", r.keyPrefix, index)
	return key, thisOffset
}

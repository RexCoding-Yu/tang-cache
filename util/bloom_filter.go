package util

import (
	"hash/fnv"
	"math"
)

type BitSetProvider interface {
	Set([]int64) error
	Test([]int64) (bool, error)
}

type BloomFilter struct {
	m      int64 // the size(bit) for the BloomFilter
	k      int64 // the hash function count
	bitSet BitSetProvider
}

func New(m int64, k int64, bitSet BitSetProvider) *BloomFilter {
	return &BloomFilter{m: m, k: k, bitSet: bitSet}
}

// EstimateParameters 根据单位个数和错误率估算所需内存和哈希函数数量
// Input: n: number of items, p: the err_rate
// Output: m: the total Size(bit), k: the hash function number.
// https://krisives.github.io/bloom-calculator/
func EstimateParameters(n uint, p float64) (int64, int64) {
	m := math.Ceil(float64(n) * math.Log(p) / math.Log(1.0/math.Pow(2.0, math.Ln2)))
	k := math.Ln2*m/float64(n) + 0.5

	return int64(m), int64(k)
}

func (f *BloomFilter) Add(data []byte) error {
	locations := f.getLocations(data)
	err := f.bitSet.Set(locations)
	if err != nil {
		return err
	}
	return nil
}

func (f *BloomFilter) Exists(data []byte) (bool, error) {
	locations := f.getLocations(data)
	isSet, err := f.bitSet.Test(locations)
	if err != nil {
		return false, err
	}
	if !isSet {
		return false, nil
	}
	return true, nil
}

func (f *BloomFilter) getLocations(data []byte) []int64 {
	// 设置一个最终定位的切片
	locations := make([]int64, f.k)
	// 初始化哈希器(Fowler-Noll-Vo)
	hashTool := fnv.New64()
	hashTool.Write(data)
	a := make([]byte, 1)
	for i := int64(0); i < f.k; i++ {
		a[0] = byte(i)
		// 改变哈希器状态，让其可以使用同一个哈希函数生成多个哈希值
		hashTool.Write(a)
		// 计算当前哈希器状态的64位的值
		hashValue := hashTool.Sum64()
		locations[i] = int64(hashValue % uint64(f.m))
	}
	return locations
}

func Set(offset []int64) error {

	return nil
}

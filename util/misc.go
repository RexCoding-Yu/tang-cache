package util

import (
	"math/rand"
	"time"
)

func ShouldCache(tableName string, tables []string) bool {
	if len(tables) == 0 {
		return true
	}
	return ContainString(tableName, tables)
}

func ContainString(target string, slice []string) bool {
	for _, s := range slice {
		if target == s {
			return true
		}
	}
	return false
}

// RandFloat64 生成一个0.95到1.05的数字
func RandFloat64() float64 {
	// 初始化随机数生成器的种子
	rand.Seed(time.Now().UnixNano())

	// 生成一个 0 到 1 之间的随机浮点数
	randomFloat := rand.Float64()

	// 缩放到 0.1 的范围并平移到 0.95 到 1.05 之间
	randomFloat = randomFloat*0.1 + 0.95
	return randomFloat
}

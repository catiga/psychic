package util

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

var myCache = cache.New(10*time.Minute, 10*time.Minute) // 设置默认失效时间为5分钟，清理间隔为10分钟

func CachePut(key string, val interface{}, invalid time.Duration) error {
	if len(key) == 0 {
		return errors.New("key empty")
	}

	myCache.Set(key, val, invalid)
	return nil
}

func CacheGet(key string) (interface{}, bool) {
	val, ok := myCache.Get(key)
	if !ok {
		return nil, false
	}

	// 这里不需要手动检查缓存是否过期，go-cache会自动处理过期的缓存项

	return val, true
}

func CacheDel(key string) {
	if len(key) == 0 {
		return
	}

	myCache.Delete(key)
}

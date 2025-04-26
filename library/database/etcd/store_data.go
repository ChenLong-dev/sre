package etcd

import (
	"context"
	"errors"
	"sync"
)

// 存储信息结构体
type StoreData struct {
	// 实际存储的map
	data *sync.Map
	// 值过滤数组
	valueFilter []string
	// context的cancel方法
	cancel context.CancelFunc
	// 是否开启监听
	enableWatch bool
}

// 获取指定键的值
func (s *StoreData) Get(key string) (string, error) {
	value, ok := s.data.Load(key)
	if !ok {
		return "", errors.New("target key is not set")
	}

	return value.(string), nil
}

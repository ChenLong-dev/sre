package agollo

import (
	"github.com/pkg/errors"
	"sync"
)

// 存储信息结构体
type StoreData struct {
	// 存储map
	data *sync.Map
}

// 获取指定键的值
func (cd *StoreData) Load(key string) (interface{}, error) {
	value, ok := cd.data.Load(key)
	if ok {
		return value, nil
	}

	return nil, errors.New("target key is not set")
}

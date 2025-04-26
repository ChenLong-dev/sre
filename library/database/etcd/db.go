package etcd

import (
	"context"
	"strings"
	"sync"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/base/runtime"
	"go.etcd.io/etcd/clientv3"
)

// 数据库结构体
type DB struct {
	// 客户端
	client *clientv3.Client
	// 键值对客户端
	kvClient clientv3.KV
	// 存储的map，key为前缀
	storeMap map[string]*StoreData
	// 钩子管理器
	manager *hook.Manager
	// 配置文件
	config *Config
	// 会话是否关闭
	done chan bool
}

// 检查是否指定键值是否被过滤，并返回过滤后的值
func (db *DB) filterPrefixKeyAndValue(prefix string, valueFilter []string, originKey, originValue []byte) (
	isFilter bool, key, value string) {
	value = string(originValue)
	key = string(originKey)
	for _, filterValue := range valueFilter {
		if strings.HasPrefix(value, filterValue) {
			return true, "", ""
		}
	}

	return false, strings.TrimPrefix(string(originKey), prefix), value
}

func (db *DB) after(ctx context.Context, prefix, key, value, extra string) {
	hk := db.manager.CreateHook(ctx).
		AddArg(render.SourceArgKey, runtime.GetDefaultFilterCallers()).
		AddArg("prefix", prefix).
		AddArg("key", key).
		AddArg("extra", extra)
	if db.config.DataValueOut {
		hk.AddArg("value", value)
	}
	hk.ProcessAfterHook()
}

// 重设指定前缀的数据
func (db *DB) ResetPrefixData(prefix string) error {
	data, ok := db.storeMap[prefix]
	if !ok {
		return errors.New("target is not set")
	}

	data.data = new(sync.Map)
	data.cancel()
	data.cancel = nil

	return nil
}

// 加载指定前缀的数据
func (db *DB) loadPrefixData(parentCtx context.Context, prefix string, valueFilter []string, enableWatch bool) (*StoreData, error) {
	if prefix == "" {
		return nil, errors.New("prefix is empty")
	}
	ctx, cancel := context.WithCancel(parentCtx)

	defaultResp, err := db.kvClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	dataMap := new(sync.Map)
	for _, item := range defaultResp.Kvs {
		isFilter, key, value := db.filterPrefixKeyAndValue(prefix, valueFilter, item.Key, item.Value)
		if isFilter {
			continue
		}
		dataMap.Store(key, value)
		db.after(ctx, prefix, string(item.Key), string(item.Value), "")
	}

	storeData := &StoreData{
		valueFilter: valueFilter,
		data:        dataMap,
		cancel:      cancel,
		enableWatch: enableWatch,
	}
	db.storeMap[prefix] = storeData

	if enableWatch {
		go func() {
			watchChan := db.client.Watch(ctx, prefix, clientv3.WithPrefix())
			for {
				select {
				case <-db.done:
					return
				case watchResp := <-watchChan:
					if watchResp.Err() != nil {
						break
					}
					// 如果watch被关闭，会返回一个空事件的响应
					// 这时候需要判断，并退出
					if len(watchResp.Events) == 0 {
						return
					}

					for _, e := range watchResp.Events {
						if e.Type != mvccpb.PUT {
							continue
						}

						isFilter, key, value := db.filterPrefixKeyAndValue(prefix, valueFilter, e.Kv.Key, e.Kv.Value)
						if isFilter {
							continue
						}

						dataMap.Store(key, value)
						db.after(ctx, prefix, string(e.Kv.Key), string(e.Kv.Value), "watch")
					}
				}
			}
		}()
	}

	return storeData, nil
}

// 强制加载指定前缀数据
// 会强制重设数据
func (db *DB) ForceLoadPrefixData(parentCtx context.Context, prefix string, valueFilter []string, enableWatch bool) (*StoreData, error) {
	_ = db.ResetPrefixData(prefix)

	return db.loadPrefixData(parentCtx, prefix, valueFilter, enableWatch)
}

// 重新加载指定前缀数据
// 会重设数据，如果没有加载，会返回错误
func (db *DB) ReloadPrefixData(parentCtx context.Context, prefix string, valueFilter []string, enableWatch bool) (*StoreData, error) {
	err := db.ResetPrefixData(prefix)
	if err != nil {
		return nil, err
	}

	return db.loadPrefixData(parentCtx, prefix, valueFilter, enableWatch)
}

// 加载指定前缀数据
// 如果已加载，会返回错误
func (db *DB) LoadPrefixData(parentCtx context.Context, prefix string, valueFilter []string, enableWatch bool) (*StoreData, error) {
	_, ok := db.storeMap[prefix]
	if ok {
		return nil, errors.New("target prefix is loaded")
	}

	return db.loadPrefixData(parentCtx, prefix, valueFilter, enableWatch)
}

// 关闭数据库
func (db *DB) Close() {
	close(db.done)

	for key, item := range db.storeMap {
		item.cancel()
		item.data = nil
		delete(db.storeMap, key)
	}
}

// 获取指定前缀的map
func (db *DB) GetPrefix(prefix string) (*StoreData, error) {
	s, ok := db.storeMap[prefix]
	if !ok {
		return nil, errors.New("prefix not load")
	}

	return s, nil
}

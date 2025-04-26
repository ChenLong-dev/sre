package agollo

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/shima-park/agollo"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/base/runtime"
)

// Apollo客户端
type Client struct {
	// Apollo的方法接口，用于启动watch监听
	client agollo.Agollo
	// 存储的map，key为namespace
	storeMap map[string]*StoreData
	// watcher切片
	watchers []Watcher
	// 会话是否关闭
	done chan bool
	// 钩子管理器
	manager *hook.Manager
	// 配置信息
	config *Config
}

// 新建Apollo客户端后台watch
func NewClient(conf *Config) *Client {
	configIdentify(conf)
	if conf.Config == nil {
		conf.Config = &render.Config{}
	}
	if conf.Config.StdoutPattern == "" {
		conf.Config.StdoutPattern = defaultPattern
	}
	if conf.Config.OutPattern == "" {
		conf.Config.OutPattern = defaultPattern
	}
	if conf.Config.OutFile == "" {
		conf.Config.OutFile = _infoFile
	}

	client, err := agollo.New(
		conf.ServerHost,
		conf.AppID,
		agollo.Cluster(conf.Cluster),
		agollo.PreloadNamespaces(conf.PreloadNamespaces...),
		agollo.AutoFetchOnCacheMiss(),
	)
	if err != nil {
		panic(err)
	}

	c := &Client{
		client:  client,
		config:  conf,
		manager: NewHookManager(conf.Config),
		done:    make(chan bool),
	}

	c.storeMap = make(map[string]*StoreData)
	if !conf.NotDaemon {
		daemonWatcher := newDaemonWatcher(c, conf.PreloadNamespaces)
		c.watchers = append(c.watchers, daemonWatcher)
		daemonWatcher.Watch(conf.PreloadNamespaces)
	}

	return c
}

// 监听namespace变化
func (c *Client) Watch(namespaces []string) (<-chan *ApolloResponse, error) {
	if len(c.config.PreloadNamespaces) == 0 {
		return nil, errors.New("preload namespace must be set")
	}
	if !c.config.NotDaemon {
		return nil, errors.New("your apollo config is daemon pattern, you dont need to watch")
	}
	if len(namespaces) == 0 {
		return nil, errors.New("watch namespace is empty")
	}

	simpleWatcher := newSimpleWatcher(c)
	c.watchers = append(c.watchers, simpleWatcher)
	simpleWatcher.Watch(namespaces)

	return simpleWatcher.Event, nil
}

// 获取指定namespace配置
func (c *Client) Get(namespace string) map[string]interface{} {
	return c.client.GetNameSpace(namespace)
}

// 关闭Apollo客户端
// TODO: 这里需要等待90秒才能释放所有监听的goroutine，这是Apollo客户端的设计
func (c *Client) Close() error {
	for _, w := range c.watchers {
		w.Close()
	}

	for key, item := range c.storeMap {
		item.data = nil
		delete(c.storeMap, key)
	}

	close(c.done)
	c.manager.Close()
	c.client.Stop()

	return nil
}

// 获取namespace下所有配置
func (c *Client) GetNamespace(namespace string) (*StoreData, error) {
	if len(c.config.PreloadNamespaces) == 0 {
		return nil, errors.New("preload namespace must be set")
	}
	if c.config.NotDaemon {
		return nil, errors.New("your apollo config is not daemon pattern")
	}

	if value, ok := c.storeMap[namespace]; ok {
		return value, nil
	} else {
		return nil, errors.New("target namespace is not preload")
	}
}

// 获取配置
func (c *Client) GetNewValue(namespace, key string) (interface{}, error) {
	if len(c.config.PreloadNamespaces) == 0 {
		return nil, errors.New("preload namespace must be set")
	}
	if c.config.NotDaemon {
		return nil, errors.New("your apollo config is not daemon pattern")
	}

	cd, ok := c.storeMap[namespace]
	if !ok {
		return nil, errors.New("namespace is not preload")
	}

	value, err := cd.Load(key)
	if err != nil {
		return nil, fmt.Errorf("getNewValue err: %v", err)
	}

	return value, nil
}

// 判断client是否close
func (c *Client) Done() <-chan bool {
	return c.done
}

// 日志打印
func (c *Client) log(msg *logInfo) {
	c.manager.CreateHook(context.Background()).
		AddArg("appID", c.config.AppID).
		AddArg("cluster", c.config.Cluster).
		AddArg("namespace", msg.NamespaceName).
		AddArg("ip", c.config.ServerHost).
		AddArg("extra", msg.extra).
		AddArg(render.SourceArgKey, runtime.GetDefaultFilterCallers()).
		AddArg("changes", msg.changes).
		AddArg("watcherType", msg.watcherType).
		ProcessAfterHook()
}

// 存apollo数据
func (c *Client) Store(namespace string, newDataMap map[string]interface{}) {
	dataMap, ok := c.storeMap[namespace]
	if !ok {
		dataMap = &StoreData{
			data: new(sync.Map),
		}
	}
	for k, v := range newDataMap {
		dataMap.data.Store(k, v)
	}

	c.storeMap[namespace] = dataMap
}

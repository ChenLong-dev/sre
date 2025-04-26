package agollo

import (
	"fmt"

	"github.com/shima-park/agollo"
	"gitlab.shanhai.int/sre/library/base/slice"
)

const (
	// 监听事件channel大小
	SimpleEventChanCount = 5
)

type EventType int

const (
	// 错误事件
	EventError EventType = iota
	// 更新事件
	EventUpdate
)

type WatcherType string

const (
	// 同步监听
	SimpleWatcher WatcherType = "SimpleWatcher"
	// 异步监听
	DaemonWatcher WatcherType = "DaemonWatcher"
)

type Watcher interface {
	// 监听方法
	Watch(namespaces []string)
	// 关闭watcher
	Close()
	// 获取watcher类型
	Type() WatcherType
}

// 变更响应
type ApolloResponse struct {
	// 变更的namespace
	Namespace string
	// 数据变更类型
	Type EventType
	// 旧值
	OldValue map[string]interface{}
	// 新值
	NewValue map[string]interface{}
	// 错误信息
	ErrInfo *agollo.LongPollerError
}

// simpleWatch监听器
type simpleWatcher struct {
	// watcher名称
	WatcherType WatcherType
	// watch的namespace
	Namespaces []string
	// watcher的事件
	Event chan *ApolloResponse
	// watch到的错误channel
	errChan <-chan *agollo.LongPollerError
	// watch到的配置变更详情
	watchResp <-chan *agollo.ApolloResponse
	// Apollo客户端
	client *Client
}

// 新建simpleWatcher
func newSimpleWatcher(c *Client) *simpleWatcher {
	return &simpleWatcher{
		errChan:   c.client.Start(),
		watchResp: c.client.Watch(),
		client:    c,
		Event:     make(chan *ApolloResponse, SimpleEventChanCount),
	}
}

// simpleWatcher监听
func (sw *simpleWatcher) Watch(namespaces []string) {
	sw.Namespaces = namespaces
	go func() {
		for {
			select {
			case err := <-sw.errChan:
				sw.client.log(&logInfo{
					NamespaceName: err.Namespace,
					extra:         fmt.Sprintf("Watch err: %s", err.Err.Error()),
					changes:       "",
					watcherType:   sw.Type(),
				})
				sw.broadcast(&ApolloResponse{
					ErrInfo: err,
					Type:    EventError,
				})
			case watchResp := <-sw.watchResp:
				if watchResp.Error != nil {
					sw.client.log(&logInfo{
						NamespaceName: watchResp.Namespace,
						extra:         fmt.Sprintf("Watch err: %v", watchResp.Error),
						changes:       "",
						watcherType:   sw.Type(),
					})
					break
				}
				sw.client.log(&logInfo{
					NamespaceName: watchResp.Namespace,
					extra:         fmt.Sprintf("Watch config changes"),
					changes:       fmt.Sprintf("diff: %s", watchResp.OldValue.Different(watchResp.NewValue)),
					watcherType:   sw.Type(),
				})
				sw.broadcast(&ApolloResponse{
					Namespace: watchResp.Namespace,
					Type:      EventUpdate,
					OldValue:  watchResp.OldValue,
					NewValue:  watchResp.NewValue,
					ErrInfo:   nil,
				})
			case <-sw.client.Done():
				return
			}
		}
	}()
}

// 关闭watcher
func (sw *simpleWatcher) Close() {
	close(sw.Event)
}

// 获取watcher类型
func (sw *simpleWatcher) Type() WatcherType {
	return SimpleWatcher
}

// 事件发event channel
func (sw *simpleWatcher) broadcast(resp *ApolloResponse) {
	if slice.StrSliceContains(sw.Namespaces, resp.Namespace) {
		sw.Event <- resp
	}
}

// daemonWatch监听器
type daemonWatcher struct {
	// watcher名称
	WatcherType WatcherType
	// watch的namespace
	Namespaces []string
	// watch到的错误channel
	errChan <-chan *agollo.LongPollerError
	// watch到的配置变更详情
	watchResp <-chan *agollo.ApolloResponse
	// Apollo客户端
	client *Client
}

// 新建daemonWatcher
func newDaemonWatcher(c *Client, namespaces []string) Watcher {
	for _, ns := range namespaces {
		c.Store(ns, c.Get(ns))
	}
	return &daemonWatcher{
		errChan:    c.client.Start(),
		watchResp:  c.client.Watch(),
		client:     c,
		Namespaces: namespaces,
	}
}

// daemonWatcher监听
func (dw *daemonWatcher) Watch(namespaces []string) {
	go func() {
		for {
			select {
			case err := <-dw.errChan:
				dw.client.log(&logInfo{
					NamespaceName: err.Namespace,
					extra:         fmt.Sprintf("Watch err: %s", err.Err.Error()),
					changes:       "",
					watcherType:   dw.Type(),
				})
			case watchResp := <-dw.watchResp:
				if watchResp.Error != nil {
					dw.client.log(&logInfo{
						NamespaceName: watchResp.Namespace,
						extra:         fmt.Sprintf("Watch err: %v", watchResp.Error),
						changes:       "",
						watcherType:   dw.Type(),
					})
					break
				}
				dw.client.log(&logInfo{
					NamespaceName: watchResp.Namespace,
					extra:         fmt.Sprintf("Watch config changes"),
					changes:       fmt.Sprintf("diff: %s", watchResp.OldValue.Different(watchResp.NewValue)),
					watcherType:   dw.Type(),
				})
				// 数据存map
				dw.client.Store(watchResp.Namespace, watchResp.NewValue)
			case <-dw.client.Done():
				return
			}
		}
	}()
}

// 关闭watcher
func (dw *daemonWatcher) Close() {

}

// 获取watcher类型
func (dw *daemonWatcher) Type() WatcherType {
	return DaemonWatcher
}

package test

import (
	"github.com/Shopify/toxiproxy"
	proxyClient "github.com/Shopify/toxiproxy/client"
	"github.com/gorilla/mux"

	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	// 单测代理Host
	UnitProxyHost = "localhost"
	// 单测代理端口
	UnitProxyPort = 8474
)

var (
	// InitClientDelay 初始化客户端延迟
	InitClientDelay = time.Second

	// proxyServerMutex 单测代理服务端的并发锁
	proxyServerMutex sync.Mutex
	// proxyServer 单测代理服务端
	proxyServer *toxiproxy.ApiServer
	// proxyHTTPServer 单测代理服务端的http监听服务
	proxyHTTPServer *http.Server
)

// 监听单测代理服务端
func listenUnitProxyServer(address string) {
	proxyServer = toxiproxy.NewServer()

	r := mux.NewRouter()
	r.HandleFunc("/reset", proxyServer.ResetState).Methods("POST")
	r.HandleFunc("/proxies", proxyServer.ProxyIndex).Methods("GET")
	r.HandleFunc("/proxies", proxyServer.ProxyCreate).Methods("POST")
	r.HandleFunc("/populate", proxyServer.Populate).Methods("POST")
	r.HandleFunc("/proxies/{proxy}", proxyServer.ProxyShow).Methods("GET")
	r.HandleFunc("/proxies/{proxy}", proxyServer.ProxyUpdate).Methods("POST")
	r.HandleFunc("/proxies/{proxy}", proxyServer.ProxyDelete).Methods("DELETE")
	r.HandleFunc("/proxies/{proxy}/toxics", proxyServer.ToxicIndex).Methods("GET")
	r.HandleFunc("/proxies/{proxy}/toxics", proxyServer.ToxicCreate).Methods("POST")
	r.HandleFunc("/proxies/{proxy}/toxics/{toxic}", proxyServer.ToxicShow).Methods("GET")
	r.HandleFunc("/proxies/{proxy}/toxics/{toxic}", proxyServer.ToxicUpdate).Methods("POST")
	r.HandleFunc("/proxies/{proxy}/toxics/{toxic}", proxyServer.ToxicDelete).Methods("DELETE")
	r.HandleFunc("/version", proxyServer.Version).Methods("GET")

	proxyHTTPServer = &http.Server{
		Addr:    address,
		Handler: r,
	}
	err := proxyHTTPServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

// 初始化单测代理服务端
func initUnitProxyServer() string {
	address := net.JoinHostPort(UnitProxyHost, strconv.Itoa(UnitProxyPort))

	if proxyServer != nil {
		return address
	}

	// 避免多次调用创建多次server
	proxyServerMutex.Lock()
	defer proxyServerMutex.Unlock()
	if proxyServer == nil {
		go listenUnitProxyServer(address)
	}

	return address
}

// 初始化单测代理客户端
func initUnitProxyClient(address string, cfg map[string]UnitProxyConfig) (*proxyClient.Client, error) {
	toxiClient := proxyClient.NewClient(address)

	proxyList := make([]proxyClient.Proxy, 0)
	for name, pc := range cfg {
		proxyList = append(proxyList, proxyClient.Proxy{
			Name:     name,
			Listen:   pc.Listen,
			Upstream: pc.Upstream,
			Enabled:  pc.Enabled,
		})
	}

	_, err := toxiClient.Populate(proxyList)
	if err != nil {
		return nil, err
	}

	return toxiClient, nil
}

// 初始化单测代理
func InitUnitProxy(configPath, pkgName string) (*proxyClient.Proxy, *UnitConfig) {
	unitConfig := DecodeUnitConfigFromLocal(configPath)

	address := initUnitProxyServer()

	time.Sleep(InitClientDelay)

	toxiClient, err := initUnitProxyClient(address, unitConfig.Proxy)
	if err != nil {
		panic(err)
	}
	proxy, err := toxiClient.Proxy(pkgName)
	if err != nil {
		panic(err)
	}

	return proxy, unitConfig
}

// 关闭单测代理
func CloseUnitProxy() {
	if proxyServer != nil {
		for _, proxy := range proxyServer.Collection.Proxies() {
			proxy.Stop()
		}
		proxyServer = nil
	}
	if proxyHTTPServer != nil {
		_ = proxyHTTPServer.Close()
		proxyHTTPServer = nil
	}
}

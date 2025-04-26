package httpclient

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/ctime"
	"gitlab.shanhai.int/sre/library/base/hook"
	"gitlab.shanhai.int/sre/library/net/circuitbreaker"
)

// 客户端
type Client struct {
	// 配置文件
	conf *Config
	// 钩子管理器
	manager *hook.Manager
	// 断路器组
	breakerGroup *circuitbreaker.BreakerGroup
	// 负载均衡器
	loadBalancer *LoadBalancer
	// 全局上下文
	globalContext context.Context
	// 全局上下文取消函数
	globalCancelFunc context.CancelFunc
	// 实际客户端
	client *http.Client
}

// 打开http客户端
func Open(c *Config) (*Client, error) {
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = DefaultMaxIdleConns
	}
	if c.MaxIdleConnsPerHost == 0 {
		c.MaxIdleConnsPerHost = DefaultMaxIdleConnsPerHost
	}
	if c.IdleConnTimeout == 0 {
		c.IdleConnTimeout = ctime.Duration(DefaultIdleConnTimeout)
	}
	if c.Config.StdoutPattern == "" {
		c.Config.StdoutPattern = defaultPattern
	}
	if c.Config.OutPattern == "" {
		c.Config.OutPattern = defaultPattern
	}
	if c.Config.OutFile == "" {
		c.Config.OutFile = _infoFile
	}

	client := new(Client)
	client.conf = c
	client.manager = NewHookManager().RegisterLogHook(c.Config, patternMap)
	client.breakerGroup = circuitbreaker.NewBreakerGroup()
	client.globalContext, client.globalCancelFunc = context.WithCancel(context.Background())
	client.client = &http.Client{
		Timeout: time.Duration(c.RequestTimeout),
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			DisableKeepAlives:     c.DisableKeepAlives,
			MaxIdleConns:          c.MaxIdleConns,
			MaxIdleConnsPerHost:   c.MaxIdleConnsPerHost,
			MaxConnsPerHost:       c.MaxConnsPerHost,
			IdleConnTimeout:       time.Duration(c.IdleConnTimeout),
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	if client.conf.EnableLoadBalancer {
		balancer, err := NewLoadBalancer(client.globalContext, client.manager.GetLogger())
		if err != nil {
			return nil, errors.Wrap(err, "http client init load balancer error")
		}
		client.loadBalancer = balancer
	}

	return client, nil
}

// 设置重定向函数
func (c *Client) CheckRedirect(f func(req *http.Request, via []*http.Request) error) *Client {
	c.client.CheckRedirect = f
	return c
}

// 设置原始Client
func (c *Client) Client(client *http.Client) *Client {
	c.client = client
	return c
}

// 获取原始Client
func (c *Client) GetClient() *http.Client {
	return c.client
}

// 设置不同域名的断路器
func (c *Client) Breaker(host string, breaker *circuitbreaker.Breaker) *Client {
	c.breakerGroup.Add(host, breaker)
	return c
}

// 创建构建器
func (c *Client) Builder() *Span {
	return NewSpan(c)
}

// Deprecated: 推荐使用 Builder() 方法
// 请求form
func (c *Client) fetchForm(ctx context.Context, method, url string, queryParams *UrlValue, headers *Header, requestForm *Form) (resp *Response) {
	if headers == nil {
		headers = NewFormURLEncodedHeader()
	}

	return c.Builder().
		Method(method).
		URL(url).
		QueryParams(queryParams).
		Headers(headers).
		FormBody(requestForm).
		Fetch(ctx)
}

// Deprecated: 推荐使用 Builder() 方法
// 请求json
func (c *Client) fetchJSON(ctx context.Context, method, url string, queryParams *UrlValue, headers *Header, requestBody interface{}) (resp *Response) {
	if headers == nil {
		headers = NewJsonHeader()
	}

	return c.Builder().
		Method(method).
		URL(url).
		QueryParams(queryParams).
		Headers(headers).
		JsonBody(requestBody).
		Fetch(ctx)
}

// Deprecated: 推荐使用 Builder() 方法
// Post请求Form
func (c *Client) PostForm(ctx context.Context, url string, queryParams *UrlValue, requestForm *Form, headers *Header) (resp *Response) {
	return c.fetchForm(ctx, http.MethodPost, url, queryParams, headers, requestForm)
}

// Deprecated: 推荐使用 Builder() 方法
// Patch请求Form
func (c *Client) PatchForm(ctx context.Context, url string, queryParams *UrlValue, requestForm *Form, headers *Header) (resp *Response) {
	return c.fetchForm(ctx, http.MethodPatch, url, queryParams, headers, requestForm)
}

// Deprecated: 推荐使用 Builder() 方法
// Put请求Form
func (c *Client) PutForm(ctx context.Context, url string, queryParams *UrlValue, requestForm *Form, headers *Header) (resp *Response) {
	return c.fetchForm(ctx, http.MethodPut, url, queryParams, headers, requestForm)
}

// Deprecated: 推荐使用 Builder() 方法
// Get请求json
func (c *Client) GetJSON(ctx context.Context, url string, queryParams *UrlValue, headers *Header) (resp *Response) {
	return c.fetchJSON(ctx, http.MethodGet, url, queryParams, headers, nil)
}

// Deprecated: 推荐使用 Builder() 方法
// Post请求json
func (c *Client) PostJSON(ctx context.Context, url string, queryParams *UrlValue, body interface{}, headers *Header) (resp *Response) {
	return c.fetchJSON(ctx, http.MethodPost, url, queryParams, headers, body)
}

// Deprecated: 推荐使用 Builder() 方法
// Patch请求json
func (c *Client) PatchJSON(ctx context.Context, url string, queryParams *UrlValue, body interface{}, headers *Header) (resp *Response) {
	return c.fetchJSON(ctx, http.MethodPatch, url, queryParams, headers, body)
}

// Deprecated: 推荐使用 Builder() 方法
// Put请求json
func (c *Client) PutJSON(ctx context.Context, url string, queryParams *UrlValue, body interface{}, headers *Header) (resp *Response) {
	return c.fetchJSON(ctx, http.MethodPut, url, queryParams, headers, body)
}

// Deprecated: 推荐使用 Builder() 方法
// Delete请求json
func (c *Client) DeleteJSON(ctx context.Context, url string, queryParams *UrlValue, headers *Header) (resp *Response) {
	return c.fetchJSON(ctx, http.MethodDelete, url, queryParams, headers, nil)
}

// Deprecated: 推荐使用 Builder() 方法
// Head请求json
func (c *Client) HeadJSON(ctx context.Context, url string, queryParams *UrlValue, headers *Header) (resp *Response) {
	return c.fetchJSON(ctx, http.MethodHead, url, queryParams, headers, nil)
}

// Deprecated: 推荐使用 Builder() 方法
// Get请求
func (c *Client) Get(ctx context.Context, url string, queryParams *UrlValue, headers *Header) (resp *Response) {
	return c.Builder().
		Method(http.MethodGet).
		URL(url).
		QueryParams(queryParams).
		Headers(headers).
		Fetch(ctx)
}

// Deprecated: 推荐使用 Builder() 方法
// Post请求
func (c *Client) Post(ctx context.Context, url string, queryParams *UrlValue, body []byte, headers *Header) (resp *Response) {
	return c.Builder().
		Method(http.MethodPost).
		URL(url).
		QueryParams(queryParams).
		Headers(headers).
		Body(body).
		Fetch(ctx)
}

// Deprecated: 推荐使用 Builder() 方法
// Patch请求
func (c *Client) Patch(ctx context.Context, url string, queryParams *UrlValue, body []byte, headers *Header) (resp *Response) {
	return c.Builder().
		Method(http.MethodPatch).
		URL(url).
		QueryParams(queryParams).
		Headers(headers).
		Body(body).
		Fetch(ctx)
}

// Deprecated: 推荐使用 Builder() 方法
// Put请求
func (c *Client) Put(ctx context.Context, url string, queryParams *UrlValue, body []byte, headers *Header) (resp *Response) {
	return c.Builder().
		Method(http.MethodPut).
		URL(url).
		QueryParams(queryParams).
		Headers(headers).
		Body(body).
		Fetch(ctx)
}

// Deprecated: 推荐使用 Builder() 方法
// Delete请求
func (c *Client) Delete(ctx context.Context, url string, queryParams *UrlValue, headers *Header) (resp *Response) {
	return c.Builder().
		Method(http.MethodDelete).
		URL(url).
		QueryParams(queryParams).
		Headers(headers).
		Fetch(ctx)
}

// Deprecated: 推荐使用 Builder() 方法
// Head请求
func (c *Client) Head(ctx context.Context, url string, queryParams *UrlValue, headers *Header) (resp *Response) {
	return c.Builder().
		Method(http.MethodHead).
		URL(url).
		QueryParams(queryParams).
		Headers(headers).
		Fetch(ctx)
}

// 关闭当前客户端
func (c *Client) Close() error {
	c.globalCancelFunc()
	c.manager.Close()
	return nil
}

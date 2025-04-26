package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/hook"
	"gitlab.shanhai.int/sre/library/base/runtime"
	"gitlab.shanhai.int/sre/library/net/circuitbreaker"
)

type Span struct {
	// 客户端
	client *Client
	// 实际客户端
	originClient http.Client

	// 配置文件
	conf Config
	// 钩子管理器
	manager hook.Manager

	// 错误
	err error
	// 调用源
	source string
	// 请求开始时间
	startTime time.Time
	// 请求结束时间
	endTime time.Time
	// 请求持续时间
	duration time.Duration

	// 降级后的响应
	degradedResponse *http.Response

	// 过滤方法
	filterFunc func(*http.Request, *http.Response) error
	// 通过的状态码数组
	accessStatusCode []int

	// 方法
	method string
	// 请求url
	url string
	// 请求host
	host string
	// 请求endpoint
	endpoint string
	// 请求查询参数
	queryParams *UrlValue
	// 请求头
	headers *Header
	// 请求体
	body []byte

	// 当前context
	ctx context.Context
	// 客户端请求
	Request *http.Request
	// 客户端响应
	Response *http.Response

	// 处理方法数组链
	handlerChain []HandlerFunc
	// 当前处理索引
	handlerIndex int
}

// 构建方法
func NewSpan(client *Client) *Span {
	span := &Span{
		conf:             *client.conf,
		manager:          *client.manager,
		accessStatusCode: []int{http.StatusOK},
		handlerIndex:     -1,
		// tracing、metrics、sentry、log通过hook来实现
		handlerChain: []HandlerFunc{
			GetBreakerHandler(), GetFilterHandler(), GetMetricsHandler(),
			GetTracingHandler(), GetSentryHandler(), GetK8sLoadBalancerHandler(),
		},
		client:       client,
		originClient: *client.client,
	}

	return span
}

// 进行下一个处理
func (b *Span) Next() {
	b.handlerIndex++
	if b.handlerIndex < len(b.handlerChain) {
		b.handlerChain[b.handlerIndex](b)
	}
}

// 实际请求
func (b *Span) fetch() (*http.Response, error) {
	var bodyReader io.Reader
	if b.body != nil {
		bodyReader = bytes.NewReader(b.body)
	}
	req, err := http.NewRequest(b.method, b.url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header = b.headers.Header

	req = req.WithContext(
		httptrace.WithClientTrace(
			b.ctx,
			&httptrace.ClientTrace{
				GotConn: func(info httptrace.GotConnInfo) {
					b.endpoint = info.Conn.RemoteAddr().String()
				},
			},
		),
	)
	b.host = req.URL.Host

	b.Request = req
	b.Next()

	return b.Response, b.err
}

// 获取当前url的断路器
func (b *Span) GetUrlBreaker(rawUrl string) (*circuitbreaker.Breaker, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	if breaker := b.client.breakerGroup.Get(parsedUrl.Host); breaker != nil {
		return breaker, nil
	} else {
		breaker = circuitbreaker.NewRateBreaker(b.conf.BreakerRate, int64(b.conf.BreakerMinSample))
		b.client.breakerGroup.Add(parsedUrl.Host, breaker)
		return breaker, nil
	}
}

// 获取配置
func (b *Span) GetConfig() Config {
	return b.conf
}

// 设置是否打印请求body
func (b *Span) RequestBodyOut(requestBodyOut bool) *Span {
	b.conf.RequestBodyOut = requestBodyOut
	return b
}

// 设置是否打印响应body
func (b *Span) ResponseBodyOut(responseBodyOut bool) *Span {
	b.conf.ResponseBodyOut = responseBodyOut
	return b
}

// 设置请求超时时间
func (b *Span) RequestTimeout(timeout time.Duration) *Span {
	b.originClient.Timeout = timeout
	return b
}

// 设置重定向函数
func (b *Span) CheckRedirect(f func(req *http.Request, via []*http.Request) error) *Span {
	b.originClient.CheckRedirect = f
	return b
}

// 设置是否关闭断路器
func (b *Span) DisableBreaker(disableBreaker bool) *Span {
	b.conf.DisableBreaker = disableBreaker
	return b
}

// 设置断路器断路最小错误比例 0~1
func (b *Span) BreakerRate(rate float64) *Span {
	if rate > 1.0 || rate < 0 {
		b.err = errors.New("breaker rate is invalid")
		return b
	}

	b.conf.BreakerRate = rate
	return b
}

// 设置断路器断路最小采样数
func (b *Span) BreakerMinSample(minSample int) *Span {
	if minSample < 0 {
		b.err = errors.New("breaker min sample is invalid")
		return b
	}

	b.conf.BreakerMinSample = minSample
	return b
}

// 设置是否关闭链路跟踪
func (b *Span) DisableTracing(disableTracing bool) *Span {
	b.conf.DisableTracing = disableTracing
	return b
}

// 设置是否关闭sentry
func (b *Span) DisableSentry(disableSentry bool) *Span {
	b.conf.DisableSentry = disableSentry
	return b
}

// 获取方法
func (b *Span) GetMethod() string {
	return b.method
}

// 设置方法
func (b *Span) Method(method string) *Span {
	b.method = method
	return b
}

// 获取请求url
func (b *Span) GetURL() string {
	return b.url
}

// 设置请求url
func (b *Span) URL(url string) *Span {
	b.url = url
	return b
}

// 获取请求查询参数
func (b *Span) GetQueryParams() *UrlValue {
	return b.queryParams
}

// 设置请求查询参数
func (b *Span) QueryParams(queryParams *UrlValue) *Span {
	b.queryParams = queryParams
	return b
}

// 获取请求头
func (b *Span) GetHeaders() *Header {
	return b.headers
}

// 设置请求头
func (b *Span) Headers(headers *Header) *Span {
	b.headers = headers
	return b
}

// 获取请求体
func (b *Span) GetBody() []byte {
	return b.body
}

// 设置请求体
func (b *Span) Body(body []byte) *Span {
	b.body = body
	return b
}

// 获取请求持续时间
func (b *Span) GetDuration() time.Duration {
	return b.duration
}

// 获取请求开始时间
func (b *Span) GetStartTime() time.Time {
	return b.startTime
}

// 获取请求结束时间
func (b *Span) GetEndTime() time.Time {
	return b.endTime
}

// 设置表单请求体
func (b *Span) FormBody(requestForm *Form) *Span {
	if requestForm == nil {
		return b
	}

	b.body = []byte(requestForm.Encode())
	return b
}

// 设置Json请求体
func (b *Span) JsonBody(requestBody interface{}) *Span {
	if requestBody == nil {
		return b
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		b.err = errors.WithStack(err)
		return b
	}

	b.body = requestJSON
	return b
}

// 获取降级后的响应
func (b *Span) GetDegradedResponse() *http.Response {
	return b.degradedResponse
}

// 设置降级后的响应
func (b *Span) DegradedResponse(statusCode int, body []byte) *Span {
	b.degradedResponse = &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(body)),
	}
	return b
}

// 设置降级后的Json响应
func (b *Span) DegradedJsonResponse(body interface{}) *Span {
	requestJSON, err := json.Marshal(body)
	if err != nil {
		b.err = errors.WithStack(err)
		return b
	}

	return b.DegradedResponse(http.StatusOK, requestJSON)
}

// 获取过滤方法
func (b *Span) GetFilterFunc() func(*http.Request, *http.Response) error {
	return b.filterFunc
}

// 设置过滤方法
func (b *Span) FilterFunc(filterFunc func(*http.Request, *http.Response) error) *Span {
	b.filterFunc = filterFunc
	return b
}

// 设置空的过滤方法
func (b *Span) EmptyFilterFunc() *Span {
	b.filterFunc = func(request *http.Request, response *http.Response) error {
		return nil
	}
	return b
}

// 获取通过的状态码
func (b *Span) GetAccessStatusCode() []int {
	return b.accessStatusCode
}

// 设置通过的状态码
func (b *Span) AccessStatusCode(statusCode ...int) *Span {
	b.accessStatusCode = statusCode
	return b
}

// 增加处理方法
func (b *Span) AddHandler(handlerFunc HandlerFunc) *Span {
	b.handlerChain = append(b.handlerChain, handlerFunc)
	return b
}

// 获取错误
func (b *Span) GetError() error {
	return b.err
}

// 设置错误
func (b *Span) SetError(err error) *Span {
	if err != nil {
		b.err = err
	}
	return b
}

// 获取上下文
func (b *Span) GetContext() context.Context {
	return b.ctx
}

// 请求
func (b *Span) Fetch(ctx context.Context) *Response {
	if b.err != nil {
		return NewResponse(nil, b.err)
	}
	b.source = runtime.GetDefaultFilterCallers()

	if b.queryParams != nil && b.queryParams.Values != nil {
		b.url = fmt.Sprintf("%s?%s", b.url, b.queryParams.Encode())
	}
	if b.headers == nil {
		b.headers = GetDefaultHeader()
	}

	b.ctx = ctx
	b.AddHandler(GetHookHandler())
	b.AddHandler(GetOriginHttpHandler())

	resp, err := b.fetch()
	return NewResponse(resp, err)
}

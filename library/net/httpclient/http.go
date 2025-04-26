package httpclient

import (
	"github.com/pkg/errors"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 新建http客户端
func NewHttpClient(c *Config) (client *Client) {
	if c == nil {
		panic("http client config is nil")
	}
	if c.RequestTimeout == 0 {
		panic("http client must be set request timeout")
	}
	if c.BreakerMinSample < 0 {
		panic("breaker min sample is invalid")
	}
	if c.BreakerRate > 1.0 || c.BreakerRate < 0 {
		panic("breaker rate is invalid")
	}

	if c.BreakerMinSample == 0 {
		c.BreakerMinSample = 10
	}
	if c.BreakerRate == 0.0 {
		c.BreakerRate = 0.5
	}

	if c.Config == nil {
		c.Config = &render.Config{}
	}

	client, err := Open(c)
	if err != nil {
		panic(errors.Wrap(err, "open http client error"))
	}

	return
}

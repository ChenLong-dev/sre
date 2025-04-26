package gin

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func Example() {
	c := Config{
		Endpoint: &EndpointConfig{
			Address: "0.0.0.0",
			Port:    80,
		},
		Timeout: ctime.Duration(time.Second * 2),
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [status: %s] [duration: %L] [ip: %C] URL: %M::%P , %E",
		},
		StdoutRouterPattern: "[%T] [%t] URL: %M::%P  %N  (%l handlers)",
	}
	binding.Validator = NewV10Validator()

	gin.DefaultWriter = GetInfoWriter(&c)
	gin.DefaultErrorWriter = GetErrorWriter(&c)
	gin.DebugPrintRouteFunc = GetDefaultRouterPrintFunc(&c)

	engine := gin.New()
	engine.Use(GetDefaultFormatter(&c))
}

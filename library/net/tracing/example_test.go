package tracing

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func ExampleNew() {
	New(&Config{
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [level: %L] %M",
		},
		AppName: "library",
		Sampler: &SamplerConfig{
			Param: "0.1",
		},
		Reporter: &ReporterConfig{
			CollectorEndpoint: "http://localhost:9411/api/v2/spans",
		},
	})

	router := gin.New()
	router.Use(ExtractFromUpstream())
	router.Use(InjectToDownstream())
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
	if err != nil {
		return
	}
	fmt.Printf("%d", r.Code)

	Close()

	// OutPut:
	// 200
}

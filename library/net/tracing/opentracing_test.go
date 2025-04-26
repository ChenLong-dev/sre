package tracing

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func TestJaeger(t *testing.T) {
	t.Skipf("should have jaeger environment\n")

	New(&Config{
		Config: &render.Config{
			Stdout: true,
		},
		AppName: "library",
		Sampler: &SamplerConfig{
			Type:              "const",
			Param:             "1",
			SamplingServerURL: "http://localhost:5778/sampling",
		},
		Reporter: &ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "127.0.0.1:6381",
			CollectorEndpoint:  "http://localhost:14268/api/traces",
		},
	})

	router := gin.New()
	router.Use(ExtractFromUpstream())
	router.Use(InjectToDownstream())
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, r.Code)

	Close()
}

func TestZipkin(t *testing.T) {
	t.Skipf("should have zipkin environment\n")

	New(&Config{
		Config: &render.Config{
			Stdout: true,
		},
		AppName: "library",
		Sampler: &SamplerConfig{
			Type:  SamplerTypeModulo,
			Param: "1",
		},
		Reporter: &ReporterConfig{
			CollectorEndpoint: "http://localhost:9411/api/v2/spans",
		},
	})

	router := gin.New()
	router.Use(ExtractFromUpstream())
	router.Use(InjectToDownstream())
	router.GET("/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	r, err := httpUtil.TestGinJsonRequest(router, "GET", "/123", nil, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, r.Code)

	Close()
}

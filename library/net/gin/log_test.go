package gin

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func TestGetDefaultFormatter(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		c := Config{
			Endpoint: &EndpointConfig{
				Address: "0.0.0.0",
				Port:    80,
			},
			Timeout:        ctime.Duration(time.Second * 2),
			RequestBodyOut: true,
			Config: &render.Config{
				Stdout: true,
				StdoutPattern: "[%b - %T] [%t] [%U] [status: %s] [duration: %L] [ip: %C] " +
					"[relative: %R] URL: %M::%P , Header: %H , " +
					"RequestBody: %B %E",
			},
			StdoutRouterPattern: "[%T] [%t] URL: %M::%P  %N  (%l handlers)",
		}

		gin.DefaultWriter = GetInfoWriter(&c)
		gin.DefaultErrorWriter = GetErrorWriter(&c)
		gin.DebugPrintRouteFunc = GetDefaultRouterPrintFunc(&c)

		engine := gin.New()
		engine.Use(GetDefaultFormatter(&c))
		engine.POST("/:id", func(c *gin.Context) {
			var i interface{}
			err := c.ShouldBindJSON(&i)
			assert.Nil(t, err)
			time.Sleep(time.Second)
			c.JSON(http.StatusOK, nil)
		})

		header := http.Header{}
		header.Add("agent", "123")
		header.Add(QTHeaderUserIDKey, "160aeddd3d13444ea735ca8099104562")
		header.Add(QTHeaderDeviceIDKey, "41160d3b-8734-33fa-af23-24718bb72a85")
		r, err := httpUtil.TestGinJsonRequest(engine, "POST",
			"/123", header, header, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("empty config", func(t *testing.T) {
		c := Config{
			Endpoint: &EndpointConfig{
				Address: "0.0.0.0",
				Port:    80,
			},
			Timeout:             ctime.Duration(time.Second * 2),
			StdoutRouterPattern: "[%T] [%t] URL: %M::%P  %N  (%l handlers)",
		}

		gin.DefaultWriter = GetInfoWriter(&c)
		gin.DefaultErrorWriter = GetErrorWriter(&c)
		gin.DebugPrintRouteFunc = GetDefaultRouterPrintFunc(&c)

		engine := gin.New()
		engine.Use(GetDefaultFormatter(&c))
		engine.GET("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(engine, "GET", "/123", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestSetCustomLog(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		c := Config{
			Endpoint: &EndpointConfig{
				Address: "0.0.0.0",
				Port:    80,
			},
			Timeout:        ctime.Duration(time.Second * 2),
			RequestBodyOut: true,
			Config: &render.Config{
				Stdout: true,
			},
		}

		gin.DefaultWriter = GetInfoWriter(&c)
		gin.DefaultErrorWriter = GetErrorWriter(&c)
		gin.DebugPrintRouteFunc = GetDefaultRouterPrintFunc(&c)

		engine := gin.New()
		engine.Use(GetDefaultFormatter(&c))
		engine.POST("/:id", func(c *gin.Context) {
			var i interface{}
			err := c.ShouldBindJSON(&i)
			assert.Nil(t, err)
			time.Sleep(time.Second)
			SetCustomLog(c, i)
			c.JSON(http.StatusOK, nil)
		})

		header := http.Header{}
		header.Add("agent", "123")
		header.Add(QTHeaderUserIDKey, "160aeddd3d13444ea735ca8099104562")
		header.Add(QTHeaderDeviceIDKey, "41160d3b-8734-33fa-af23-24718bb72a85")
		r, err := httpUtil.TestGinJsonRequest(engine, "POST",
			"/123", header, header, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

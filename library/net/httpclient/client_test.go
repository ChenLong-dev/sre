package httpclient

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
	"gitlab.shanhai.int/sre/library/net/tracing"
)

var (
	c *Client
)

func TestMain(m *testing.M) {
	c = NewHttpClient(&Config{
		RequestTimeout:     ctime.Duration(5 * time.Second),
		RequestBodyOut:     true,
		ResponseBodyOut:    true,
		DisableTracing:     true,
		DisableBreaker:     true,
		EnableLoadBalancer: true,
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%a - %T] [%t] [%U] [status: %s] [duration: %D] %S  Endpoint: %e  URL: %M::%u , Header: %h , RequestBody: %b , ResponseBody: %B , Extra: %E",
		},
	})

	os.Exit(m.Run())
}

type GithubRepoOwner struct {
	ID      int    `json:"id"`
	Login   string `json:"login"`
	Private bool   `json:"private"`
}

type GithubRepo struct {
	ID        int              `json:"id"`
	Name      string           `json:"name"`
	Private   bool             `json:"private"`
	CreatedAt *time.Time       `json:"created_at"`
	Owner     *GithubRepoOwner `json:"owner"`
}

func TestClient_GetJSON(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		result := make([]GithubRepo, 0)
		err := c.GetJSON(
			context.Background(),
			"https://api.github.com/users/cmonoceros/repos",
			NewUrlValue().Add("page", "1").Add("per_page", "2"),
			NewJsonHeader(),
		).DecodeJSON(&result)
		assert.Nil(t, err)
		assert.Equal(t, true, len(result) > 0)
		assert.NotEqual(t, 0, result[0].ID)
		assert.NotEqual(t, "", result[0].Name)
		assert.NotNil(t, result[0].CreatedAt)
		assert.NotNil(t, result[0].Owner)
		assert.NotEqual(t, 0, result[0].Owner.ID)
		assert.NotEqual(t, "", result[0].Owner.Login)
	})

	t.Run("body", func(t *testing.T) {
		body, err := c.GetJSON(
			context.Background(),
			"https://api.github.com/users/cmonoceros/repos",
			NewUrlValue().Add("page", "1").Add("per_page", "2"),
			NewJsonHeader(),
		).Body()
		assert.Nil(t, err)
		assert.NotEqual(t, "", body)
	})
}

func TestClient_PostForm(t *testing.T) {
	t.Run("form", func(t *testing.T) {
		router := gin.New()
		router.POST("/form", func(c *gin.Context) {
			v := c.PostForm("test")
			if v != "123" {
				c.Writer.WriteHeader(http.StatusInternalServerError)
				c.Abort()
				return
			}

			c.JSON(http.StatusOK, nil)
		})

		go func() {
			router.Run("0.0.0.0:80")
		}()

		time.Sleep(time.Second)
		resp := c.PostForm(
			context.Background(),
			"http://0.0.0.0:80/form",
			nil,
			NewForm().Add("test", "123"),
			NewFormURLEncodedHeader(),
		)
		assert.Nil(t, resp.Error())
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("post", func(t *testing.T) {
		router := gin.New()
		router.POST("/form", func(c *gin.Context) {
			v := c.PostForm("test")
			if v != "123" {
				c.Writer.WriteHeader(http.StatusInternalServerError)
				c.Abort()
				return
			}

			c.JSON(http.StatusOK, nil)
		})

		go func() {
			router.Run("0.0.0.0:80")
		}()

		time.Sleep(time.Second)
		resp := c.Post(
			context.Background(),
			"http://0.0.0.0:80/form",
			nil,
			[]byte(NewForm().Add("test", "123").Encode()),
			NewFormURLEncodedHeader(),
		)
		assert.Nil(t, resp.Error())
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("timeout", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			time.Sleep(time.Second * 10)

			c.JSON(http.StatusOK, nil)
		})

		go func() {
			router.Run("0.0.0.0:80")
		}()

		time.Sleep(time.Second)
		preTime := time.Now()
		resp := c.PostJSON(
			context.Background(),
			"http://0.0.0.0:80/",
			nil,
			nil,
			NewJsonHeader(),
		)
		assert.NotNil(t, resp.Error())
		assert.Equal(t, true, time.Now().Sub(preTime).Seconds() < 10)
	})
}

func TestClient_Get(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		result := make([]GithubRepo, 0)
		err := c.Get(
			context.Background(),
			"https://api.github.com/users/cmonoceros/repos",
			NewUrlValue().Add("page", "1").Add("per_page", "2"),
			NewJsonHeader(),
		).DecodeJSON(&result)
		assert.Nil(t, err)
		assert.Equal(t, true, len(result) > 0)
		assert.NotEqual(t, 0, result[0].ID)
		assert.NotEqual(t, "", result[0].Name)
		assert.NotNil(t, result[0].CreatedAt)
		assert.NotNil(t, result[0].Owner)
		assert.NotEqual(t, 0, result[0].Owner.ID)
		assert.NotEqual(t, "", result[0].Owner.Login)
	})
}

func TestTracingClient(t *testing.T) {
	t.Skipf("should have tracing environment\n")

	c.conf.DisableTracing = false

	t.Run("zipkin", func(t *testing.T) {
		tracing.New(&tracing.Config{
			Config: &render.Config{
				Stdout: true,
			},
			AppName: "library",
			Sampler: &tracing.SamplerConfig{
				Type:  tracing.SamplerTypeModulo,
				Param: "1",
			},
			Reporter: &tracing.ReporterConfig{
				CollectorEndpoint: "http://localhost:9411/api/v2/spans",
			},
		})

		router := gin.New()
		router.Use(tracing.ExtractFromUpstream())
		router.Use(tracing.InjectToDownstream())
		router.GET("/", func(ctx *gin.Context) {
			result := make([]GithubRepo, 0)
			err := c.GetJSON(
				ctx,
				"https://api.github.com/users/cmonoceros/repos",
				NewUrlValue().Add("page", "1").Add("per_page", "2"),
				NewJsonHeader(),
			).DecodeJSON(&result)
			assert.Nil(t, err)
			ctx.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)

		tracing.Close()
	})

	t.Run("jaeger", func(t *testing.T) {
		tracing.New(&tracing.Config{
			Config: &render.Config{
				Stdout: true,
			},
			AppName: "library",
			Sampler: &tracing.SamplerConfig{
				Type:              "const",
				Param:             "1",
				SamplingServerURL: "http://localhost:5778/sampling",
			},
			Reporter: &tracing.ReporterConfig{
				LogSpans:           true,
				LocalAgentHostPort: "127.0.0.1:6381",
				CollectorEndpoint:  "http://localhost:14268/api/traces",
			},
		})

		router := gin.New()
		router.Use(tracing.ExtractFromUpstream())
		router.Use(tracing.InjectToDownstream())
		router.GET("/", func(ctx *gin.Context) {
			result := make([]GithubRepo, 0)
			err := c.GetJSON(
				ctx,
				"https://api.github.com/users/cmonoceros/repos",
				NewUrlValue().Add("page", "1").Add("per_page", "2"),
				NewJsonHeader(),
			).DecodeJSON(&result)
			assert.Nil(t, err)
			ctx.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)

		tracing.Close()
	})
}

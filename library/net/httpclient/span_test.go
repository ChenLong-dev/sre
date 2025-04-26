package httpclient

import (
	"context"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func TestClient_Span(t *testing.T) {
	t.Run("ResponseBodyOut-normal", func(t *testing.T) {
		result := make([]GithubRepo, 0)
		err := c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/users/cmonoceros/repos").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			Fetch(context.Background()).
			DecodeJSON(&result)
		assert.Nil(t, err)
	})

	t.Run("ResponseBodyOut-both", func(t *testing.T) {
		result := make([]GithubRepo, 0)
		err := c.GetJSON(
			context.Background(),
			"https://api.github.com/users/cmonoceros/repos",
			NewUrlValue().Add("page", "1").Add("per_page", "2"),
			NewJsonHeader(),
		).DecodeJSON(&result)
		assert.Nil(t, err)

		err = c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/users/cmonoceros/repos").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			Fetch(context.Background()).
			DecodeJSON(&result)
		assert.Nil(t, err)

		err = c.GetJSON(
			context.Background(),
			"https://api.github.com/users/cmonoceros/repos",
			NewUrlValue().Add("page", "1").Add("per_page", "2"),
			NewJsonHeader(),
		).DecodeJSON(&result)
		assert.Nil(t, err)
	})
}

func TestGetBreakerHandler(t *testing.T) {
	errSpan := func() *Span {
		return c.Builder().
			ResponseBodyOut(false).
			URL("http://localhost:90/thisisatest").
			Headers(NewJsonHeader()).
			DisableBreaker(false)
	}

	normalSpan := func() *Span {
		return c.Builder().
			ResponseBodyOut(false).
			URL("http://localhost:90/").
			Headers(NewJsonHeader()).
			DisableBreaker(false)
	}

	type degradedResponse struct {
		IsDegraded bool
	}
	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, degradedResponse{
			IsDegraded: false,
		})
	})

	go func() {
		router.Run("0.0.0.0:90")
	}()

	time.Sleep(time.Second)

	t.Run("normal", func(t *testing.T) {
		isOpen := false
		wg := new(sync.WaitGroup)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

				resp := normalSpan().Fetch(context.Background())
				if errcode.EqualError(errcode.BreakerOpenError, resp.Error()) {
					isOpen = true
					return
				}
				assert.Nil(t, resp.Error())
			}()
		}
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

				resp := errSpan().Fetch(context.Background())
				assert.NotNil(t, resp.Error())
			}()
		}

		wg.Wait()
		assert.Equal(t, true, isOpen)
	})

	t.Run("degradation", func(t *testing.T) {
		isDegraded := false
		wg := new(sync.WaitGroup)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

				result := new(degradedResponse)
				err := normalSpan().
					DegradedJsonResponse(degradedResponse{
						IsDegraded: true,
					}).
					Fetch(context.Background()).
					DecodeJSON(result)
				assert.Nil(t, err)
				if result.IsDegraded == true {
					isDegraded = true
				}
			}()
		}
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

				resp := errSpan().Fetch(context.Background())
				assert.NotNil(t, resp.Error())
			}()
		}

		wg.Wait()
		assert.Equal(t, true, isDegraded)
	})

	t.Run("is_degraded", func(t *testing.T) {
		isDegraded := false
		wg := new(sync.WaitGroup)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

				result := new(degradedResponse)
				response := normalSpan().
					DegradedJsonResponse(degradedResponse{
						IsDegraded: true,
					}).
					Fetch(context.Background())
				if response.IsDegraded() {
					isDegraded = true
				}
				err := response.DecodeJSON(result)
				assert.Nil(t, err)
			}()
		}
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

				resp := errSpan().Fetch(context.Background())
				assert.NotNil(t, resp.Error())
			}()
		}

		wg.Wait()
		assert.Equal(t, true, isDegraded)
	})
}

func TestSpan_AccessStatusCode(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		resp := c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/users/cmonoceros/repos").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			Fetch(context.Background())
		assert.Nil(t, resp.Error())
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("filter", func(t *testing.T) {
		resp := c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/thisisatest").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			Fetch(context.Background())
		assert.NotNil(t, resp.Error())
	})

	t.Run("all", func(t *testing.T) {
		resp := c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/users/cmonoceros/repos").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			AccessStatusCode().
			Fetch(context.Background())
		assert.Nil(t, resp.Error())
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		resp = c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/thisisatest").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			AccessStatusCode().
			Fetch(context.Background())
		assert.Nil(t, resp.Error())
		assert.Equal(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("access", func(t *testing.T) {
		resp := c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/thisisatest").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			AccessStatusCode(http.StatusNotFound, http.StatusInternalServerError).
			Fetch(context.Background())
		assert.Nil(t, resp.Error())
		assert.Equal(t, resp.StatusCode, http.StatusNotFound)
	})
}

func TestSpan_RequestTimeout(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		// 默认5s
		result := make([]GithubRepo, 0)
		err := c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/users/cmonoceros/repos").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			Fetch(context.Background()).
			DecodeJSON(&result)
		assert.Nil(t, err)

		// 修改为100ms
		result = make([]GithubRepo, 0)
		err = c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/users/cmonoceros/repos").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			RequestTimeout(time.Millisecond * 100).
			Fetch(context.Background()).
			DecodeJSON(&result)
		assert.NotNil(t, err)
	})

	t.Run("side-effect", func(t *testing.T) {
		// 修改为100ms
		result := make([]GithubRepo, 0)
		err := c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/users/cmonoceros/repos").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			RequestTimeout(time.Millisecond * 100).
			Fetch(context.Background()).
			DecodeJSON(&result)
		assert.NotNil(t, err)

		// 查看是否影响
		result = make([]GithubRepo, 0)
		err = c.Builder().
			ResponseBodyOut(false).
			URL("https://api.github.com/users/cmonoceros/repos").
			QueryParams(NewUrlValue().Add("page", "1").Add("per_page", "2")).
			Headers(NewJsonHeader()).
			Fetch(context.Background()).
			DecodeJSON(&result)
		assert.Nil(t, err)
	})
}

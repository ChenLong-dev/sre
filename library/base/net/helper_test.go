package net

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestTestGinRequest(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		router := gin.New()
		router.GET("/", func(c *gin.Context) {
			c.Writer.WriteHeader(http.StatusGatewayTimeout)
			c.Abort()
		})

		w, err := TestGinJsonRequest(router, http.MethodGet, "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	})

	t.Run("body", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			m := make(map[string]interface{})
			err := c.ShouldBindJSON(&m)
			assert.Nil(t, err)

			if m["a"] == "123" {
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				c.Abort()
			} else {
				c.JSON(http.StatusOK, nil)
			}
		})

		w, err := TestGinJsonRequest(router, http.MethodPost, "/", nil, map[string]interface{}{
			"a": "123",
		}, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	})

	t.Run("form", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			m := struct {
				A string `form:"a"`
			}{}
			err := c.ShouldBindWith(&m, binding.Form)
			assert.Nil(t, err)

			if m.A == "123" {
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				c.Abort()
			} else {
				c.JSON(http.StatusOK, nil)
			}
		})

		form := make(url.Values)
		form.Add("a", "123")
		w, err := TestGinJsonRequest(router, http.MethodPost, "/", nil, nil, form)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	})

	t.Run("query", func(t *testing.T) {
		router := gin.New()
		router.GET("/", func(c *gin.Context) {
			var m struct {
				A string `form:"a"`
			}
			err := c.ShouldBindQuery(&m)
			assert.Nil(t, err)

			if m.A == "123" {
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				c.Abort()
			} else {
				c.JSON(http.StatusOK, nil)
			}
		})

		w, err := TestGinJsonRequest(router, http.MethodGet, "/?a=123", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	})

	t.Run("header", func(t *testing.T) {
		router := gin.New()
		router.GET("/", func(c *gin.Context) {
			value := c.GetHeader("a")
			if value == "123" {
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				c.Abort()
			} else {
				c.JSON(http.StatusOK, nil)
			}
		})

		header := make(http.Header)
		header.Add("a", "123")
		w, err := TestGinJsonRequest(router, http.MethodGet, "/", header, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	})
}

package gin

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
	"gitlab.shanhai.int/sre/library/base/null"
)

type bindingStruct struct {
	A null.Int `form:"a" json:"a,omitempty" binding:"gte=1"`
	B int      `form:"b" json:"b,omitempty" binding:"gte=1"`
	C *int     `form:"c" json:"c,omitempty" binding:"gte=1"`
}

func TestNormalBinding(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		router := gin.New()
		binding.Validator = NewV10Validator()
		router.GET("/", func(c *gin.Context) {
			s := new(bindingStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?b=0&a=1&c=1", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})

	t.Run("ptr", func(t *testing.T) {
		router := gin.New()
		binding.Validator = NewV10Validator()
		router.GET("/", func(c *gin.Context) {
			s := new(bindingStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?b=1&a=1&c=0", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})
}

func TestNullBinding(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		router := gin.New()
		binding.Validator = NewV10Validator()
		router.GET("/", func(c *gin.Context) {
			s := new(bindingStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?a=0&b=1&c=1", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})
}

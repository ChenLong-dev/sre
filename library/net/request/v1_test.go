package request

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func TestV1PaginationRequest(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		type GetUsersRequest struct {
			*V1PaginationRequest
			Type string `form:"type" json:"type"`
		}

		router := gin.New()
		router.GET("/", func(c *gin.Context) {
			queryParams := new(GetUsersRequest)
			err := c.ShouldBindQuery(queryParams)
			assert.Nil(t, err)
			assert.Equal(t, 0, queryParams.Page)
			assert.Equal(t, 10, queryParams.PageSize)
			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?page=0&pagesize=10&type=audio",
			nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

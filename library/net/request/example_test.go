package request

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func ExampleV1PaginationRequest() {
	type GetUsersRequest struct {
		*V1PaginationRequest
		Type string `form:"type" json:"type"`
	}

	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		queryParams := new(GetUsersRequest)
		err := c.ShouldBindQuery(queryParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
		fmt.Printf("%#v\n", queryParams.V1PaginationRequest)
		c.JSON(http.StatusOK, nil)
	})

	r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?page=0&pagesize=10&type=audio",
		nil, nil, nil)
	if err != nil {
		return
	}
	fmt.Println(r.Code)

	// OutPut:
	// &request.V1PaginationRequest{Page:0, PageSize:10}
	// 200
}

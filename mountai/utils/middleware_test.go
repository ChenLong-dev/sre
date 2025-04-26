package utils

import (
	"rulai/config"
	"rulai/models/resp"

	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
	"gitlab.shanhai.int/sre/library/net/httpclient"
)

func TestParseJWTTokenMiddleware(t *testing.T) {
	config.Read("../config/config.yaml")

	t.Run("normal", func(t *testing.T) {
		router := gin.New()
		router.Use(ParseJWTTokenMiddleware(&resp.UserProfileResp{
			ID:        "1",
			Name:      "k8s-system",
			Email:     "tars.qingtingfm.com",
			AvatarURL: "https://www.gravatar.com/avatar/2dab064cd9af9709f45e9b3f9e5c0f2e?Operation=80&d=identicon",
		}))
		router.GET("/", func(c *gin.Context) {
			v, ok := c.Get(ContextUserIDKey)
			if !ok || v != "1" {
				c.JSON(http.StatusUnauthorized, nil)
				return
			}

			c.JSON(http.StatusOK, nil)
		})

		token, err := GenerateJWTToken(context.Background(), &resp.GitUserProfileResp{
			ID:        1,
			Name:      "k8s-system",
			UserName:  "k8s-system",
			AvatarURL: "https://www.gravatar.com/avatar/2dab064cd9af9709f45e9b3f9e5c0f2e?Operation=80&d=identicon",
			Email:     "tars.qingtingfm.com",
		})
		assert.Nil(t, err)

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/",
			httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %v Operation", token)).Header,
			nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("invalid", func(t *testing.T) {
		router := gin.New()
		router.Use(ParseJWTTokenMiddleware(&resp.UserProfileResp{
			ID:        "-1",
			Name:      "k8s-system",
			Email:     "tars.qingtingfm.com",
			AvatarURL: "https://www.gravatar.com/avatar/2dab064cd9af9709f45e9b3f9e5c0f2e?Operation=80&d=identicon",
		}))
		router.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, nil)
		})

		token := "invalid token"
		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/",
			httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %v Operation", token)).Header,
			nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})

	t.Run("empty", func(t *testing.T) {
		router := gin.New()
		router.Use(ParseJWTTokenMiddleware(&resp.UserProfileResp{
			ID:        "-1",
			Name:      "k8s-system",
			Email:     "tars.qingtingfm.com",
			AvatarURL: "https://www.gravatar.com/avatar/2dab064cd9af9709f45e9b3f9e5c0f2e?Operation=80&d=identicon",
		}))
		router.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/",
			nil,
			nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
}

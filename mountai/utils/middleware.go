package utils

import (
	"context"
	"net/http"
	"strings"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gitlab.shanhai.int/sre/library/base/slice"
)

const (
	// context中用于存储用户id的key
	ContextUserIDKey = "user_id"

	// context中用于存储用户的key
	ContextUserClaimKey = "user_claim"

	// Internal user key.
	ContextInternalUserKey = "internal_user"
)

// 解析jwt token中间件
func ParseJWTTokenMiddleware(k8sSystemUser *resp.UserProfileResp) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := c.GetHeader("Authorization")
		token := strings.TrimPrefix(value, "Bearer ")
		var userClaim *entity.UserClaims
		if slice.StrSliceContains(config.Conf.JWT.K8sSystemUserTokens, token) {
			userClaim = &entity.UserClaims{
				Name:  k8sSystemUser.Name,
				Email: k8sSystemUser.Email,
				StandardClaims: jwt.StandardClaims{
					Id: k8sSystemUser.ID,
				},
			}
		} else {
			res, err := jwt.ParseWithClaims(token, new(entity.UserClaims), func(token *jwt.Token) (i interface{}, err error) {
				return []byte(config.Conf.JWT.SignKey), nil
			})
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			userClaim = res.Claims.(*entity.UserClaims)
		}

		c.Set(ContextUserIDKey, userClaim.Id)
		c.Set(ContextUserClaimKey, *userClaim)

		c.Next()
	}
}

// Service interface, it is just to solve import cycle.
type ServiceInterface interface {
	GetInternalSingleUser(context.Context, *req.GetInternalUsersReq) (*entity.InternalUser, error)
}

// 内部用户校验中间件
func ValidateInternalUserMiddleware(svc ServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(ContextUserClaimKey)
		if !exists {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userClaims, ok := user.(entity.UserClaims)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if userClaims.Id == entity.K8sSystemUserID {
			// Set default value.
			c.Set(ContextInternalUserKey, entity.InternalUser{})
			return
		}

		// internalUser, err := svc.GetInternalSingleUser(context.Background(), &req.GetInternalUsersReq{
		// 	Email: userClaims.Email,
		// })
		// if err != nil {
		// 	_ = c.AbortWithError(http.StatusInternalServerError, err).
		// 		SetMeta("内部用户不存在或邮箱未设置")
		// 	return
		// }

		c.Set(ContextInternalUserKey, &entity.InternalUser{})
	}
}

package utils

import (
	"context"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
)

const (
	// JWTExpireTimeSecond 用户jwt过期时间
	JWTExpireTimeSecond = 60 * 60 * 24
)

func GenerateJWTToken(ctx context.Context, user *resp.GitUserProfileResp) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, entity.UserClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        strconv.Itoa(user.ID),
			ExpiresAt: time.Now().Add(JWTExpireTimeSecond * time.Second).Unix(),
		},
		Name:  user.Name,
		Email: user.Email,
	})

	signString, err := token.SignedString([]byte(config.Conf.JWT.SignKey))
	if err != nil {
		return "", errors.Wrap(_errcode.JWTGenerateError, err.Error())
	}

	return signString, nil
}

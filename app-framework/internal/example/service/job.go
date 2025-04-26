package service

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"math/rand"
)

// 打印随机数
func (s *Service) PrintRandomNumber(ctx context.Context, i int) (err error) {
	// 防止panic中断整个程序
	defer func() {
		if e := recover(); e != nil {
			err = errors.Wrapf(errcode.InternalError, "%s", e)
		}
	}()

	log.Infoc(ctx, "Random number %d", rand.Intn(i+1))

	return nil
}

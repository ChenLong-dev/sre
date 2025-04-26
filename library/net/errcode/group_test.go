package errcode

import (
	pkgErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGroup(t *testing.T) {
	t.Run("codes", func(t *testing.T) {
		group := NewGroup(InvalidParams).
			AddChildren(
				pkgErrors.Wrap(MysqlError, "user id:1 not found"),
				pkgErrors.Wrap(MongoError, "user id:2 not found"),
			)

		_, ok := interface{}(group).(Codes)
		assert.Equal(t, true, ok)
	})

	t.Run("wrap", func(t *testing.T) {
		group := NewGroup(InvalidParams).
			AddChildren(
				pkgErrors.Wrap(MysqlError, "user id:1 not found"),
				pkgErrors.Wrap(MongoError, "user id:2 not found"),
			)

		err := pkgErrors.Wrap(group, "this is a test")
		assert.Equal(t, "this is a test: 1060003:参数错误", err.Error())
	})
}

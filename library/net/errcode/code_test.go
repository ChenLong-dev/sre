package errcode

import (
	"fmt"
	pkgErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCause(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		t.Run("code", func(t *testing.T) {
			e := NotFound
			result := Cause(e)
			assert.Equal(t, e.Code(), result.Code())
			assert.Equal(t, e.Message(), result.Message())
			assert.Equal(t, e.Details(), result.Details())
		})

		t.Run("group", func(t *testing.T) {
			t.Run("group", func(t *testing.T) {
				group := NewGroup(InvalidParams).
					AddChildren(
						pkgErrors.Wrap(MysqlError, "user id:1 not found"),
						pkgErrors.Wrap(MongoError, "user id:2 not found"),
					)
				result := Cause(group)
				assert.Equal(t, group.Code(), result.Code())
				assert.Equal(t, group.Message(), result.Message())
				assert.Equal(t, group.Details(), result.Details())
			})
		})
	})

	t.Run("empty", func(t *testing.T) {
		result := Cause(nil)
		assert.Equal(t, OK.Code(), result.Code())
		assert.Equal(t, OK.Message(), result.Message())
	})

	t.Run("new", func(t *testing.T) {
		e := pkgErrors.New(NotFound.Error())
		result := Cause(e)
		assert.Equal(t, NotFound.Code(), result.Code())
		assert.Equal(t, NotFound.Message(), result.Message())
		assert.Equal(t, NotFound.Details(), result.Details())
	})

	t.Run("wrap", func(t *testing.T) {
		t.Run("code", func(t *testing.T) {
			e := NotFound
			result := Cause(pkgErrors.Wrap(e, "this is a code"))
			assert.Equal(t, e.Code(), result.Code())
			assert.Equal(t, e.Message(), result.Message())
			assert.Equal(t, e.Details(), result.Details())
		})

		t.Run("group", func(t *testing.T) {
			group := NewGroup(InvalidParams).
				AddChildren(
					pkgErrors.Wrap(MysqlError, "user id:1 not found"),
					pkgErrors.Wrap(MongoError, "user id:2 not found"),
				)
			result := Cause(pkgErrors.Wrap(group, "this is a group"))
			assert.Equal(t, group.Code(), result.Code())
			assert.Equal(t, group.Message(), result.Message())
			assert.Equal(t, group.Details(), result.Details())
		})
	})
}

func TestEqualError(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		t.Run("code", func(t *testing.T) {
			e := NotFound
			result := EqualError(NotFound, e)
			assert.Equal(t, true, result)
		})

		t.Run("group", func(t *testing.T) {
			group := NewGroup(InvalidParams).
				AddChildren(
					pkgErrors.Wrap(MysqlError, "user id:1 not found"),
					pkgErrors.Wrap(MongoError, "user id:2 not found"),
				)
			result := EqualError(InvalidParams, group)
			assert.Equal(t, true, result)
		})
	})

	t.Run("new", func(t *testing.T) {
		e := pkgErrors.New(NotFound.Error())
		result := EqualError(NotFound, e)
		assert.Equal(t, true, result)
	})
}

func TestEqual(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		t.Run("code", func(t *testing.T) {
			e := NotFound
			result := Equal(NotFound, e)
			assert.Equal(t, true, result)
		})

		t.Run("group", func(t *testing.T) {
			group := NewGroup(InvalidParams).
				AddChildren(
					pkgErrors.Wrap(MysqlError, "user id:1 not found"),
					pkgErrors.Wrap(MongoError, "user id:2 not found"),
				)
			result := Equal(InvalidParams, group)
			assert.Equal(t, true, result)
		})
	})

	t.Run("nil", func(t *testing.T) {
		result := Equal(OK, nil)
		assert.Equal(t, true, result)
	})
}

func TestDetails(t *testing.T) {
	t.Run("code", func(t *testing.T) {
		err := InvalidParams

		code, ok := interface{}(err).(Codes)
		assert.Equal(t, true, ok)
		assert.Equal(t, 1, len(code.Details()))
		assert.Equal(t, "1060003:参数错误", code.Details()[0])
	})

	t.Run("wrap", func(t *testing.T) {
		err := pkgErrors.Wrap(InvalidParams, "user id:1")

		code := Cause(err)
		assert.Equal(t, 1, len(code.Details()))
		assert.Equal(t, "1060003:参数错误", code.Details()[0])
	})

	t.Run("group", func(t *testing.T) {
		group := NewGroup(InvalidParams).
			AddChildren(
				pkgErrors.Wrap(MysqlError, "user id:1 not found"),
				pkgErrors.Wrap(MongoError, "user id:2 not found"),
			)

		code, ok := interface{}(group).(Codes)
		assert.Equal(t, true, ok)
		assert.Equal(t, 2, len(code.Details()))
		assert.Equal(t, "user id:1 not found: 1060001:Mysql数据库错误", code.Details()[0])
		assert.Equal(t, "user id:2 not found: 1060002:Mongodb数据库错误", code.Details()[1])
	})
}

func TestGetErrorMessageMap(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		t.Run("code", func(t *testing.T) {
			err := InternalError

			result := GetErrorMessageMap(err)

			fmt.Printf("%#v\n", result)
		})

		t.Run("wrap", func(t *testing.T) {
			err := pkgErrors.Wrap(InternalError, "this is a test")

			result := GetErrorMessageMap(err)

			fmt.Printf("%#v\n", result)
		})

		t.Run("group", func(t *testing.T) {
			group := NewGroup(InvalidParams).
				AddChildren(
					pkgErrors.Wrap(MysqlError, "user id:1 not found"),
					pkgErrors.Wrap(MongoError, "user id:2 not found"),
				)

			result := GetErrorMessageMap(group)

			fmt.Printf("%#v\n", result)
		})
	})
}

func BenchmarkGetErrorMessageMap(b *testing.B) {
	b.Run("normal", func(b *testing.B) {
		err := pkgErrors.Wrap(InternalError, "this is a test")
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = GetErrorMessageMap(err)
		}
	})
}

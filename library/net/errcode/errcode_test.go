package errcode

import (
	"errors"
	"fmt"
	pkgErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestString(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		e := NotFound
		result := String(fmt.Sprintf("%s", e))
		assert.Equal(t, e.Code(), result.Code())
		assert.Equal(t, e.Message(), result.Message())
	})

	t.Run("invalid code", func(t *testing.T) {
		result := String(fmt.Sprintf("%s:%s", "test", "this is a test"))
		assert.Equal(t, InternalError.Code(), result.Code())
		assert.Equal(t, InternalError.Message(), result.Message())
	})

	t.Run("unregister code", func(t *testing.T) {
		result := String(fmt.Sprintf("%d:%s", 1111, "this is a test"))
		assert.Equal(t, InternalError.Code(), result.Code())
		assert.Equal(t, InternalError.Message(), result.Message())
	})
}

func TestErrCode_Error(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		e := NotFound
		result := fmt.Sprintf("%s", e)
		assert.Equal(t, "1040404:没有找到路由", result)
	})
}

func TestErrCode_WithMessage(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		e := NotFound
		f := NotFound.WithMessage("path found")
		assert.Equal(t, e.Code(), f.Code())
		assert.NotEqual(t, e.Message(), f.Message())
	})

	t.Run("override", func(t *testing.T) {
		e := NotFound
		f := e.WithMessage("path found")
		assert.Equal(t, e.Code(), f.Code())
		assert.NotEqual(t, e.Message(), f.Message())
	})
}

func TestErrCode_FrontendCode(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		e := NotFound
		f := NotFound.WithFrontCode(2040404)
		assert.Equal(t, e.Message(), f.Message())
		assert.NotEqual(t, e.FrontendCode(), f.FrontendCode())
	})

	t.Run("override", func(t *testing.T) {
		e := NotFound
		f := e.WithFrontCode(2040404)
		assert.Equal(t, e.Message(), f.Message())
		assert.NotEqual(t, e.FrontendCode(), f.FrontendCode())
	})
}

func TestErrCode_Is(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		e := fmt.Errorf("current error is %w", UnknownError)

		assert.Equal(t, true, errors.Is(e, UnknownError))
		assert.Equal(t, true, errors.Is(e, pkgErrors.Wrapf(UnknownError, "%s", "wrap of ")))
	})
}

func TestErrCode_As(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		e := fmt.Errorf("current error is %w", UnknownError)
		target := new(ErrCode)

		assert.Equal(t, true, errors.As(e, target))
		assert.Equal(t, UnknownError.Code(), target.Code())
	})
}

func TestGetOrNewErrCode(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		e := GetOrNewErrCode(0, "success")
		assert.Equal(t, 0, e.code)
		assert.Equal(t, 0, e.frontendCode)
		assert.Equal(t, "success", e.message)
	})

	t.Run("new", func(t *testing.T) {
		e := GetOrNewErrCode(9999999, "test")
		assert.Equal(t, 9999999, e.code)
		assert.Equal(t, 9999999, e.frontendCode)
		assert.Equal(t, "test", e.message)
	})

	t.Run("multi", func(t *testing.T) {
		e := GetOrNewErrCode(9999999, "test")
		assert.Equal(t, 9999999, e.code)
		assert.Equal(t, 9999999, e.frontendCode)
		assert.Equal(t, "test", e.message)
		_ = GetOrNewErrCode(9999999, "test")
	})
}

func TestErrCode_Clone(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		e := InternalError
		c := e.Clone()
		c = c.WithFrontCode(-1)
		c = c.WithMessage("test")

		assert.NotEqual(t, e.FrontendCode(), c.FrontendCode())
		assert.NotEqual(t, e.Message(), c.Message())
	})
}

func TestNew(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_ = New(2000001, "some message")
		})
	})

	t.Run("invalid", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = New(1999999, "some message")
		})
	})

	t.Run("already", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = New(0, "some message")
		})
	})
}

func TestFromStatusCode(t *testing.T) {
	t.Run("StatusPartialContent", func(t *testing.T) {
		c := FromStatusCode(http.StatusPartialContent)
		assert.Equal(t, http.StatusPartialContent, c.StatusCode())
		assert.Equal(t, 1020206, c.Code())
		assert.Equal(t, http.StatusText(http.StatusPartialContent), c.Message())
	})

	t.Run("exist", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = FromStatusCode(http.StatusNotFound)
		})
	})

	t.Run("invalid", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = FromStatusCode(999)
		})
	})
}

package response

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	pkgErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

type CustomResponse struct {
	Code         int         `json:"code"`
	Message      string      `json:"msg"`
	ErrorCode    int         `json:"errorno"`
	ErrorMessage string      `json:"errormsg"`
	Data         interface{} `json:"data"`

	ErrorString string `json:"-"`
}

func (r *CustomResponse) WithErrCode(code errcode.Codes) interface{} {
	return CustomResponse{
		Code:         200,
		Message:      "ok",
		ErrorCode:    code.Code(),
		ErrorMessage: code.Message(),
		Data:         r.Data,
	}
}

func (r *CustomResponse) PrintError(ctx context.Context, err error) {
	id := ctx.Value("id").(string)
	r.ErrorString = fmt.Sprintf("%s:%s", id, err)
}

func (r *CustomResponse) GetStatusCode(code errcode.Codes) int {
	switch code.Code() {
	case errcode.NoRowsFoundError.Code():
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}

func TestInjectJson(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		InjectJson(ctx, &CustomResponse{
			Data: "this is a data",
		}, nil)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"code":200,"msg":"ok","errorno":0,"errormsg":"success","data":"this is a data"}`), jsonStringBody)
	})

	t.Run("error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)
		ctx.Set("id", "test123456")

		response := CustomResponse{
			Data: "this is a data",
		}
		InjectJson(ctx, &response, errors.New("test error"))

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())
		assert.Equal(t, "test123456:test error", response.ErrorString)
	})

	t.Run("status_code", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)
		ctx.Set("id", "test123456")

		response := CustomResponse{
			Data: "this is a data",
		}
		InjectJson(ctx, &response, errcode.NoRowsFoundError)

		assert.Equal(t, http.StatusInternalServerError, ctx.Writer.Status())
	})
}

func TestJSON(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		JSON(ctx, "this is a data", nil)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":0,"errmsg":"success","data":"this is a data"}`), jsonStringBody)
	})

	t.Run("nil data", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		JSON(ctx, nil, nil)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":0,"errmsg":"success"}`), jsonStringBody)
	})

	t.Run("invalid params", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		InvalidParamsJSON(ctx, nil, errors.New("this is a test"))

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":1060003,"errmsg":"参数错误","errdetail":["1060003:参数错误"]}`), jsonStringBody)
	})

	t.Run("errcode group", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		group := errcode.NewGroup(errcode.InvalidParams).
			AddChildren(
				pkgErrors.Wrap(errcode.MysqlError, "user id:1 not found"),
				pkgErrors.Wrap(errcode.MongoError, "user id:2 not found"),
			)
		JSON(ctx, nil, group)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":1060003,"errmsg":"参数错误","errdetail":["user id:1 not found: 1060001:Mysql数据库错误","user id:2 not found: 1060002:Mongodb数据库错误"]}`), jsonStringBody)
	})

	t.Run("error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)
		ctx.Set("id", "test123456")

		JSON(ctx, nil, errors.New("test error"))

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	})

	t.Run("errors.Wrapf", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)
		ctx.Set("id", "test123456")

		JSON(ctx, nil, pkgErrors.Wrapf(errcode.UnknownError, "%s", "errors.Wrapf = "))

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	})

	t.Run("fmt.Errof", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)
		ctx.Set("id", "test123456")

		JSON(ctx, nil, fmt.Errorf("fmt.Errof = %w", errcode.UnknownError))

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	})

	t.Run("http_code", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		JSON(ctx, nil, errcode.NotFound)

		assert.Equal(t, http.StatusNotFound, ctx.Writer.Status())
	})

	t.Run("timeout", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		r, _ := http.NewRequest("GET", "/", nil)
		ctx.Request = r.WithContext(timeoutCtx)

		time.Sleep(time.Second)

		JSON(ctx, nil, errcode.NotFound)

		assert.NotEqual(t, http.StatusNotFound, ctx.Writer.Status())
	})
}

func TestStandardJSON(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		StandardJSON(ctx, "this is a data", nil)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":0,"errmsg":"success","data":"this is a data"}`), jsonStringBody)
	})

	t.Run("nil data", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		StandardJSON(ctx, nil, nil)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":0,"errmsg":"success"}`), jsonStringBody)
	})

	t.Run("invalid params", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		StandardInvalidParamsJSON(ctx, nil, errors.New("this is a test"))

		assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":1060003,"errmsg":"参数错误","errdetail":["1060003:参数错误"]}`), jsonStringBody)
	})

	t.Run("errcode group", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		group := errcode.NewGroup(errcode.InvalidParams).
			AddChildren(
				pkgErrors.Wrap(errcode.MysqlError, "user id:1 not found"),
				pkgErrors.Wrap(errcode.MongoError, "user id:2 not found"),
			)
		StandardJSON(ctx, nil, group)

		assert.Equal(t, http.StatusBadRequest, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":1060003,"errmsg":"参数错误","errdetail":["user id:1 not found: 1060001:Mysql数据库错误","user id:2 not found: 1060002:Mongodb数据库错误"]}`), jsonStringBody)
	})

	t.Run("error", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)
		ctx.Set("id", "test123456")

		StandardJSON(ctx, nil, errors.New("test error"))

		assert.Equal(t, http.StatusInternalServerError, ctx.Writer.Status())
	})

	t.Run("errors.Wrapf", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)
		ctx.Set("id", "test123456")

		StandardJSON(ctx, nil, pkgErrors.Wrapf(errcode.UnknownError, "%s", "errors.Wrapf = "))

		assert.Equal(t, http.StatusInternalServerError, ctx.Writer.Status())
	})

	t.Run("fmt.Errof", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)
		ctx.Set("id", "test123456")

		StandardJSON(ctx, nil, fmt.Errorf("fmt.Errof = %w", errcode.UnknownError))

		assert.Equal(t, http.StatusInternalServerError, ctx.Writer.Status())
	})

	t.Run("http_code", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		StandardJSON(ctx, nil, errcode.NotFound)

		assert.Equal(t, http.StatusNotFound, ctx.Writer.Status())
	})

	t.Run("timeout", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		r, _ := http.NewRequest("GET", "/", nil)
		ctx.Request = r.WithContext(timeoutCtx)

		time.Sleep(time.Second)

		StandardJSON(ctx, nil, errcode.NotFound)

		assert.NotEqual(t, http.StatusNotFound, ctx.Writer.Status())
	})
}

func TestStandardPaginationJSON(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		StandardPaginationJSON(ctx, 10, 0, 2, []int{1, 2}, nil)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":0,"errmsg":"success","data":{"total":10,"page":0,"pagesize":2,"items":[1,2]}}`), jsonStringBody)
	})

	t.Run("empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		StandardPaginationJSON(ctx, 0, 0, 10, []int{}, nil)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":0,"errmsg":"success","data":{"total":0,"page":0,"pagesize":10,"items":[]}}`), jsonStringBody)
	})

	t.Run("struct", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request, _ = http.NewRequest("GET", "/", nil)

		type testPagination struct {
			A int    `json:"a"`
			B string `json:"b"`
		}
		list := []*testPagination{
			{
				A: 1,
				B: "a",
			},
			{
				A: 2,
				B: "b",
			},
		}

		StandardPaginationJSON(ctx, 10, 0, 2, list, nil)

		assert.Equal(t, http.StatusOK, ctx.Writer.Status())

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(w.Body)
		assert.Nil(t, err)
		jsonStringBody := buf.String()
		assert.Equal(t, fmt.Sprint(`{"errcode":0,"errmsg":"success","data":{"total":10,"page":0,"pagesize":2,"items":[{"a":1,"b":"a"},{"a":2,"b":"b"}]}}`), jsonStringBody)
	})
}

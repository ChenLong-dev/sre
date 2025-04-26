package response

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func ExampleJSON() {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)

	StandardJSON(ctx, "this is a data", nil)

	fmt.Printf("%d\n", ctx.Writer.Status())
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(w.Body)
	if err != nil {
		return
	}
	jsonStringBody := buf.String()
	fmt.Printf("%s\n", jsonStringBody)

	// OutPut:
	// 200
	// {"errcode":0,"errmsg":"success","data":"this is a data"}
}

type CustomInjectResponse struct {
	Code         int         `json:"code"`
	Message      string      `json:"msg"`
	ErrorCode    int         `json:"errorno"`
	ErrorMessage string      `json:"errormsg"`
	Data         interface{} `json:"data"`
}

func (r *CustomInjectResponse) WithErrCode(code errcode.Codes) interface{} {
	return CustomResponse{
		Code:         200,
		Message:      "ok",
		ErrorCode:    code.FrontendCode(),
		ErrorMessage: code.Message(),
		Data:         r.Data,
	}
}

func (r *CustomInjectResponse) PrintError(ctx context.Context, err error) {
	id := ctx.Value("id").(string)
	fmt.Printf("%s:%s\n", id, err)
}

func (r *CustomInjectResponse) GetStatusCode(code errcode.Codes) int {
	switch code.Code() {
	case errcode.NotFound.Code():
		return http.StatusNotFound
	case errcode.ServiceUnavailable.Code():
		return http.StatusServiceUnavailable
	default:
		return http.StatusOK
	}
}

func CustomJson(ctx *gin.Context, data interface{}, err error) {
	InjectJson(ctx, &CustomInjectResponse{
		Data: data,
	}, err)
}

func ExampleInjectJson() {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", "/", nil)

	CustomJson(ctx, "this is a data", nil)

	fmt.Printf("%d\n", ctx.Writer.Status())
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(w.Body)
	if err != nil {
		return
	}
	jsonStringBody := buf.String()
	fmt.Printf("%s\n", jsonStringBody)

	// OutPut:
	// 200
	// {"code":200,"msg":"ok","errorno":0,"errormsg":"success","data":"this is a data"}
}

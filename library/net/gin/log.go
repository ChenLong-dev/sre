package gin

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_context "gitlab.shanhai.int/sre/library/base/context"
	"gitlab.shanhai.int/sre/library/base/filewriter"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	_infoFile  = "ginInfo.log"
	_errorFile = "ginError.log"

	defaultPattern = "%J{tbTUG}"

	defaultRouterPattern = "%J{tTMPNl}"

	QTHeaderUserIDKey      = "QT-User-Id"
	QTHeaderDeviceIDKey    = "Qt-Device-Id"
	QTHeaderAccessTokenKey = "QT-Access-Token"

	LogExtraKey = "LogExtraKey"
)

var (
	_logger *Logger
)

type LogWriter struct {
	io.Writer
	render.Render
}

type Logger struct {
	writers []LogWriter
}

func (logger Logger) Print(m map[string]interface{}) {
	for _, writer := range logger.writers {
		writer.Render.Render(writer.Writer, m)
	}
}

func (logger Logger) Close() {
	for _, writer := range logger.writers {
		writer.Render.Close()
	}
}

// 获取默认logger
func getDefaultRouterLogger(conf *Config) *Logger {
	if conf.Config == nil {
		conf.Config = &render.Config{}
	}

	if conf.StdoutRouterPattern == "" {
		conf.StdoutRouterPattern = defaultRouterPattern
	}
	if conf.OutRouterPattern == "" {
		conf.OutRouterPattern = defaultRouterPattern
	}

	var writers []LogWriter
	if conf.Stdout {
		writers = append(writers, LogWriter{
			Writer: os.Stdout,
			Render: render.NewPatternRender(patternMap, conf.StdoutRouterPattern),
		})
	}
	if conf.OutDir != "" {
		fw := filewriter.NewSingleFileWriter(
			filepath.Join(conf.OutDir, _infoFile),
			conf.FileBufferSize, conf.RotateSize, conf.MaxLogFile,
		)
		writers = append(writers, LogWriter{
			Writer: fw,
			Render: render.NewPatternRender(patternMap, conf.OutRouterPattern),
		})
	}

	return &Logger{
		writers: writers,
	}
}

// 获取info级别writer
func GetInfoWriter(conf *Config) io.Writer {
	return getGinWriter(conf, _infoFile)
}

// 获取error级别writer
func GetErrorWriter(conf *Config) io.Writer {
	return getGinWriter(conf, _errorFile)
}

// 获取writer
func getGinWriter(conf *Config, name string) io.Writer {
	if conf.Config == nil {
		conf.Config = &render.Config{}
	}

	var writers []io.Writer
	if conf.Stdout {
		writers = append(writers, os.Stdout)
	}
	if conf.OutDir != "" {
		fw := filewriter.NewSingleFileWriter(
			filepath.Join(conf.OutDir, name),
			conf.FileBufferSize, conf.RotateSize, conf.MaxLogFile,
		)
		writers = append(writers, fw)
	}

	return io.MultiWriter(writers...)
}

// 获取默认格式化渲染器
// todo:只能打印控制台或文件，不能同时打印
func GetDefaultFormatter(conf *Config) gin.HandlerFunc {
	if conf.Config == nil {
		conf.Config = &render.Config{}
	}

	if conf.StdoutPattern == "" {
		conf.StdoutPattern = defaultPattern
	}
	if conf.OutPattern == "" {
		conf.OutPattern = defaultPattern
	}

	ginRender := render.NewPatternRender(patternMap, conf.StdoutPattern)
	return func(c *gin.Context) {
		relativePath := GetGinRelativePath(c)

		var header http.Header
		var requestBody string
		request := c.Request
		if request != nil {
			header = request.Header
			if conf.RequestBodyOut {
				bodyBytes, err := ioutil.ReadAll(request.Body)
				if err != nil {
					goto SkipRequest
				}
				err = request.Body.Close()
				if err != nil {
					goto SkipRequest
				}
				request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
				requestBody = string(bodyBytes)
			}
		}

	SkipRequest:
		gin.LoggerWithFormatter(
			func(param gin.LogFormatterParams) string {
				msg := make(map[string]interface{})
				msg["end_time"] = param.TimeStamp
				msg["uuid"] = c.Value(_context.ContextUUIDKey)
				msg["status_code"] = param.StatusCode
				msg["latency"] = float64(param.Latency) / float64(time.Millisecond)
				msg["start_time"] = param.TimeStamp.Add(-param.Latency)
				msg["client_ip"] = param.ClientIP
				msg["method"] = param.Method
				msg["path"] = param.Path
				msg["relative_path"] = relativePath
				msg["error_message"] = param.ErrorMessage
				msg["header"] = header
				msg["request_body"] = requestBody
				msg["extra"] = c.Value(LogExtraKey)

				return ginRender.RenderString(msg)
			},
		)(c)
	}
}

func GetDefaultRouterPrintFunc(conf *Config) func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
	if conf == nil {
		conf = &Config{
			Config: &render.Config{
				Stdout: true,
			},
		}
	}
	if conf.Config == nil {
		conf.Config = &render.Config{}
	}

	_logger = getDefaultRouterLogger(conf)

	return func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		msg := make(map[string]interface{})
		msg["end_time"] = time.Now()
		msg["method"] = httpMethod
		msg["path"] = absolutePath
		msg["handler_name"] = handlerName
		msg["handler_length"] = nuHandlers

		_logger.Print(msg)
	}
}

// 注入自定义日志
func SetCustomLog(c *gin.Context, v interface{}) {
	c.Set(LogExtraKey, v)
}

var patternMap = map[string]render.PatternFunc{
	// 通用支持字符
	"T": render.PatternEndTime,
	"t": title,
	"M": method,
	"P": path,
	// 普通日志支持字符
	"U": render.PatternUUID,
	"G": ginExtra,
	"s": statusCode,
	"b": render.PatternStartTime,
	"L": latency,
	"C": clientIP,
	"E": errorMessage,
	"e": extra,
	"R": relativePath,
	"H": header,
	"B": requestBody,
	// 路由日志支持字符
	"N": handlerName,
	"l": handlerLength,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "GIN")
}

// 汇总的gin参数
func ginExtra(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("gin", []render.PatternFunc{
		statusCode, latency, clientIP, method,
		path, relativePath, header, requestBody, errorMessage,
		extra,
	})(args)
}

// 状态码
func statusCode(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("status_code", args["status_code"])
}

// 请求时间
func latency(args render.PatternArgs) render.PatternResult {
	originValue := args["latency"].(float64)
	duration, err := strconv.ParseFloat(fmt.Sprintf("%.2f", originValue), 64)
	if err != nil {
		duration = originValue
	}
	return render.NewPatternResult("latency", duration)
}

// 客户端ip
func clientIP(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("client_ip", args["client_ip"])
}

// 请求方法
func method(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("method", args["method"])
}

// 请求路径
func path(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("path", args["path"])
}

// 错误信息
func errorMessage(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("error_message", args["error_message"])
}

// 自定义参数
func extra(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("extra", args["extra"])
}

// 路由方法
func relativePath(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("relative_path", args["relative_path"])
}

// 请求头
func header(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("header", args["header"])
}

// 请求体
func requestBody(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("request_body", args["request_body"])
}

// 路由处理器名称
func handlerName(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("handler_name", args["handler_name"])
}

// 路由经过的处理器数量
func handlerLength(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("handler_length", args["handler_length"])
}

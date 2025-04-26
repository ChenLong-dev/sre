package tracing

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gitlab.shanhai.int/sre/library/base/filewriter"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	_infoFile = "jaegerInfo.log"

	ErrorLevel = "ERROR"
	InfoLevel  = "INFO"

	defaultPattern = "%J{tTE}"
)

type LogWriter struct {
	io.Writer
	render.Render
}

type Logger struct {
	writers []LogWriter
}

// 实现jaeger中logger接口
func (logger Logger) Infof(msg string, args ...interface{}) {
	m := make(map[string]interface{})
	m["level"] = InfoLevel
	m["message"] = fmt.Sprintf(msg, args...)
	for _, writer := range logger.writers {
		writer.Render.Render(writer.Writer, m)
	}
}

func (logger Logger) Error(msg string) {
	m := make(map[string]interface{})
	m["level"] = ErrorLevel
	m["message"] = msg
	for _, writer := range logger.writers {
		writer.Render.Render(writer.Writer, m)
	}
}

// 获取默认writer
func getDefaultWriter(conf *Config) *Logger {
	if conf.StdoutPattern == "" {
		conf.StdoutPattern = defaultPattern
	}
	if conf.OutPattern == "" {
		conf.OutPattern = defaultPattern
	}

	var writers []LogWriter
	if conf.Stdout {
		writers = append(writers, LogWriter{
			Writer: os.Stdout,
			Render: render.NewPatternRender(patternMap, conf.StdoutPattern),
		})
	}
	if conf.OutDir != "" {
		fw := filewriter.NewSingleFileWriter(
			filepath.Join(conf.OutDir, _infoFile),
			conf.FileBufferSize, conf.RotateSize, conf.MaxLogFile,
		)
		writers = append(writers, LogWriter{
			Writer: fw,
			Render: render.NewPatternRender(patternMap, conf.OutPattern),
		})
	}

	return &Logger{
		writers: writers,
	}
}

var patternMap = map[string]render.PatternFunc{
	"T": render.PatternTime,
	"t": title,
	"E": tracingExtra,
	"L": level,
	"M": message,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "OPENTRACING")
}

// 汇总的tracing参数
func tracingExtra(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("opentracing", []render.PatternFunc{level, message})(args)
}

// 日志级别
func level(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("level", args["level"])
}

// 日志消息
func message(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("message", args["message"])
}

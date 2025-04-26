package sentry

import (
	"io"
	"os"
	"path/filepath"

	"gitlab.shanhai.int/sre/library/base/filewriter"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	_infoFile      = "sentry.log"
	defaultPattern = "%J{tLSs}"
)

type logger interface {
	Print(m map[string]interface{})
	Close()
}

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
func getDefaultLogger(config *Config) *Logger {
	var writers []LogWriter
	if config.Stdout {
		writers = append(writers, LogWriter{
			Writer: os.Stdout,
			Render: render.NewPatternRender(patternMap, config.StdoutPattern),
		})
	}
	if config.OutDir != "" {
		fw := filewriter.NewSingleFileWriter(
			filepath.Join(config.OutDir, config.OutFile),
			config.FileBufferSize, config.RotateSize, config.MaxLogFile,
		)
		writers = append(writers, LogWriter{
			Writer: fw,
			Render: render.NewPatternRender(patternMap, config.OutPattern),
		})
	}

	return &Logger{
		writers: writers,
	}
}

// 渲染模版
var patternMap = map[string]render.PatternFunc{
	"L": render.PatternTime,
	"S": source,
	"t": title,
	"D": DSN,
	"T": tags,
	"e": eventID,
	"E": environment,
	"s": sentryParams,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "sentry")
}

// 日志打印的调用源
func source(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("source", args["source"])
}

// DSN
func DSN(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("DSN", args["DSN"])
}

// Sentry的标签
func tags(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("tags", args["tags"])
}

// 事件ID
func eventID(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("eventID", args["eventID"])
}

// 开发环境
func environment(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("env", args["env"])
}

// Sentry参数汇总
func sentryParams(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("sentry", []render.PatternFunc{
		DSN, tags, eventID, environment,
	})(args)
}

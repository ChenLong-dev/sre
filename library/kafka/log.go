package kafka

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gitlab.shanhai.int/sre/library/base/filewriter"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	_infoFile = "kafkaInfo.log"

	defaultPattern = "%J{tTK}"
)

type logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Close()
}

type LogWriter struct {
	io.Writer
	render.Render
}

type Logger struct {
	writers []LogWriter
}

func (logger Logger) Print(v ...interface{}) {
	m := make(map[string]interface{})
	m["message"] = v

	for _, writer := range logger.writers {
		writer.Render.Render(writer.Writer, m)
	}
}

func (logger Logger) Printf(format string, v ...interface{}) {
	m := make(map[string]interface{})
	m["message"] = []string{fmt.Sprintf(format, v...)}

	for _, writer := range logger.writers {
		writer.Render.Render(writer.Writer, m)
	}
}

func (logger Logger) Println(v ...interface{}) {
	m := make(map[string]interface{})
	m["message"] = v

	for _, writer := range logger.writers {
		writer.Render.Render(writer.Writer, m)
	}
}

func (logger Logger) Close() {
	for _, writer := range logger.writers {
		writer.Render.Close()
	}
}

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
	"K": kafkaExtra,
	"M": message,
}

// 日志标题
func title(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("title", "Kafka")
}

// 汇总的kafka参数
func kafkaExtra(args render.PatternArgs) render.PatternResult {
	return render.AggregatePatternFunc("kafka", []render.PatternFunc{message})(args)
}

// 消息
func message(args render.PatternArgs) render.PatternResult {
	return render.NewPatternResult("message", args["message"])
}

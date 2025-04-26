package hook

import (
	"io"
	"os"
	"path/filepath"

	"gitlab.shanhai.int/sre/library/base/filewriter"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 日志记录器接口
type Logger interface {
	// 打印日志
	Print(m map[string]interface{})
	// 关闭日志
	Close()
}

// 简单日志写入
type SampleLogWriter struct {
	io.Writer
	render.Render
}

// 简单日志记录器
type SampleLogger struct {
	writers []SampleLogWriter
}

func (logger SampleLogger) Print(m map[string]interface{}) {
	for _, writer := range logger.writers {
		writer.Render.Render(writer.Writer, m)
	}
}

func (logger SampleLogger) Close() {
	for _, writer := range logger.writers {
		writer.Render.Close()
	}
}

// 获取默认日志记录器
func GetDefaultLogger(conf *render.Config, patternMap map[string]render.PatternFunc) Logger {
	var writers []SampleLogWriter
	if conf.Stdout {
		writers = append(writers, SampleLogWriter{
			Writer: os.Stdout,
			Render: render.NewPatternRender(patternMap, conf.StdoutPattern),
		})
	}
	if conf.OutDir != "" {
		fw := filewriter.NewSingleFileWriter(
			filepath.Join(conf.OutDir, conf.OutFile),
			conf.FileBufferSize, conf.RotateSize, conf.MaxLogFile,
		)
		writers = append(writers, SampleLogWriter{
			Writer: fw,
			Render: render.NewPatternRender(patternMap, conf.OutPattern),
		})
	}

	return &SampleLogger{
		writers: writers,
	}
}

package log

import (
	"context"
	"io"
	"path/filepath"

	"gitlab.shanhai.int/sre/library/base/filewriter"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 不同日志级别对应文件索引
const (
	_infoIdx = iota
	_warnIdx
	_errorIdx
	_totalIdx
)

// 日志文件名
var _fileNames = map[int]string{
	_infoIdx:  "info.log",
	_warnIdx:  "warning.log",
	_errorIdx: "error.log",
}

// 文件处理器
type FileHandler struct {
	render render.Render
	// 文件writer
	fws [_totalIdx]*filewriter.FileWriter
}

// 新建文件处理器
func NewFile(customPattern string, dir string, bufferSize, rotateSize int64, maxLogFile int) *FileHandler {
	if customPattern == "" {
		customPattern = defaultPattern
	}
	handler := &FileHandler{
		render: render.NewPatternRender(patternMap, customPattern),
	}
	for idx, name := range _fileNames {
		handler.fws[idx] = filewriter.NewSingleFileWriter(
			filepath.Join(dir, name), bufferSize, rotateSize, maxLogFile)
	}
	return handler
}

func (h *FileHandler) Log(ctx context.Context, lv Level, args map[string]interface{}) {
	// 增加额外参数
	addExtraField(ctx, args)

	// 写入不同级别文件
	var w io.Writer
	switch lv {
	case _warnLevel:
		w = h.fws[_warnIdx]
	case _errorLevel:
		w = h.fws[_errorIdx]
	default:
		w = h.fws[_infoIdx]
	}
	h.render.Render(w, args)
}

func (h *FileHandler) Close() error {
	for _, fw := range h.fws {
		fw.Close()
	}
	h.render.Close()
	return nil
}

func (h *FileHandler) SetFormat(format string) {
	h.render = render.NewPatternRender(patternMap, format)
}

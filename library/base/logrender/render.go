package render

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"time"
)

// 渲染接口
type Render interface {
	// 渲染写入writer中
	Render(io.Writer, map[string]interface{}) error
	// 获取渲染字符串
	RenderString(map[string]interface{}) string
	// 关闭渲染器，异步定时写入时必须
	Close() error
}

// 渲染buffer结构体
type renderBuffer struct {
	// 数据buffer
	buffer *bytes.Buffer
	// 写入writer
	writer io.Writer
}

// 新建渲染器
// 	patternMap为用于渲染的函数字典，key为用于渲染的格式化字符，值为渲染函数
//	format为渲染的模版
func NewPatternRender(patternMap map[string]PatternFunc, format string) Render {
	// J为保留格式化字符，不可使用
	if _, ok := patternMap["J"]; ok {
		panic("pattern map shouldn't use 'J'")
	}

	p := &patternRender{
		renderBufPool: sync.Pool{
			New: func() interface{} {
				return &renderBuffer{
					buffer: new(bytes.Buffer),
				}
			},
		},
		// 缓冲区管道大小应根据服务请求量配置，以防止阻塞
		renderBufChan: make(chan *renderBuffer, 1024),
		// 如开启异步写入，刷新时间应在1s以上，否则会因为频繁写入，更加影响打印时间
		flushTime: 0 * time.Millisecond,
	}

	b := make([]byte, 0, len(format))
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			b = append(b, format[i])
			continue
		}
		if i+1 >= len(format) {
			b = append(b, format[i])
			continue
		}

		var curFunc func(PatternArgs, *bytes.Buffer) (bool, string, interface{})
		// 渲染json
		if format[i+1] == 'J' && i+2 < len(format) && format[i+2] == '{' {
			jsonFuncArray := make([]PatternFunc, 0)
			for i = i + 2; i+1 < len(format) && format[i+1] != '}'; i++ {
				curFormat := string(format[i+1])
				f, ok := patternMap[curFormat]
				if !ok {
					continue
				}
				jsonFuncArray = append(jsonFuncArray, f)
			}
			curFunc = jsonFormatFactory(jsonFuncArray)
		} else { // 普通渲染
			f, ok := patternMap[string(format[i+1])]
			if !ok {
				b = append(b, format[i])
				continue
			}
			curFunc = convertFunc(f)
		}
		// 非格式化字符，使用普通文本渲染
		if len(b) != 0 {
			p.funcs = append(p.funcs, textFactory(string(b)))
			b = b[:0]
		}

		p.funcs = append(p.funcs, curFunc)
		i++
	}

	if len(b) != 0 {
		p.funcs = append(p.funcs, textFactory(string(b)))
	}

	// 若为定时异步写入，启动守护协程
	if p.flushTime > 0 {
		p.wg.Add(1)
		go p.daemon()
	}

	return p
}

// 外部函数转换
func convertFunc(unwrapFunc PatternFunc) func(PatternArgs, *bytes.Buffer) (bool, string, interface{}) {
	return func(args PatternArgs, buffer *bytes.Buffer) (isRender bool, key string, value interface{}) {
		res := unwrapFunc(args)
		if res.IsSkip() {
			return true, res.Key, res.Value
		}
		return false, res.Key, res.Value
	}
}

// 普通文本渲染
func textFactory(text string) func(PatternArgs, *bytes.Buffer) (bool, string, interface{}) {
	return func(PatternArgs, *bytes.Buffer) (bool, string, interface{}) {
		return false, "", text
	}
}

// json渲染map对象池
var jsonFormatMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{})
	},
}

// json渲染buffer对象池
var jsonFormatBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// json渲染
func jsonFormatFactory(funArray []PatternFunc) func(PatternArgs, *bytes.Buffer) (bool, string, interface{}) {
	return func(args PatternArgs, buf *bytes.Buffer) (bool, string, interface{}) {
		jsonMap := jsonFormatMapPool.Get().(map[string]interface{})
		for k := range jsonMap {
			delete(jsonMap, k)
		}

		jsonBuf := jsonFormatBufferPool.Get().(*bytes.Buffer)
		jsonBuf.Reset()

		for _, f := range funArray {
			res := f(args)
			if res.Key != "" && !res.IsSkip() {
				jsonMap[res.Key] = res.Value
			}
		}
		err := json.NewEncoder(jsonBuf).Encode(jsonMap)
		if err != nil {
			return false, "json", jsonMap
		}

		// todo:json encoder方法会自动追加换行符，真的自作聪明
		//  为了去掉换行符，所以使用额外的buffer做中间转换，有额外内存开销
		b := jsonBuf.Bytes()
		buf.Write(b[:len(b)-1])

		jsonFormatMapPool.Put(jsonMap)
		jsonFormatBufferPool.Put(jsonBuf)

		// 已处理，无需返回value
		return true, "json", nil
	}
}

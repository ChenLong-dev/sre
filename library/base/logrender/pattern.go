package render

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// 模版函数
type PatternFunc func(args PatternArgs) PatternResult

// 汇总模版函数
func AggregatePatternFunc(aggKey string, patternList []PatternFunc) PatternFunc {
	return func(args PatternArgs) PatternResult {
		m := make(map[string]interface{})

		for _, f := range patternList {
			res := f(args)
			if !res.IsSkip() {
				m[res.Key] = res.Value
			}
		}

		return NewPatternResult(aggKey, m)
	}
}

// 模版参数
type PatternArgs map[string]interface{}

// 获取参数或默认值
func (args PatternArgs) GetOrDefault(key string, defaultValue interface{}) interface{} {
	if value, ok := args[key]; !ok {
		return defaultValue
	} else {
		return value
	}
}

// 模版函数结果字典
type PatternResultMap map[string]interface{}

// 新建模版函数结果字典
func NewPatternResultMap() PatternResultMap {
	return make(map[string]interface{})
}

// 添加结果
func (m PatternResultMap) Add(r PatternResult) PatternResultMap {
	m[r.Key] = r.Value
	return m
}

// 模版函数结果
type PatternResult struct {
	// json的key
	Key string
	// 值
	Value interface{}
	// 是否跳过渲染
	skipRender bool
}

// 新建模版函数结果
func NewPatternResult(key string, value interface{}) PatternResult {
	return PatternResult{
		Key:        key,
		Value:      value,
		skipRender: false,
	}
}

// 新建默认模版结果
func DefaultPatternResult() PatternResult {
	return PatternResult{
		skipRender: true,
	}
}

// 是否跳过
func (r PatternResult) IsSkip() bool {
	return r.skipRender
}

// time类型值
func (r PatternResult) TimeValue() time.Time {
	t, ok := r.Value.(time.Time)
	if ok {
		return t
	}
	return time.Time{}
}

// bool类型值
func (r PatternResult) BoolValue() bool {
	b, ok := r.Value.(bool)
	if ok {
		return b
	}
	return false
}

// int类型值
func (r PatternResult) IntValue() int {
	i, ok := r.Value.(int)
	if ok {
		return i
	}
	return 0
}

// int32类型值
func (r PatternResult) Int32Value() int32 {
	i, ok := r.Value.(int32)
	if ok {
		return i
	}
	return 0
}

// int64类型值
func (r PatternResult) Int64Value() int64 {
	i, ok := r.Value.(int64)
	if ok {
		return i
	}
	return 0
}

// string类型值
func (r PatternResult) StringValue() string {
	s, ok := r.Value.(string)
	if ok {
		return s
	}
	return ""
}

// float64类型值
func (r PatternResult) Float64Value() float64 {
	f, ok := r.Value.(float64)
	if ok {
		return f
	}
	return 0
}

// 模版渲染器
type patternRender struct {
	// 渲染函数
	funcs []func(args PatternArgs, buf *bytes.Buffer) (isRender bool, key string, value interface{})
	// 渲染buffer缓冲池
	renderBufPool sync.Pool
	// 协程wait group
	wg sync.WaitGroup
	// 渲染buffer管道
	renderBufChan chan *renderBuffer
	// 是否关闭
	closed int32
	// 定时异步写入时间
	flushTime time.Duration
}

// 定时异步写入守护方法
// 	异步写入相对于同步写入大约减少20%CPU运行时间，内存占用大约额外提高30%
// 	具体对比数值要根据实际刷新时间来决定
//	目前暂不建议开启异步写入，cpu提升不明显，且占用更高内存
//	绝大部分日志库也不带异步写入功能，所以开启后，也较难注入进其他依赖库中
func (p *patternRender) daemon() {
	// 待写入buffer区
	sumBuf := renderBuffer{
		buffer: new(bytes.Buffer),
	}
	// 写入定时器
	writeTicker := time.NewTicker(p.flushTime)
	var err error
	for {
		select {
		// 从缓冲区读取数据塞入待写入区中，并重用buf
		case buf, ok := <-p.renderBufChan:
			if ok {
				sumBuf.writer = buf.writer
				sumBuf.buffer.Write(buf.buffer.Bytes())

				buf.buffer.Reset()
				buf.writer = nil
				p.renderBufPool.Put(buf)
			}
		// 从待写入区读取并写入，并重置待写入区
		case <-writeTicker.C:
			if sumBuf.buffer.Len() > 0 {
				if _, err = sumBuf.buffer.WriteTo(sumBuf.writer); err != nil {
					fmt.Printf("%#v\n", err)
				}

				sumBuf.buffer.Reset()
				sumBuf.writer = nil
			}
		}
		// 检查是否关闭
		if atomic.LoadInt32(&p.closed) != 1 {
			continue
		}
		// 关闭，写入剩下的数据
		if _, err = sumBuf.buffer.WriteTo(sumBuf.writer); err != nil {
			fmt.Printf("%#v\n", err)
		}
		for buf := range p.renderBufChan {
			if _, err = buf.buffer.WriteTo(buf.writer); err != nil {
				fmt.Printf("%#v\n", err)
			}

			buf.buffer.Reset()
			buf.writer = nil
			p.renderBufPool.Put(buf)
		}
		break
	}
	p.wg.Done()
}

// 渲染写入writer中
func (p *patternRender) Render(w io.Writer, d map[string]interface{}) error {
	buf := p.renderBufPool.Get().(*renderBuffer)

	for _, f := range p.funcs {
		isRender, _, v := f(d, buf.buffer)
		if !isRender {
			buf.buffer.WriteString(fmt.Sprint(v))
		}
	}
	buf.buffer.WriteString("\n")
	buf.writer = w

	var err error

	if p.flushTime > 0 {
		// 定时异步写入
		p.renderBufChan <- buf
	} else {
		// 立即写入
		_, err = buf.buffer.WriteTo(w)

		buf.buffer.Reset()
		p.renderBufPool.Put(buf)
	}

	return err
}

// 获取渲染字符串
func (p *patternRender) RenderString(d map[string]interface{}) string {
	buf := p.renderBufPool.Get().(*renderBuffer)

	for _, f := range p.funcs {
		isRender, _, v := f(d, buf.buffer)
		if !isRender {
			buf.buffer.WriteString(fmt.Sprint(v))
		}
	}
	buf.buffer.WriteString("\n")

	return buf.buffer.String()
}

// 关闭渲染器，异步定时写入时必须
func (p *patternRender) Close() error {
	atomic.StoreInt32(&p.closed, 1)
	close(p.renderBufChan)
	p.wg.Wait()
	return nil
}

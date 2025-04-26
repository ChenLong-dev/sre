package filewriter

import (
	"bytes"
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// FileWriter
type FileWriter struct {
	// 配置
	opt option
	// 文件目录
	dir string
	// 文件名
	fname string
	// 管道缓冲区
	ch chan *bytes.Buffer
	// 日志logger
	stdlog *log.Logger
	// buffer对象池
	pool *sync.Pool

	// 最新的文件分割时间格式
	lastRotateFormat string
	// 最新的文件分割编号
	lastSplitNum int

	// 当前分割文件
	current *wrapFile
	// 当前目录下的分割文件信息列表
	files *list.List

	// 用于表示文件是否关闭写入的标志符
	closed int32
	// 用于内部协程的waitgroup
	wg sync.WaitGroup
}

// 每个分割文件的相关信息
type rotateItem struct {
	rotateTime int64
	rotateNum  int
	fname      string
}

// 获取当前目录下符合条件的分割文件信息列表
func parseRotateItem(dir, fname, rotateFormat string) (*list.List, error) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// parse exists log file filename
	parse := func(s string) (rt rotateItem, err error) {
		rt.fname = s

		// 去除文件前的 '.' error.log.2018-09-12.001 -> 2018-09-12.001
		s = strings.TrimLeft(s[len(fname):], ".")

		seqs := strings.Split(s, ".")
		var t time.Time
		switch len(seqs) {
		case 2:
			// 文件编号后缀
			if rt.rotateNum, err = strconv.Atoi(seqs[1]); err != nil {
				return
			}
			fallthrough
		case 1:
			// 文件时间后缀
			if t, err = time.Parse(rotateFormat, seqs[0]); err != nil {
				return
			}
			rt.rotateTime = t.Unix()
		}
		return
	}

	// 查找当前目录下文件，并获取符合条件的分割信息
	var items []rotateItem
	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), fname) && fi.Name() != fname {
			rt, err := parse(fi.Name())
			if err != nil {
				// todo:处理错误
				continue
			}
			items = append(items, rt)
		}
	}

	// 排序
	sort.Slice(items, func(i, j int) bool {
		if items[i].rotateTime == items[j].rotateTime {
			return items[i].rotateNum > items[j].rotateNum
		}
		return items[i].rotateTime > items[j].rotateTime
	})

	l := list.New()
	for _, item := range items {
		l.PushBack(item)
	}

	return l, nil
}

// 每个单个文件
type wrapFile struct {
	// 文件大小
	fsize int64
	// 文件本身
	fp *os.File
}

// 文件大小
func (w *wrapFile) size() int64 {
	return w.fsize
}

// 写入文件
func (w *wrapFile) write(p []byte) (n int, err error) {
	n, err = w.fp.Write(p)
	w.fsize += int64(n)
	return
}

// 新建单个文件
func newWrapFile(fpath string) (*wrapFile, error) {
	fp, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	fi, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	return &wrapFile{fp: fp, fsize: fi.Size()}, nil
}

// 新建基础的FileWriter
// 	dir为文件位置，bufferSize为缓冲区大小，rotateSize为分割每个文件大小，maxLogFile为最大日志数量
func NewSingleFileWriter(dir string, bufferSize, rotateSize int64, maxLogFile int) *FileWriter {
	var options []OptionFunc
	if rotateSize > 0 {
		options = append(options, MaxSize(rotateSize))
	}
	if maxLogFile > 0 {
		options = append(options, MaxFile(maxLogFile))
	}
	w, err := New(dir, options...)
	if err != nil {
		panic(err)
	}
	return w
}

// 新建FileWriter
func New(fpath string, fns ...OptionFunc) (*FileWriter, error) {
	opt := defaultOption
	for _, fn := range fns {
		fn(&opt)
	}

	fname := filepath.Base(fpath)
	if fname == "" {
		return nil, fmt.Errorf("filename can't empty")
	}
	dir := filepath.Dir(fpath)
	fi, err := os.Stat(dir)
	if err == nil && !fi.IsDir() {
		return nil, fmt.Errorf("%s already exists and not a directory", dir)
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create dir %s error: %s", dir, err.Error())
		}
	}

	current, err := newWrapFile(fpath)
	if err != nil {
		return nil, err
	}

	// 用于打印错误日志
	stdlog := log.New(os.Stderr, "flog ", log.LstdFlags)

	files, err := parseRotateItem(dir, fname, opt.RotateFormat)
	if err != nil {
		// set files a empty list
		files = list.New()
		stdlog.Printf("parseRotateItem error: %s", err)
	}

	// 获取最晚的编号
	lastRotateFormat := time.Now().Format(opt.RotateFormat)
	var lastSplitNum int
	if files.Len() > 0 {
		rt := files.Front().Value.(rotateItem)
		//  check contains is mush esay than compared with timestamp
		if strings.Contains(rt.fname, lastRotateFormat) {
			lastSplitNum = rt.rotateNum
		}
	}

	ch := make(chan *bytes.Buffer, opt.ChanSize)
	fw := &FileWriter{
		opt:    opt,
		dir:    dir,
		fname:  fname,
		stdlog: stdlog,
		ch:     ch,
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},

		lastSplitNum:     lastSplitNum,
		lastRotateFormat: lastRotateFormat,

		files:   files,
		current: current,
	}

	fw.wg.Add(1)
	// 开启写入及检查协程
	go fw.daemon()

	return fw, nil
}

// 实现io.Writer写入接口
func (f *FileWriter) Write(p []byte) (int, error) {
	// 检查文件是否关闭
	if atomic.LoadInt32(&f.closed) == 1 {
		f.stdlog.Printf("%s", p)
		return 0, fmt.Errorf("filewriter already closed")
	}
	// 由于写入是异步操作，为防止写入时数据发生变化，将其先存入buf中
	buf := f.getBuf()
	buf.Write(p)

	if f.opt.WriteTimeout == 0 {
		select {
		case f.ch <- buf:
			return len(p), nil
		// 缓冲区被占用，返回错误
		default:
			return 0, fmt.Errorf("log channel is full, discard log")
		}
	}

	timeout := time.NewTimer(f.opt.WriteTimeout)
	select {
	case f.ch <- buf:
		return len(p), nil
	// 写入超时，返回错误
	case <-timeout.C:
		return 0, fmt.Errorf("log channel is full, discard log")
	}
}

// 写入及检查分割操作
func (f *FileWriter) daemon() {
	// todo:检查aggsbuf大小，防止过大
	// 待写入区
	aggsbuf := &bytes.Buffer{}
	tk := time.NewTicker(f.opt.RotateInterval)
	// todo:可配置aggstk
	// 待写入区写入延迟
	aggstk := time.NewTicker(10 * time.Millisecond)
	var err error
	for {
		select {
		// 定时检查分割
		case t := <-tk.C:
			f.checkRotate(t)
		// 从缓冲区读取数据塞入待写入区中，并重用buf
		case buf, ok := <-f.ch:
			if ok {
				aggsbuf.Write(buf.Bytes())
				f.putBuf(buf)
			}
		// 从待写入区读取并写入文件，并重置待写入区
		case <-aggstk.C:
			if aggsbuf.Len() > 0 {
				if err = f.write(aggsbuf.Bytes()); err != nil {
					f.stdlog.Printf("write log error: %s", err)
				}
				aggsbuf.Reset()
			}
		}
		// 检查文件是否关闭，如未关闭则继续
		if atomic.LoadInt32(&f.closed) != 1 {
			continue
		}
		// 文件关闭，写入剩下的数据
		if err = f.write(aggsbuf.Bytes()); err != nil {
			f.stdlog.Printf("write log error: %s", err)
		}
		for buf := range f.ch {
			if err = f.write(buf.Bytes()); err != nil {
				f.stdlog.Printf("write log error: %s", err)
			}
			f.putBuf(buf)
		}
		break
	}
	f.wg.Done()
}

// 关闭文件写入
func (f *FileWriter) Close() error {
	atomic.StoreInt32(&f.closed, 1)
	close(f.ch)
	f.wg.Wait()
	return nil
}

// 检查文件分割
func (f *FileWriter) checkRotate(t time.Time) {
	formatFname := func(format string, num int) string {
		if num == 0 {
			return fmt.Sprintf("%s.%s", f.fname, format)
		}
		return fmt.Sprintf("%s.%s.%03d", f.fname, format, num)
	}
	format := t.Format(f.opt.RotateFormat)

	if f.opt.MaxFile != 0 {
		// 若大于最大文件数，则删除文件
		for f.files.Len() > f.opt.MaxFile {
			rt := f.files.Remove(f.files.Front()).(rotateItem)
			fpath := filepath.Join(f.dir, rt.fname)
			if err := os.Remove(fpath); err != nil {
				f.stdlog.Printf("remove file %s error: %s", fpath, err)
			}
		}
	}

	// 检查条件并分割文件
	if format != f.lastRotateFormat || (f.opt.MaxSize != 0 && f.current.size() > f.opt.MaxSize) {
		var err error
		// 关闭当前文件，防止写入
		if err = f.current.fp.Close(); err != nil {
			f.stdlog.Printf("close current file error: %s", err)
		}

		// 重命名文件
		fname := formatFname(f.lastRotateFormat, f.lastSplitNum)
		oldpath := filepath.Join(f.dir, f.fname)
		newpath := filepath.Join(f.dir, fname)
		if err = os.Rename(oldpath, newpath); err != nil {
			f.stdlog.Printf("rename file %s to %s error: %s", oldpath, newpath, err)
			return
		}

		// 塞入新文件信息
		f.files.PushBack(rotateItem{fname: fname})

		// 若是新的时间，则更改时间，重置编号
		// 否则增加标号
		if format != f.lastRotateFormat {
			f.lastRotateFormat = format
			f.lastSplitNum = 0
		} else {
			f.lastSplitNum++
		}

		// 重新创建当前文件
		f.current, err = newWrapFile(filepath.Join(f.dir, f.fname))
		if err != nil {
			f.stdlog.Printf("create log file error: %s", err)
		}
	}
}

// 写入数据
func (f *FileWriter) write(p []byte) error {
	if f.current == nil {
		f.stdlog.Printf("can't write log to file, please check stderr log for detail")
		f.stdlog.Printf("%s", p)
	}
	_, err := f.current.write(p)
	return err
}

// 放入对象池，用于重用
func (f *FileWriter) putBuf(buf *bytes.Buffer) {
	buf.Reset()
	f.pool.Put(buf)
}

// 获取对象池中可重用的buffer
func (f *FileWriter) getBuf() *bytes.Buffer {
	return f.pool.Get().(*bytes.Buffer)
}

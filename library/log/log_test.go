package log

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func initStdout() {
	conf := &Config{
		Config: &render.Config{
			Stdout: true,
		},
	}
	Init(conf)
}

func initFile() {
	conf := &Config{
		Config: &render.Config{
			OutDir: "/tmp",
		},
	}
	Init(conf)
}

type TestLog struct {
	A string
	B int
	C string
	D string
}

func testLog(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		Error("hello %s", "world")
		Errorv(context.Background(), map[string]interface{}{
			"key":   2222222,
			"test2": "test",
		})
		Errorc(context.Background(), "keys: %s %s...", "key1", "key2")
	})
	t.Run("Warn", func(t *testing.T) {
		Warn("hello %s", "world")
		Warnv(context.Background(), map[string]interface{}{
			"key":   2222222,
			"test2": "test",
		})
		Warnc(context.Background(), "keys: %s %s...", "key1", "key2")
	})
	t.Run("Info", func(t *testing.T) {
		Info("hello %s", "world")
		Infov(context.Background(), map[string]interface{}{
			"key":   2222222,
			"test2": "test",
		})
		Infoc(context.Background(), "keys: %s %s...", "key1", "key2")
	})
}

func TestFile(t *testing.T) {
	initFile()
	testLog(t)
	assert.Equal(t, nil, Close())
}

func TestStdout(t *testing.T) {
	initStdout()
	testLog(t)
	assert.Equal(t, nil, Close())
}

func TestMinLevel(t *testing.T) {
	Init(&Config{
		Config: &render.Config{
			Stdout: true,
		},
		V: int(_warnLevel),
	})
	t.Run("Info", func(t *testing.T) {
		Info("hello %s", "world")
		Infov(context.Background(), map[string]interface{}{
			"key":   2222222,
			"test2": "test",
		})
		Infoc(context.Background(), "keys: %s %s...", "key1", "key2")
	})
	t.Run("Warn", func(t *testing.T) {
		Warn("hello %s", "world")
		Warnv(context.Background(), map[string]interface{}{
			"key":   2222222,
			"test2": "test",
		})
		Warnc(context.Background(), "keys: %s %s...", "key1", "key2")
	})
}

func TestFilter(t *testing.T) {
	Init(&Config{
		Config: &render.Config{
			Stdout: true,
		},
		Filter: []string{
			"key",
		},
	})
	t.Run("Info", func(t *testing.T) {
		Info("hello %s", "world")
		Infov(context.Background(), map[string]interface{}{
			"key":   2222222,
			"test2": "test",
		})
		Infoc(context.Background(), "keys: %s %s...", "key1", "key2")
	})
}

func TestOverwriteSource(t *testing.T) {
	ctx := context.Background()
	t.Run("test source kv string", func(t *testing.T) {
		Infov(ctx, map[string]interface{}{
			"source": "test",
		})
	})
	t.Run("test source kv string", func(t *testing.T) {
		Infov(ctx, map[string]interface{}{
			"source": "test",
		})
	})
}

func BenchmarkLog(b *testing.B) {
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Infov(ctx, map[string]interface{}{
				"int":  34,
				"test": "hello",
				"hhh":  "hhhh",
			})
		}
	})
}

func BenchmarkLogrus(b *testing.B) {
	b.Run("logrus", func(b *testing.B) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)
		logrus.SetReportCaller(true)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logrus.Info("this is a message")
		}
	})

	b.Run("library", func(b *testing.B) {
		ctx := context.Background()
		conf := &Config{
			Config: &render.Config{
				Stdout: true,
			},
		}
		Init(conf)
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			Infov(ctx, map[string]interface{}{
				"message": "this is a message",
			})
		}

		b.StopTimer()
		err := Close()
		if err != nil {
			fmt.Printf("%#v\n", err)
		}
	})
}

func BenchmarkLogrusMemmory(b *testing.B) {
	ctx := context.Background()
	conf := &Config{
		Config: &render.Config{
			Stdout: true,
		},
	}
	Init(conf)

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetReportCaller(true)

	b.ResetTimer()

	b.Run("library", func(b *testing.B) {
		for i := 0; i < 10000; i++ {
			Infov(ctx, map[string]interface{}{
				"message": "this is a message",
			})
		}
	})

	b.Run("logrus", func(b *testing.B) {
		for i := 0; i < 10000; i++ {
			logrus.Info("this is a message")
		}
	})
}

func BenchmarkParallelInfoc(b *testing.B) {
	conf := &Config{
		Config: &render.Config{
			Stdout: true,
		},
	}

	type Response struct {
		Code    int    `json:"errcode"`
		Message string `json:"errmsg"`
	}

	Init(conf)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := Response{
				Code:    10,
				Message: "abc",
			}
			Infoc(context.Background(), "%#v", r)
		}
	})

	b.StopTimer()
}

package render

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func BenchmarkPattern_Render(b *testing.B) {
	t, err := time.Parse("2006/01/02 15:04:05", "2019/01/02 11:36:28")
	if err != nil {
		return
	}
	args := make(map[string]interface{})
	args["time"] = t

	r := NewPatternRender(patternMap, "%J{Tt}")
	for i := 0; i < 200000; i++ {
		buf := &bytes.Buffer{}
		err = r.Render(buf, args)
		if err != nil {
			return
		}
		fmt.Printf("%s\n", buf.String())
	}
	err = r.Close()
	if err != nil {
		return
	}
}

func BenchmarkPattern_RenderString(b *testing.B) {
	t, err := time.Parse("2006/01/02 15:04:05", "2019/01/02 11:36:28")
	if err != nil {
		return
	}
	args := make(map[string]interface{})
	args["time"] = t

	r := NewPatternRender(patternMap, "%J{Tt}")
	for i := 0; i < 200000; i++ {
		s := r.RenderString(args)
		fmt.Printf("%s\n", s)
	}
	err = r.Close()
	if err != nil {
		return
	}
}

func BenchmarkPattern_RenderStringNotJSON(b *testing.B) {
	t, err := time.Parse("2006/01/02 15:04:05", "2019/01/02 11:36:28")
	if err != nil {
		return
	}
	args := make(map[string]interface{})
	args["time"] = t

	r := NewPatternRender(patternMap, "%T %t")
	for i := 0; i < 200000; i++ {
		s := r.RenderString(args)
		fmt.Printf("%s\n", s)
	}
	err = r.Close()
	if err != nil {
		return
	}
}

func BenchmarkPattern_RenderNotJSON(b *testing.B) {
	t, err := time.Parse("2006/01/02 15:04:05", "2019/01/02 11:36:28")
	if err != nil {
		return
	}
	args := make(map[string]interface{})
	args["time"] = t

	r := NewPatternRender(patternMap, "%T %t")
	for i := 0; i < 200000; i++ {
		buf := &bytes.Buffer{}
		err = r.Render(buf, args)
		if err != nil {
			return
		}
		fmt.Printf("%s\n", buf.String())
	}
	err = r.Close()
	if err != nil {
		return
	}
}

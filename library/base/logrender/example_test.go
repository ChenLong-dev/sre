package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

var patternMap = map[string]PatternFunc{
	"T": longTime,
	"t": title,
	"R": render,
}

func longTime(args PatternArgs) PatternResult {
	return NewPatternResult("time", args["time"].(time.Time).Format("2006/01/02 15:04:05.000"))
}

func title(args PatternArgs) PatternResult {
	return NewPatternResult("title", "RENDER")
}

func render(args PatternArgs) PatternResult {
	m := make(map[string]interface{})

	funcArray := []PatternFunc{longTime, title}
	for _, f := range funcArray {
		res := f(args)
		m[res.Key] = res.Value
	}

	jsonLog, _ := json.Marshal(m)

	return NewPatternResult("render", fmt.Sprintf("%s", jsonLog))
}

func ExamplePattern_Render() {
	t, err := time.Parse("2006/01/02 15:04:05", "2019/01/02 11:36:28")
	if err != nil {
		return
	}
	args := make(map[string]interface{})
	args["time"] = t
	buf := &bytes.Buffer{}

	r := NewPatternRender(patternMap, "%J{TtR}")
	err = r.Render(buf, args)
	if err != nil {
		return
	}
	err = r.Close()
	if err != nil {
		return
	}
	fmt.Printf("%s\n", buf.String())

	// OutPut:
	// {"render":"{\"time\":\"2019/01/02 11:36:28.000\",\"title\":\"RENDER\"}","time":"2019/01/02 11:36:28.000","title":"RENDER"}
}

func ExamplePattern_RenderString() {
	t, err := time.Parse("2006/01/02 15:04:05", "2019/01/02 11:36:28")
	if err != nil {
		return
	}
	args := make(map[string]interface{})
	args["time"] = t

	r := NewPatternRender(patternMap, "[%T] title:%t %J{R}")
	s := r.RenderString(args)
	err = r.Close()
	if err != nil {
		return
	}
	fmt.Printf("%s\n", s)

	// OutPut:
	// [2019/01/02 11:36:28.000] title:RENDER {"render":"{\"time\":\"2019/01/02 11:36:28.000\",\"title\":\"RENDER\"}"}

}

package log

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	render "gitlab.shanhai.int/sre/library/base/logrender"

	"github.com/stretchr/testify/assert"
)

func TestPatternDefault(t *testing.T) {
	buf := &bytes.Buffer{}
	p := render.NewPatternRender(patternMap, "%L %T %M")
	p.Render(buf, map[string]interface{}{_level: _infoLevel.String(), _log: "hello", _time: time.Now().Format(_timeFormat), _source: "xxx:123"})
	p.Close()

	fields := strings.Fields(buf.String())
	assert.Equal(t, 4, len(fields))
	assert.Equal(t, "INFO", fields[0])
	assert.Equal(t, "hello", fields[3])
}

func TestJsonPatternDefault(t *testing.T) {
	buf := &bytes.Buffer{}
	p := render.NewPatternRender(patternMap, "%J{LTSm}")
	p.Render(buf, map[string]interface{}{_level: _infoLevel.String(), _log: "hello", _time: time.Now().Format(_timeFormat), _source: "xxx:123"})
	p.Close()

	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(buf.String()), &jsonMap)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(jsonMap))
	assert.Equal(t, "INFO", jsonMap["level"])
	assert.Equal(t, "hello", jsonMap["message"].(map[string]interface{})["log"])
}

func TestGoodSymbol(t *testing.T) {
	buf := &bytes.Buffer{}
	p := render.NewPatternRender(patternMap, "%M")
	p.Render(buf, map[string]interface{}{_level: _infoLevel.String(), _log: "2233", "hello": "test"})
	p.Close()

	assert.Equal(t, "hello=test 2233\n", buf.String())
}

func TestBadSymbol(t *testing.T) {
	buf := &bytes.Buffer{}
	p := render.NewPatternRender(patternMap, "%12 %% %xd %M")
	p.Render(buf, map[string]interface{}{_level: _infoLevel.String(), _log: "2233"})
	p.Close()

	assert.Equal(t, "%12 %% %xd 2233\n", buf.String())
}

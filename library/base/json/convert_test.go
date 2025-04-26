package json

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestJsonStruct struct {
	A string  `json:"a"`
	B bool    `json:"b"`
	C int     `json:"c"`
	D float64 `json:"d"`

	Nil   string `json:"-"`
	Empty string `json:"empty,omitempty"`
}

func TestStructToJsonMap(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		s := TestJsonStruct{
			A: "abc",
			B: true,
			C: 123,
			D: 1.23,
		}
		m := StructToJsonMap(s)
		assert.Equal(t, s.A, m["a"])
		assert.Equal(t, s.B, m["b"])
		assert.Equal(t, s.C, m["c"])
		assert.Equal(t, s.D, m["d"])
	})

	t.Run("nil", func(t *testing.T) {
		s := TestJsonStruct{
			Nil: "abc",
		}
		m := StructToJsonMap(s)
		_, ok := m["Nil"]
		assert.Equal(t, false, ok)
	})

	t.Run("empty", func(t *testing.T) {
		s := TestJsonStruct{
			Empty: "abc",
		}
		m := StructToJsonMap(s)
		assert.Equal(t, s.Empty, m["empty"])

		s = TestJsonStruct{
			Empty: "",
		}
		m = StructToJsonMap(s)
		_, ok := m["empty"]
		assert.Equal(t, false, ok)
	})
}

func BenchmarkStructToJsonMap(b *testing.B) {
	b.Run("library", func(b *testing.B) {
		s := TestJsonStruct{
			A: "abc",
			B: true,
			C: 123,
			D: 1.23,
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			m := StructToJsonMap(s)
			m["dynamic"] = m["a"]
			delete(m, "a")
			_, _ = json.Marshal(m)
		}
	})

	b.Run("json", func(b *testing.B) {
		s := TestJsonStruct{
			A: "abc",
			B: true,
			C: 123,
			D: 1.23,
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			m := make(map[string]interface{})
			j, _ := json.Marshal(s)
			_ = json.Unmarshal(j, &m)
			m["dynamic"] = m["a"]
			delete(m, "a")
			_, _ = json.Marshal(m)
		}
	})
}

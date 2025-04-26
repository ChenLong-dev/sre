package httpclient

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestHeader_Add(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		h := NewJsonHeader().
			Add("User-Agent",
				"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 "+
					"(KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36")
		assert.Equal(t, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36", h.Header.Get("User-Agent"))
	})

	t.Run("multi", func(t *testing.T) {
		h := NewJsonHeader().
			Add("User-Agent",
				"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 "+
					"(KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36").
			Add("User-Agent",
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
					"(KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")
		assert.Equal(t, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36", h.Header.Get("User-Agent"))
	})
}

func TestHeader_Set(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		h := NewJsonHeader().
			Set("User-Agent",
				"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 "+
					"(KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36")
		assert.Equal(t, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36", h.Header.Get("User-Agent"))
	})

	t.Run("multi", func(t *testing.T) {
		h := NewJsonHeader().
			Set("User-Agent",
				"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 "+
					"(KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36").
			Set("User-Agent",
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
					"(KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")
		assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36", h.Header.Get("User-Agent"))
	})
}

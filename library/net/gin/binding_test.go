package gin

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
	"gitlab.shanhai.int/sre/library/base/null"
)

type nullInStruct struct {
	A null.String `form:"a" json:"a,omitempty"`
	B null.Time   `form:"b" json:"b,omitempty"`
	C null.Bool   `form:"c" json:"c,omitempty"`
	D null.Float  `form:"d" json:"d,omitempty"`
	E null.Int    `form:"e" json:"e,omitempty"`
	F null.Int64  `form:"f" json:"f,omitempty"`
	G null.Time   `form:"g" json:"g,omitempty" time_format:"2006-01-0215:04:05"`
	H null.Time   `form:"h" json:"h,omitempty" time_format:"unix"`
	I null.String `form:"-" json:"i,omitempty"`
}

type normalInStruct struct {
	A string     `form:"a" json:"a,omitempty"`
	B *time.Time `form:"b" json:"b,omitempty"`
	C bool       `form:"c" json:"c,omitempty"`
	D float64    `form:"d" json:"d,omitempty"`
	E int        `form:"e" json:"e,omitempty"`
	F int64      `form:"f" json:"f,omitempty"`
	G *time.Time `form:"g" json:"g,omitempty" time_format:"2006-01-0215:04:05"`
	H *time.Time `form:"h" json:"h,omitempty" time_format:"unix"`
	I string     `form:"-" json:"i,omitempty"`
	J *int       `form:"j" json:"j,omitempty"`
}

type anonymousStruct struct {
	Test null.String `form:"test" json:"test,omitempty"`
	nullInStruct
}

func TestNullQuery(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		router := gin.New()
		router.GET("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, "test", s.A.ValueOrZero())
			assert.Equal(t, true, s.B.ValueOrZero().Equal(time.Date(2019, 11, 28, 7, 16, 12, 0, time.UTC)))
			assert.Equal(t, true, s.C.ValueOrZero())
			assert.Equal(t, 1.23, s.D.ValueOrZero())
			assert.Equal(t, 123, s.E.ValueOrZero())
			assert.Equal(t, int64(234), s.F.ValueOrZero())
			assert.Equal(t, false, s.I.Valid)

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?a=test&b=2019-11-28T07:16:12.000Z&c=true&d=1.23&e=123&f=234&i=test", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("empty", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, true, s.A.Valid)
			assert.Equal(t, false, s.B.Valid)
			assert.Equal(t, false, s.C.Valid)
			assert.Equal(t, false, s.D.Valid)
			assert.Equal(t, false, s.E.Valid)
			assert.Equal(t, false, s.F.Valid)

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/?a=&b=&c=&d=&e=&f=", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("null", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, false, s.A.Valid)
			assert.Equal(t, false, s.B.Valid)
			assert.Equal(t, false, s.C.Valid)
			assert.Equal(t, false, s.D.Valid)
			assert.Equal(t, false, s.E.Valid)
			assert.Equal(t, false, s.F.Valid)

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("time_format", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, true, s.G.ValueOrZero().Equal(time.Date(2019, 11, 29, 15, 42, 0, 0, time.Local)))
			assert.Equal(t, true, s.H.ValueOrZero().Equal(time.Date(2019, 12, 3, 0, 0, 0, 0, time.Local)))

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/?g=2019-11-2915:42:00&h=1575302400", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("anonymous", func(t *testing.T) {
		router := gin.New()
		router.GET("/", func(c *gin.Context) {
			s := new(anonymousStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, "123", s.Test.ValueOrZero())
			assert.Equal(t, "test", s.A.ValueOrZero())
			assert.Equal(t, true, s.B.ValueOrZero().Equal(time.Date(2019, 11, 28, 7, 16, 12, 0, time.UTC)))
			assert.Equal(t, true, s.C.ValueOrZero())
			assert.Equal(t, 1.23, s.D.ValueOrZero())
			assert.Equal(t, 123, s.E.ValueOrZero())
			assert.Equal(t, int64(234), s.F.ValueOrZero())

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?test=123&a=test&b=2019-11-28T07:16:12.000Z&c=true&d=1.23&e=123&f=234", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestNullFormPost(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, FormPost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, "test", s.A.ValueOrZero())
			assert.Equal(t, true, s.B.ValueOrZero().Equal(time.Date(2019, 11, 28, 7, 16, 12, 0, time.UTC)))
			assert.Equal(t, true, s.C.ValueOrZero())
			assert.Equal(t, 1.23, s.D.ValueOrZero())
			assert.Equal(t, 123, s.E.ValueOrZero())
			assert.Equal(t, int64(234), s.F.ValueOrZero())

			c.JSON(http.StatusOK, nil)
		})

		forms := make(url.Values)
		forms.Add("a", "test")
		forms.Add("b", "2019-11-28T07:16:12.000Z")
		forms.Add("c", "true")
		forms.Add("d", "1.23")
		forms.Add("e", "123")
		forms.Add("f", "234")
		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", nil, nil, forms)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("empty", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, FormPost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, true, s.A.Valid)
			assert.Equal(t, false, s.B.Valid)
			assert.Equal(t, false, s.C.Valid)
			assert.Equal(t, false, s.D.Valid)
			assert.Equal(t, false, s.E.Valid)
			assert.Equal(t, false, s.F.Valid)

			c.JSON(http.StatusOK, nil)
		})

		forms := make(url.Values)
		forms.Add("a", "")
		forms.Add("b", "")
		forms.Add("c", "")
		forms.Add("d", "")
		forms.Add("e", "")
		forms.Add("f", "")
		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", nil, nil, forms)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("null", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, FormPost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, false, s.A.Valid)
			assert.Equal(t, false, s.B.Valid)
			assert.Equal(t, false, s.C.Valid)
			assert.Equal(t, false, s.D.Valid)
			assert.Equal(t, false, s.E.Valid)
			assert.Equal(t, false, s.F.Valid)

			c.JSON(http.StatusOK, nil)
		})

		forms := make(url.Values)
		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", nil, nil, forms)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestNormalQuery(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		router := gin.New()
		router.GET("/", func(c *gin.Context) {
			s := new(normalInStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, "test", s.A)
			assert.Equal(t, true, s.B.Equal(time.Date(2019, 11, 28, 7, 16, 12, 0, time.UTC)))
			assert.Equal(t, true, s.C)
			assert.Equal(t, 1.23, s.D)
			assert.Equal(t, 123, s.E)
			assert.Equal(t, int64(234), s.F)
			assert.Equal(t, "", s.I)
			assert.Equal(t, true, s.J != nil)
			assert.Equal(t, 123, *s.J)

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?a=test&b=2019-11-28T07:16:12.000Z&c=true&d=1.23&e=123&f=234&i=test&j=123", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("empty", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(normalInStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, "", s.A)
			assert.Equal(t, true, s.B.IsZero())
			assert.Equal(t, false, s.C)
			assert.Equal(t, float64(0), s.D)
			assert.Equal(t, 0, s.E)
			assert.Equal(t, int64(0), s.F)
			assert.Equal(t, true, s.J != nil)

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/?a=&b=&c=&d=&e=&f=&j=", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("null", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(normalInStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, "", s.A)
			assert.Nil(t, s.B)
			assert.Equal(t, false, s.C)
			assert.Equal(t, float64(0), s.D)
			assert.Equal(t, 0, s.E)
			assert.Equal(t, int64(0), s.F)
			assert.Equal(t, false, s.J != nil)

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("time_format", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(normalInStruct)
			err := c.ShouldBindWith(s, Query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, true, s.G.Equal(time.Date(2019, 11, 29, 15, 42, 0, 0, time.Local)))
			assert.Equal(t, true, s.H.Equal(time.Date(2019, 12, 3, 0, 0, 0, 0, time.Local)))

			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/?g=2019-11-2915:42:00&h=1575302400", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestNullHeaderPost(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, Header)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, "test", s.A.ValueOrZero())
			assert.Equal(t, true, s.B.ValueOrZero().Equal(time.Date(2019, 11, 28, 7, 16, 12, 0, time.UTC)))
			assert.Equal(t, true, s.C.ValueOrZero())
			assert.Equal(t, 1.23, s.D.ValueOrZero())
			assert.Equal(t, 123, s.E.ValueOrZero())
			assert.Equal(t, int64(234), s.F.ValueOrZero())

			c.JSON(http.StatusOK, nil)
		})

		headers := http.Header{}
		headers.Add("a", "test")
		headers.Add("b", "2019-11-28T07:16:12.000Z")
		headers.Add("c", "true")
		headers.Add("d", "1.23")
		headers.Add("e", "123")
		headers.Add("f", "234")
		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", headers, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("empty", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, Header)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, true, s.A.Valid)
			assert.Equal(t, false, s.B.Valid)
			assert.Equal(t, false, s.C.Valid)
			assert.Equal(t, false, s.D.Valid)
			assert.Equal(t, false, s.E.Valid)
			assert.Equal(t, false, s.F.Valid)

			c.JSON(http.StatusOK, nil)
		})

		headers := http.Header{}
		headers.Add("a", "")
		headers.Add("b", "")
		headers.Add("c", "")
		headers.Add("d", "")
		headers.Add("e", "")
		headers.Add("f", "")
		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", headers, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("null", func(t *testing.T) {
		router := gin.New()
		router.POST("/", func(c *gin.Context) {
			s := new(nullInStruct)
			err := c.ShouldBindWith(s, Header)
			if err != nil {
				c.JSON(http.StatusInternalServerError, nil)
				return
			}
			fmt.Printf("%#v\n", s)
			assert.Equal(t, false, s.A.Valid)
			assert.Equal(t, false, s.B.Valid)
			assert.Equal(t, false, s.C.Valid)
			assert.Equal(t, false, s.D.Valid)
			assert.Equal(t, false, s.E.Valid)
			assert.Equal(t, false, s.F.Valid)

			c.JSON(http.StatusOK, nil)
		})

		headers := http.Header{}
		r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", headers, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestNullForm(t *testing.T) {
	t.Run("query", func(t *testing.T) {
		t.Run("all", func(t *testing.T) {
			router := gin.New()
			router.GET("/", func(c *gin.Context) {
				s := new(nullInStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, "test", s.A.ValueOrZero())
				assert.Equal(t, true, s.B.ValueOrZero().Equal(time.Date(2019, 11, 28, 7, 16, 12, 0, time.UTC)))
				assert.Equal(t, true, s.C.ValueOrZero())
				assert.Equal(t, 1.23, s.D.ValueOrZero())
				assert.Equal(t, 123, s.E.ValueOrZero())
				assert.Equal(t, int64(234), s.F.ValueOrZero())
				assert.Equal(t, false, s.I.Valid)

				c.JSON(http.StatusOK, nil)
			})

			r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?a=test&b=2019-11-28T07:16:12.000Z&c=true&d=1.23&e=123&f=234&i=test", nil, nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})

		t.Run("empty", func(t *testing.T) {
			router := gin.New()
			router.POST("/", func(c *gin.Context) {
				s := new(nullInStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, true, s.A.Valid)
				assert.Equal(t, false, s.B.Valid)
				assert.Equal(t, false, s.C.Valid)
				assert.Equal(t, false, s.D.Valid)
				assert.Equal(t, false, s.E.Valid)
				assert.Equal(t, false, s.F.Valid)

				c.JSON(http.StatusOK, nil)
			})

			r, err := httpUtil.TestGinJsonRequest(router, "POST", "/?a=&b=&c=&d=&e=&f=", nil, nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})

		t.Run("null", func(t *testing.T) {
			router := gin.New()
			router.POST("/", func(c *gin.Context) {
				s := new(nullInStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, false, s.A.Valid)
				assert.Equal(t, false, s.B.Valid)
				assert.Equal(t, false, s.C.Valid)
				assert.Equal(t, false, s.D.Valid)
				assert.Equal(t, false, s.E.Valid)
				assert.Equal(t, false, s.F.Valid)

				c.JSON(http.StatusOK, nil)
			})

			r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", nil, nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})

		t.Run("time_format", func(t *testing.T) {
			router := gin.New()
			router.POST("/", func(c *gin.Context) {
				s := new(nullInStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, true, s.G.ValueOrZero().Equal(time.Date(2019, 11, 29, 15, 42, 0, 0, time.Local)))
				assert.Equal(t, true, s.H.ValueOrZero().Equal(time.Date(2019, 12, 3, 0, 0, 0, 0, time.Local)))

				c.JSON(http.StatusOK, nil)
			})

			r, err := httpUtil.TestGinJsonRequest(router, "POST", "/?g=2019-11-2915:42:00&h=1575302400", nil, nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})

		t.Run("anonymous", func(t *testing.T) {
			router := gin.New()
			router.GET("/", func(c *gin.Context) {
				s := new(anonymousStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, "123", s.Test.ValueOrZero())
				assert.Equal(t, "test", s.A.ValueOrZero())
				assert.Equal(t, true, s.B.ValueOrZero().Equal(time.Date(2019, 11, 28, 7, 16, 12, 0, time.UTC)))
				assert.Equal(t, true, s.C.ValueOrZero())
				assert.Equal(t, 1.23, s.D.ValueOrZero())
				assert.Equal(t, 123, s.E.ValueOrZero())
				assert.Equal(t, int64(234), s.F.ValueOrZero())

				c.JSON(http.StatusOK, nil)
			})

			r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?test=123&a=test&b=2019-11-28T07:16:12.000Z&c=true&d=1.23&e=123&f=234", nil, nil, nil)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})
	})

	t.Run("form", func(t *testing.T) {
		t.Run("all", func(t *testing.T) {
			router := gin.New()
			router.GET("/", func(c *gin.Context) {
				s := new(nullInStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, "test", s.A.ValueOrZero())
				assert.Equal(t, true, s.B.ValueOrZero().Equal(time.Date(2019, 11, 28, 7, 16, 12, 0, time.UTC)))
				assert.Equal(t, true, s.C.ValueOrZero())
				assert.Equal(t, 1.23, s.D.ValueOrZero())
				assert.Equal(t, 123, s.E.ValueOrZero())
				assert.Equal(t, int64(234), s.F.ValueOrZero())
				assert.Equal(t, false, s.I.Valid)

				c.JSON(http.StatusOK, nil)
			})

			fb := url.Values{}
			fb.Add("a", "test")
			fb.Add("b", "2019-11-28T07:16:12.000Z")
			fb.Add("c", "true")
			fb.Add("d", "1.23")
			fb.Add("e", "123")
			fb.Add("f", "234")
			fb.Add("i", "test")
			r, err := httpUtil.TestGinJsonRequest(router, "GET", "/",
				nil, nil, fb)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})

		t.Run("empty", func(t *testing.T) {
			router := gin.New()
			router.POST("/", func(c *gin.Context) {
				s := new(nullInStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, true, s.A.Valid)
				assert.Equal(t, false, s.B.Valid)
				assert.Equal(t, false, s.C.Valid)
				assert.Equal(t, false, s.D.Valid)
				assert.Equal(t, false, s.E.Valid)
				assert.Equal(t, false, s.F.Valid)

				c.JSON(http.StatusOK, nil)
			})

			fb := url.Values{}
			fb.Add("a", "")
			fb.Add("b", "")
			fb.Add("c", "")
			fb.Add("d", "")
			fb.Add("e", "")
			fb.Add("f", "")
			r, err := httpUtil.TestGinJsonRequest(router, "POST", "/",
				nil, nil, fb)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})

		t.Run("null", func(t *testing.T) {
			router := gin.New()
			router.POST("/", func(c *gin.Context) {
				s := new(nullInStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, false, s.A.Valid)
				assert.Equal(t, false, s.B.Valid)
				assert.Equal(t, false, s.C.Valid)
				assert.Equal(t, false, s.D.Valid)
				assert.Equal(t, false, s.E.Valid)
				assert.Equal(t, false, s.F.Valid)

				c.JSON(http.StatusOK, nil)
			})

			fb := url.Values{}
			r, err := httpUtil.TestGinJsonRequest(router, "POST", "/", nil, nil, fb)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})

		t.Run("time_format", func(t *testing.T) {
			router := gin.New()
			router.POST("/", func(c *gin.Context) {
				s := new(nullInStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, true, s.G.ValueOrZero().Equal(time.Date(2019, 11, 29, 15, 42, 0, 0, time.Local)))
				assert.Equal(t, true, s.H.ValueOrZero().Equal(time.Date(2019, 12, 3, 0, 0, 0, 0, time.Local)))

				c.JSON(http.StatusOK, nil)
			})

			fb := url.Values{}
			fb.Add("g", "2019-11-2915:42:00")
			fb.Add("h", "1575302400")
			r, err := httpUtil.TestGinJsonRequest(router, "POST", "/",
				nil, nil, fb)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})

		t.Run("anonymous", func(t *testing.T) {
			router := gin.New()
			router.GET("/", func(c *gin.Context) {
				s := new(anonymousStruct)
				err := c.ShouldBindWith(s, Form)
				if err != nil {
					c.JSON(http.StatusInternalServerError, nil)
					return
				}
				fmt.Printf("%#v\n", s)
				assert.Equal(t, "123", s.Test.ValueOrZero())
				assert.Equal(t, "test", s.A.ValueOrZero())
				assert.Equal(t, true, s.B.ValueOrZero().Equal(time.Date(2019, 11, 28, 7, 16, 12, 0, time.UTC)))
				assert.Equal(t, true, s.C.ValueOrZero())
				assert.Equal(t, 1.23, s.D.ValueOrZero())
				assert.Equal(t, 123, s.E.ValueOrZero())
				assert.Equal(t, int64(234), s.F.ValueOrZero())

				c.JSON(http.StatusOK, nil)
			})

			fb := url.Values{}
			fb.Add("test", "123")
			fb.Add("a", "test")
			fb.Add("b", "2019-11-28T07:16:12.000Z")
			fb.Add("c", "true")
			fb.Add("d", "1.23")
			fb.Add("e", "123")
			fb.Add("f", "234")
			r, err := httpUtil.TestGinJsonRequest(router, "GET", "/",
				nil, nil, fb)
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, r.Code)
		})
	})
}

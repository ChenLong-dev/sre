package sql

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestStruct struct {
	CreatedAt time.Time
	UpdatedAt time.Time  `gorm:"column:update_time;convertible"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	Desc      string     `gorm:"column:desc"`
	Title     string     `gorm:"column:title;convertible"`
	TestAnonymousStruct
}

type TestAnonymousStruct struct {
	AnonymousDesc  string `gorm:"column:anonymous_desc"`
	AnonymousTitle string `gorm:"column:anonymous_title;convertible"`
}

func TestStructToGORMMap(t *testing.T) {
	t.Run("column", func(t *testing.T) {
		s := TestStruct{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Desc:      "test",
			TestAnonymousStruct: TestAnonymousStruct{
				AnonymousDesc: "test",
			},
		}
		m := StructToGORMMap(s)

		assert.NotNil(t, m["CreatedAt"])
		assert.NotNil(t, m["update_time"])
		assert.NotNil(t, m["desc"])
		assert.NotNil(t, m["anonymous_desc"])
	})

	t.Run("empty", func(t *testing.T) {
		s := TestStruct{
			CreatedAt:           time.Now(),
			Desc:                "test",
			Title:               "",
			TestAnonymousStruct: TestAnonymousStruct{},
		}
		m := StructToGORMMap(s)

		assert.NotNil(t, m["CreatedAt"])
		assert.NotNil(t, m["update_time"])
		assert.NotNil(t, m["desc"])
		assert.NotNil(t, m["title"])
		assert.NotNil(t, m["anonymous_title"])
		assert.Nil(t, m["deleted_at"])
	})
}

package sql

import (
	"github.com/pkg/errors"
	render "gitlab.shanhai.int/sre/library/base/logrender"

	// database driver
	_ "github.com/go-sql-driver/mysql"
)

// 新建mysql客户端
func NewMySQL(c *Config) (db *OrmDB) {
	if c == nil {
		panic("mysql config is nil")
	}

	if c.Config == nil {
		c.Config = &render.Config{}
	}
	if c.Config.StdoutPattern == "" {
		c.Config.StdoutPattern = defaultPattern
	}
	if c.Config.OutPattern == "" {
		c.Config.OutPattern = defaultPattern
	}
	if c.Config.OutFile == "" {
		c.Config.OutFile = _InfoFile
	}

	db, err := OpenOrm(c)
	if err != nil {
		panic(errors.Wrap(err, "open mysql error"))
	}
	return
}

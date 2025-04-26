package mongo

import (
	"context"

	"github.com/pkg/errors"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 新建mongo客户端
func NewMongo(c *Config) (db *DB) {
	if c == nil {
		panic("mongo config is nil")
	}
	if c.DSN == nil {
		panic("mongo must be set dsn")
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
		c.Config.OutFile = _infoFile
	}

	db, err := Open(c)
	if err != nil {
		panic(errors.Wrap(err, "open mongo error"))
	}

	err = db.Ping(context.Background())
	if err != nil {
		panic(errors.Wrap(err, "mongo health check error"))
	}

	return
}

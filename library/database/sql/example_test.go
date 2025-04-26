package sql

import (
	"context"
	"time"

	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func ExampleNewMySQL() {
	c := Config{
		DSN: &DSNConfig{
			UserName: "root",
			Password: "123456",
			Endpoint: &EndpointConfig{
				Address: "rm-xxxxxxxxxxxxxxxxx.mysql.rds.aliyuncs.com",
				Port:    3306,
			},
			DBName: "db",
			Options: []string{
				"charset=utf8mb4",
				"parseTime=true",
				"loc=Local",
			},
		},
		ReadDSN: []*DSNConfig{
			{
				UserName: "root",
				Password: "123456",
				Endpoint: &EndpointConfig{
					Address: "rm-xxxxxxxxxxxxxxxxx.mysql.rds.aliyuncs.com",
					Port:    3306,
				},
				DBName: "db",
				Options: []string{
					"charset=utf8mb4",
					"parseTime=true",
					"loc=Local",
				},
			},
		},
		Active:       20,
		Idle:         10,
		ExecTimeout:  ctime.Duration(time.Millisecond * 300),
		QueryTimeout: ctime.Duration(time.Millisecond * 200),
		TranTimeout:  ctime.Duration(time.Millisecond * 400),
		IdleTimeout:  ctime.Duration(time.Hour * 4),
		Config: &render.Config{
			Stdout: true,
		},
	}

	client := NewMySQL(&c)

	var item interface{}
	err := client.ReadOnlyTable(context.Background(), "table").
		Where("id = 1").
		Find(&item).
		Error
	if err != nil {
		return
	}

	err = client.Table(context.Background(), "table").
		Where("id = 1").
		Update(&item).Scan(&item).
		Error
	if err != nil {
		return
	}

	err = client.Transaction(context.Background(), func(ctx context.Context, tx *OrmDB) error {
		err = tx.Table(ctx, "table").
			Where("id = 1").
			Update(&item).Scan(&item).
			Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
}

package mongo

import (
	"context"
	"fmt"
	"time"

	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ExampleNewMongo() {
	c := Config{
		DSN: &DSNConfig{
			UserName: "root",
			Password: "123456",
			Endpoints: []*EndpointConfig{
				{
					Address: "dds-xxxxxxxxxxxxxxxxx.mongodb.rds.aliyuncs.com",
					Port:    3717,
				},
				{
					Address: "dds-xxxxxxxxxxxxxxxxx.mongodb.rds.aliyuncs.com",
					Port:    3717,
				},
			},
			DBName:  "dbName",
			Options: []string{"replicaSet=mgset-xxxxxxxx"},
		},
		ReadDSN: []*DSNConfig{
			{
				UserName: "readonly",
				Password: "123456",
				Endpoints: []*EndpointConfig{
					{
						Address: "dds-xxxxxxxxxxxxxxxxx.mongodb.rds.aliyuncs.com",
						Port:    3717,
					},
					{
						Address: "dds-xxxxxxxxxxxxxxxxx.mongodb.rds.aliyuncs.com",
						Port:    3717,
					},
				},
				DBName:  "dbName",
				Options: []string{"replicaSet=mgset-xxxxxxxx"},
			},
		},
		ExecTimeout:  ctime.Duration(time.Second * 10),
		QueryTimeout: ctime.Duration(time.Second * 10),
		IdleTimeout:  ctime.Duration(time.Hour * 10),
		MaxPoolSize:  10,
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [dsn: %D] [duration: %d] %S  DB: %N , Collection: %n , Func: %F , Fields: %J{fCEO}",
		},
	}

	client := NewMongo(&c)

	var items []interface{}
	err := client.ReadOnlyCollection("collection").
		FindPage(context.Background(), bson.M{
			"name": bson.M{
				"$regex": primitive.Regex{
					Pattern: "name",
					Options: "im",
				},
			},
		}, 0, 10).
		Decode(&items)
	if err != nil {
		return
	}
	fmt.Printf("%v\n", items)

	var item interface{}
	_, err = client.Collection("collection").
		InsertOne(context.Background(), item)
	if err != nil {
		return
	}
}

package main

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/service"

	"context"
	"time"

	"github.com/pkg/errors"
	framework "gitlab.shanhai.int/sre/app-framework"
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	_mongo "gitlab.shanhai.int/sre/library/database/mongo"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	config.Read("./config/config.yaml")
	config.Conf.Log.Config = &render.Config{
		Stdout: true,
		OutDir: "./log",
	}
	config.Conf.Mongo.Config = &render.Config{
		Stdout: true,
		OutDir: "./log",
	}
	config.Conf.HTTPClient.Config = &render.Config{
		Stdout: true,
		OutDir: "./log",
	}

	framework.Run(
		config.Conf.Config,
		service.New(),
		GetInitProjectResourceSpecServer(),
	)
}

func GetInitProjectResourceSpecServer() framework.ServerInterface {
	svr := new(framework.JobServer)

	svr.SetJob("init project resource spec", func(ctx context.Context) error {
		ams := _mongo.NewMongo(&_mongo.Config{
			DSN: &_mongo.DSNConfig{
				UserName: "ams",
				Password: "QingTing0108",
				Endpoints: []*_mongo.EndpointConfig{
					{
						Address: "dds-bp19f2a965a784741.mongodb.rds.aliyuncs.com",
						Port:    3717,
					},
					{
						Address: "dds-bp19f2a965a784742.mongodb.rds.aliyuncs.com",
						Port:    3717,
					},
				},
				DBName:  "ams_prd",
				Options: []string{"replicaSet=mgset-25604411"},
			},
			ExecTimeout:  ctime.Duration(time.Second * 30),
			QueryTimeout: ctime.Duration(time.Second * 30),
			IdleTimeout:  ctime.Duration(time.Second * 30),
			MaxPoolSize:  20,
			Config: &render.Config{
				Stdout: true,
				OutDir: "./log",
			},
		})

		res := make([]*entity.Project, 0)
		err := ams.ReadOnlyCollection(new(entity.Project).TableName()).
			Find(ctx, bson.M{}).
			Decode(&res)
		if err != nil {
			panic(err)
		}

		for _, project := range res {
			_, err := ams.Collection(new(entity.Project).TableName()).
				UpdateOne(ctx, bson.M{
					"_id": project.ID,
				}, bson.M{
					"$set": bson.M{
						"resource_spec": map[entity.AppEnvName]entity.ProjectResourceSpec{
							entity.AppEnvFat: {
								CPURequestList: entity.CPUStgRequestResourceList,
								CPULimitList:   entity.CPUStgLimitResourceList,
								MemRequestList: entity.MemStgRequestResourceList,
								MemLimitList:   entity.MemStgLimitResourceList,
							},
							entity.AppEnvStg: {
								CPURequestList: entity.CPUStgRequestResourceList,
								CPULimitList:   entity.CPUStgLimitResourceList,
								MemRequestList: entity.MemStgRequestResourceList,
								MemLimitList:   entity.MemStgLimitResourceList,
							},
							entity.AppEnvPre: {
								CPURequestList: entity.CPUPrdRequestResourceList,
								CPULimitList:   entity.CPUPrdLimitResourceList,
								MemRequestList: entity.MemPrdRequestResourceList,
								MemLimitList:   entity.MemPrdLimitResourceList,
							},
							entity.AppEnvPrd: {
								CPURequestList: entity.CPUPrdRequestResourceList,
								CPULimitList:   entity.CPUPrdLimitResourceList,
								MemRequestList: entity.MemPrdRequestResourceList,
								MemLimitList:   entity.MemPrdLimitResourceList,
							},
						},
					},
				})
			if err != nil {
				return errors.Wrapf(errcode.MongoError, "%s", err)
			}
		}

		return nil
	})

	return svr
}

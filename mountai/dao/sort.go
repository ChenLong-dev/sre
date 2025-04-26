package dao

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 通用排序参数
var (
	MongoSortByIDAsc          = bson.M{"_id": 1}
	MongoSortByCreateTimeDesc = bson.M{"create_time": -1}

	MongoFindOptionWithSortByIDAsc = options.Find().SetSort(MongoSortByIDAsc)
)

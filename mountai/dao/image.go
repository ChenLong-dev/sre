package dao

import (
	"context"

	"rulai/models/entity"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d *Dao) CreateJenkinsBuildImage(ctx context.Context, build *entity.JenkinsBuildImage) error {
	_, err := d.Mongo.Collection(new(entity.JenkinsBuildImage).TableName()).
		InsertOne(ctx, build)
	if err != nil {
		return errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return nil
}

func (d *Dao) GetLastJenkinsBuildImage(ctx context.Context, filter bson.M) (*entity.JenkinsBuildImage, error) {
	build := new(entity.JenkinsBuildImage)
	err := d.Mongo.ReadOnlyCollection(build.TableName()).
		FindOne(ctx, filter, &options.FindOneOptions{
			Sort: bson.M{
				"_id": -1,
			},
		}).
		Decode(build)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrapf(errcode.NoRowsFoundError, "%s", err)
	} else if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return build, err
}

func (d *Dao) GetProjectLastJenkinsBuildImages(ctx context.Context, projectID string) ([]*entity.JenkinsBuildImage, error) {
	matchStage := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "project_id", Value: projectID}}}}

	sortStage := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "branch_name", Value: 1},
			{Key: "_id", Value: -1}}}}
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$branch_name"},
			{Key: "document", Value: bson.D{
				{Key: "$first", Value: "$$ROOT"}}}}}}

	projectStage := bson.D{
		{Key: "$replaceRoot", Value: bson.D{
			{Key: "newRoot", Value: "$document"}}}}

	images := make([]*entity.JenkinsBuildImage, 0)

	err := d.Mongo.ReadOnlyCollection(new(entity.JenkinsBuildImage).TableName()).
		Aggregate(ctx, mongo.Pipeline{
			matchStage,
			sortStage,
			groupStage,
			projectStage,
		}).
		Decode(&images)

	if err != nil {
		return nil, errors.Wrapf(errcode.MongoError, "%s", err)
	}

	return images, nil
}

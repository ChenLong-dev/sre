package mongo

import (
	"context"
	"time"

	"gitlab.shanhai.int/sre/library/base/ctime"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 集合
type Collection struct {
	// mongo集合
	*mongo.Collection
	// 配置文件
	conf Config
	// 所属连接
	con *Connection
}

// 设置执行超时时间
func (c *Collection) ExecTimeout(duration time.Duration) *Collection {
	c.conf.ExecTimeout = ctime.Duration(duration)
	return c
}

// 设置查询超时时间
func (c *Collection) QueryTimeout(duration time.Duration) *Collection {
	c.conf.QueryTimeout = ctime.Duration(duration)
	return c
}

// 批量写操作
func (c *Collection) BulkWrite(ctx context.Context, models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	return c.Collection.BulkWrite(c.getExecContext(ctx), models, opts...)
}

// 插入单个文档
func (c *Collection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (result *mongo.InsertOneResult, err error) {
	ctx, hk := c.con.before(ctx, "InsertOne", c.Name(), nil, document, nil, opts)
	result, err = c.Collection.InsertOne(c.getExecContext(ctx), document, opts...)
	c.con.after(hk, err)

	return
}

// 插入多个文档
func (c *Collection) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (result *mongo.InsertManyResult, err error) {
	ctx, hk := c.con.before(ctx, "InsertMany", c.Name(), nil, documents, nil, opts)
	result, err = c.Collection.InsertMany(c.getExecContext(ctx), documents, opts...)
	c.con.after(hk, err)

	return
}

// 删除单个文档
func (c *Collection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (result *mongo.DeleteResult, err error) {
	ctx, hk := c.con.before(ctx, "DeleteOne", c.Name(), filter, nil, nil, opts)
	result, err = c.Collection.DeleteOne(c.getExecContext(ctx), filter, opts...)
	c.con.after(hk, err)

	return
}

// 删除多个文档
func (c *Collection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (result *mongo.DeleteResult, err error) {
	ctx, hk := c.con.before(ctx, "DeleteMany", c.Name(), filter, nil, nil, opts)
	result, err = c.Collection.DeleteMany(c.getExecContext(ctx), filter, opts...)
	c.con.after(hk, err)

	return
}

// 更新单个文档
func (c *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (result *mongo.UpdateResult, err error) {
	ctx, hk := c.con.before(ctx, "UpdateOne", c.Name(), filter, update, nil, opts)
	result, err = c.Collection.UpdateOne(c.getExecContext(ctx), filter, update, opts...)
	c.con.after(hk, err)

	return
}

// 更新多个文档
func (c *Collection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (result *mongo.UpdateResult, err error) {
	ctx, hk := c.con.before(ctx, "UpdateMany", c.Name(), filter, update, nil, opts)
	result, err = c.Collection.UpdateMany(c.getExecContext(ctx), filter, update, opts...)
	c.con.after(hk, err)

	return
}

// 替换单个文档
func (c *Collection) ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.ReplaceOptions) (result *mongo.UpdateResult, err error) {
	ctx, hk := c.con.before(ctx, "ReplaceOne", c.Name(), filter, replacement, nil, opts)
	result, err = c.Collection.ReplaceOne(c.getExecContext(ctx), filter, replacement, opts...)
	c.con.after(hk, err)

	return
}

// 聚合管道
func (c *Collection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (result *FindResult) {
	execCtx := c.getExecContext(ctx)
	raws := make([]bson.Raw, 0)

	ctx, hk := c.con.before(ctx, "Aggregate", c.Name(), nil, nil, pipeline, opts)
	cur, err := c.Collection.Aggregate(execCtx, pipeline, opts...)
	c.con.after(hk, err)

	if err != nil {
		return &FindResult{
			raws: nil,
			err:  err,
		}
	}

	defer cur.Close(execCtx)
	for cur.Next(execCtx) {
		raws = append(raws, cur.Current)
	}

	return &FindResult{
		raws: raws,
		err:  nil,
	}
}

// 统计文档数
func (c *Collection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (result int64, err error) {
	ctx, hk := c.con.before(ctx, "CountDocuments", c.Name(), filter, nil, nil, opts)
	result, err = c.Collection.CountDocuments(c.getQueryContext(ctx), filter, opts...)
	c.con.after(hk, err)

	return
}

// 预估文档数，从info中直接获取
func (c *Collection) EstimatedDocumentCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (result int64, err error) {
	ctx, hk := c.con.before(ctx, "EstimatedDocumentCount", c.Name(), nil, nil, nil, opts)
	result, err = c.Collection.EstimatedDocumentCount(c.getQueryContext(ctx), opts...)
	c.con.after(hk, err)

	return
}

// 去重查找文档
func (c *Collection) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) (result []interface{}, err error) {
	ctx, hk := c.con.before(ctx, "Distinct", c.Name(), filter, nil, fieldName, opts)
	result, err = c.Collection.Distinct(c.getQueryContext(ctx), fieldName, filter, opts...)
	c.con.after(hk, err)

	return
}

// 查找文档
func (c *Collection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (result *FindResult) {
	queryCtx := c.getQueryContext(ctx)
	raws := make([]bson.Raw, 0)

	ctx, hk := c.con.before(ctx, "Find", c.Name(), filter, nil, nil, opts)
	cur, err := c.Collection.Find(queryCtx, filter, opts...)
	c.con.after(hk, err)

	if err != nil {
		return &FindResult{
			raws: nil,
			err:  err,
		}
	}

	defer cur.Close(queryCtx)
	for cur.Next(queryCtx) {
		raws = append(raws, cur.Current)
	}

	return &FindResult{
		raws: raws,
		err:  nil,
	}
}

// 分页查找
func (c *Collection) FindPage(ctx context.Context, filter interface{}, page, limit int, opts ...*options.FindOptions) (result *FindResult) {
	Limit := int64(limit)
	Skip := int64(page * limit)
	opts = append(opts, &options.FindOptions{
		Limit: &Limit,
		Skip:  &Skip,
	})

	queryCtx := c.getQueryContext(ctx)
	raws := make([]bson.Raw, 0)

	ctx, hk := c.con.before(ctx, "FindPage", c.Name(), filter, nil, nil, opts)
	cur, err := c.Collection.Find(ctx, filter, opts...)
	c.con.after(hk, err)

	if err != nil {
		return &FindResult{
			err: err,
		}
	}

	defer cur.Close(queryCtx)
	for cur.Next(queryCtx) {
		raws = append(raws, cur.Current)
	}

	return &FindResult{
		raws: raws,
		err:  nil,
	}
}

// 查找单个文档
func (c *Collection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *SingleResult {
	ctx, hk := c.con.before(ctx, "FindOne", c.Name(), filter, nil, nil, opts)
	result := c.Collection.FindOne(c.getQueryContext(ctx), filter, opts...)
	c.con.after(hk, result.Err())

	return &SingleResult{
		SingleResult: result,
	}
}

// 查找单个文档并删除
func (c *Collection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *SingleResult {
	ctx, hk := c.con.before(ctx, "FindOneAndDelete", c.Name(), filter, nil, nil, opts)
	result := c.Collection.FindOneAndDelete(c.getExecContext(ctx), filter, opts...)
	c.con.after(hk, result.Err())

	return &SingleResult{
		SingleResult: result,
	}
}

// 查找单个文档并替换
func (c *Collection) FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *SingleResult {
	ctx, hk := c.con.before(ctx, "FindOneAndReplace", c.Name(), filter, replacement, nil, opts)
	result := c.Collection.FindOneAndReplace(c.getExecContext(ctx), filter, replacement, opts...)
	c.con.after(hk, result.Err())

	return &SingleResult{
		SingleResult: result,
	}
}

// 查找单个文档并更新
func (c *Collection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *SingleResult {
	ctx, hk := c.con.before(ctx, "FindOneAndUpdate", c.Name(), filter, update, nil, opts)
	result := c.Collection.FindOneAndUpdate(c.getExecContext(ctx), filter, update, opts...)
	c.con.after(hk, result.Err())

	return &SingleResult{
		SingleResult: result,
	}
}

// 观察数据库变动流
func (c *Collection) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (result *mongo.ChangeStream, err error) {
	ctx, hk := c.con.before(ctx, "Watch", c.Name(), nil, nil, pipeline, opts)
	result, err = c.Collection.Watch(c.getQueryContext(ctx), pipeline, opts...)
	c.con.after(hk, err)

	return
}

// 获取索引
func (c *Collection) Indexes(ctx context.Context) (result mongo.IndexView, err error) {
	ctx, hk := c.con.before(ctx, "Indexes", c.Name(), nil, nil, nil, nil)
	result = c.Collection.Indexes()
	c.con.after(hk, nil)

	return
}

// 丢弃
func (c *Collection) Drop(ctx context.Context) error {
	ctx, hk := c.con.before(ctx, "Drop", c.Name(), nil, nil, nil, nil)
	err := c.Collection.Drop(c.getExecContext(ctx))
	c.con.after(hk, err)

	return err
}

// 获取写context
func (c *Collection) getExecContext(ctx context.Context) context.Context {
	_, execCtx, _ := c.conf.ExecTimeout.Shrink(ctx)
	return execCtx
}

// 获取读context
func (c *Collection) getQueryContext(ctx context.Context) context.Context {
	_, queryCtx, _ := c.conf.QueryTimeout.Shrink(ctx)
	return queryCtx
}

package mongo

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/base/runtime"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// 连接
type Connection struct {
	// mongo客户端
	*mongo.Client
	// 配置文件
	conf *Config
	// 钩子管理器
	manager *hook.Manager
	// 数据库名称
	dbName string
}

// 开启会话
func (con *Connection) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return con.Client.StartSession(opts...)
}

// 设置数据库
// 如切换不同数据库，需要使用admin鉴权
func (con *Connection) Database(name string) *Connection {
	return &Connection{
		Client:  con.Client,
		conf:    con.conf,
		manager: con.manager,
		dbName:  name,
	}
}

// 设置集合
func (con *Connection) Collection(collectionName string, opts ...*options.CollectionOptions) *Collection {
	collection := con.Client.Database(con.dbName).Collection(collectionName, opts...)
	return &Collection{
		Collection: collection,
		conf:       *con.conf,
		con:        con,
	}
}

// 开启事务
func (con *Connection) Transaction(ctx context.Context, callback func(con *Connection, ctx mongo.SessionContext) (interface{}, error),
	opts ...*options.SessionOptions) (interface{}, error) {
	session, err := con.StartSession(opts...)
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	ctx, hk := con.before(ctx, "Transaction", "", nil, nil, nil, opts)
	res, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return callback(con, sessCtx)
	})
	con.after(hk, err)

	return res, err
}

// Ping检查
func (con *Connection) ping(c context.Context, rp *readpref.ReadPref) (err error) {
	_, hk := con.before(c, "Ping", "", rp.String(), nil, nil, nil)
	err = con.Ping(c, rp)
	con.after(hk, err)
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

// 操作前注入
func (con *Connection) before(ctx context.Context, funcName, collectionName string,
	filterField, changeField, extraField, optionField interface{}) (context.Context, *hook.Hook) {
	hk := con.manager.CreateHook(ctx).
		AddArg(render.StartTimeArgKey, time.Now()).
		AddArg(render.SourceArgKey, runtime.GetDefaultFilterCallers()).
		AddArg("func_name", funcName).
		AddArg("collection_name", collectionName).
		AddArg("db_name", con.dbName).
		AddArg("filter_field", filterField).
		AddArg("change_field", changeField).
		AddArg("extra_field", extraField).
		AddArg("option_field", optionField).
		ProcessPreHook()

	return hk.Context(), hk
}

// 操作后注入
func (con *Connection) after(hk *hook.Hook, err error) {
	endTime := time.Now()
	duration := endTime.Sub(hk.Arg(render.StartTimeArgKey).(time.Time))

	hk = hk.AddArg(render.EndTimeArgKey, endTime).
		AddArg(render.DurationArgKey, duration).
		AddArg(render.ErrorArgKey, err).
		ProcessAfterHook()
}

package mongo

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

// DB
type DB struct {
	// 主连接
	write *Connection
	// 读连接数组
	read []*Connection
	// 读索引号，用于分配读连接
	idx int64
	// 主数据库
	master *DB
	// 配置文件
	conf *Config
}

// 打开数据库
func Open(c *Config) (*DB, error) {
	db := new(DB)
	w, err := connect(c, c.DSN)
	if err != nil {
		return nil, err
	}

	if len(c.ReadDSN) == 0 {
		c.ReadDSN = []*DSNConfig{c.DSN}
	}
	rs := make([]*Connection, 0, len(c.ReadDSN))
	for _, rd := range c.ReadDSN {
		r, err := connect(c, rd)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	db.conf = c
	db.write = w
	db.read = rs
	db.master = &DB{write: db.write}

	return db, nil
}

// 获取配置文件
func (db *DB) GetConfig() *Config {
	return db.conf
}

// 获取写连接
func (db *DB) Connection() *Connection {
	return db.write
}

// 获取读连接
func (db *DB) ReadOnlyConnection() *Connection {
	return db.read[db.readIndex()]
}

// 设置写连接的collection
func (db *DB) Collection(collectionName string) *Collection {
	return db.Connection().Collection(collectionName)
}

// 设置读连接的collection
func (db *DB) ReadOnlyCollection(collectionName string) *Collection {
	return db.ReadOnlyConnection().Collection(collectionName)
}

// 关闭连接
func (db *DB) Close(c context.Context) (err error) {
	db.write.manager.Close()
	if e := db.write.Disconnect(c); e != nil {
		err = errors.WithStack(e)
	}
	for _, rd := range db.read {
		rd.manager.Close()
		if e := rd.Disconnect(c); e != nil {
			err = errors.WithStack(e)
		}
	}
	return
}

// Ping操作
func (db *DB) Ping(c context.Context) (err error) {
	if err = db.Connection().ping(c, readpref.Primary()); err != nil {
		return
	}
	for _, rd := range db.read {
		if err = rd.ping(c, readpref.Secondary()); err != nil {
			return
		}
	}
	return
}

func concatConnectURI(dsnConfig *DSNConfig) string {
	endpoints := make([]string, 0)
	for _, endpoint := range dsnConfig.Endpoints {
		endpoints = append(endpoints, fmt.Sprintf("%s:%d", endpoint.Address, endpoint.Port))
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s/%s",
		dsnConfig.UserName, dsnConfig.Password, strings.Join(endpoints, ","),
		dsnConfig.DBName)
	if len(dsnConfig.Options) != 0 {
		uri = fmt.Sprintf("%s?%s", uri, strings.Join(dsnConfig.Options, "&"))
	}

	return uri
}

// 进行数据库连接
func connect(c *Config, dsnConfig *DSNConfig) (*Connection, error) {
	opt := options.Client()
	opt.SetLocalThreshold(time.Duration(c.ExecTimeout))
	opt.SetMaxConnIdleTime(time.Duration(c.IdleTimeout))
	opt.SetMaxPoolSize(uint64(c.MaxPoolSize))
	opt.SetMinPoolSize(uint64(c.MinPoolSize))
	readPreference, err := readpref.New(readpref.SecondaryMode)
	if err != nil {
		return nil, err
	}
	opt.SetReadPreference(readPreference)
	opt.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	opt.ApplyURI(concatConnectURI(dsnConfig))

	client, err := mongo.NewClient(opt)
	if err != nil {
		err = errors.WithStack(err)
		return nil, err
	}

	err = client.Connect(context.Background())
	if err != nil {
		err = errors.WithStack(err)
		return nil, err
	}
	return &Connection{
		Client:  client,
		conf:    c,
		dbName:  dsnConfig.DBName,
		manager: NewHookManager(c.Config, dsnConfig.UserName),
	}, nil
}

// 获取读连接索引
func (db *DB) readIndex() int {
	if len(db.read) == 0 {
		return 0
	}
	v := atomic.AddInt64(&db.idx, 1) % int64(len(db.read))
	atomic.StoreInt64(&db.idx, v)
	return int(v)
}

// 获取mongo标签
func getMongoTags(field reflect.StructField) map[string]interface{} {
	tagString := field.Tag.Get("bson")

	tags := make(map[string]interface{})

	for idx, tag := range strings.Split(tagString, ",") {
		if idx == 0 && tag != "" {
			tags["column"] = tag
		} else {
			tags[tag] = true
		}
	}

	return tags
}

// 结构体转mongo可读取的map结构
// 如果需要转换空值，则转换的字段必须有convertible的标签
// todo:由于无法判断空值，所以只要包含convertible标签，都会转换
// Deprecated
// 	方法已废弃。如需更新零值，实体及request模型请使用 null包 中的对应类型
func StructToMongoMap(obj interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		var (
			vField = v.Field(i)
			tField = v.Type().Field(i)
			tagMap = getMongoTags(tField)
		)

		if tField.Type.Kind() == reflect.Struct && tField.Anonymous {
			childMap := StructToMongoMap(vField.Interface())
			for childKey, childValue := range childMap {
				data[childKey] = childValue
			}
			continue
		} else if _, ok := tagMap["convertible"]; vField.IsZero() && !ok {
			continue
		} else if columnName, ok := tagMap["column"]; ok {
			data[columnName.(string)] = vField.Interface()
		} else {
			data[tField.Name] = vField.Interface()
		}
	}
	return data
}

package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
)

// DB
type OrmDB struct {
	// 主连接
	*gorm.DB
	// 只读连接
	read []*gorm.DB
	// 只读连接索引号
	idx int64
	// 配置文件
	conf *Config
}

// 事务函数
type OrmTransactionFunction func(ctx context.Context, tx *OrmDB) error

// 打开db
func OpenOrm(c *Config) (*OrmDB, error) {
	ormDB := new(OrmDB)
	ormDB.conf = c

	d, err := connectGORM(c, c.DSN)
	if err != nil {
		return nil, err
	}
	ormDB.DB = d

	if len(c.ReadDSN) == 0 {
		c.ReadDSN = []*DSNConfig{c.DSN}
	}
	rs := make([]*gorm.DB, 0, len(c.ReadDSN))
	for _, rd := range c.ReadDSN {
		d, err := connectGORM(c, rd)
		if err != nil {
			return nil, err
		}
		rs = append(rs, d)
	}
	ormDB.read = rs

	return ormDB, nil
}

// 获取配置文件
func (db *OrmDB) GetConfig() *Config {
	return db.conf
}

// clone DB
func (db *OrmDB) Clone(write *gorm.DB, read []*gorm.DB) *OrmDB {
	return &OrmDB{
		DB:   write,
		read: read,
		idx:  0,
		conf: db.conf,
	}
}

// Ping操作
func (db *OrmDB) Ping(c context.Context) (err error) {
	if err = db.ping(c, db.DB); err != nil {
		return
	}
	for _, rd := range db.read {
		if err = db.ping(c, rd); err != nil {
			return
		}
	}
	return
}

// 关闭数据库连接
func (db *OrmDB) Close() (err error) {
	if e := db.DB.Close(); e != nil {
		err = errors.WithStack(e)
	}
	for _, rd := range db.read {
		if e := rd.Close(); e != nil {
			err = errors.WithStack(e)
		}
	}
	return
}

// 设定context
func (db *OrmDB) Context(ctx context.Context) *gorm.DB {
	return db.Set(ContextStoreKey, ctx)
}

// 设定数据源
func (db *OrmDB) DataSource(ctx context.Context, isReadOnly bool) *gorm.DB {
	return db.getCurrentDB(isReadOnly).Set(ContextStoreKey, ctx)
}

// 设定模型
func (db *OrmDB) Model(ctx context.Context, value interface{}) *gorm.DB {
	return db.DataSource(ctx, false).Model(value)
}

// 设定表名
func (db *OrmDB) Table(ctx context.Context, name string) *gorm.DB {
	return db.DataSource(ctx, false).Table(name)
}

// 设定只读表名
func (db *OrmDB) ReadOnlyTable(ctx context.Context, name string) *gorm.DB {
	return db.DataSource(ctx, true).Table(name)
}

// 设定只读模型
func (db *OrmDB) ReadOnlyModel(ctx context.Context, value interface{}) *gorm.DB {
	return db.DataSource(ctx, true).Model(value)
}

// 获取只读连接
func (db *OrmDB) ReadOnly() *gorm.DB {
	idx := db.readIndex()
	for i := range db.read {
		if rd := db.read[(idx+i)%len(db.read)]; rd != nil {
			return rd
		}
	}
	return db.DB
}

// 使用原生sql查询
func (db *OrmDB) Raw(ctx context.Context, sql string, values ...interface{}) *gorm.DB {
	return db.Context(ctx).Raw(sql, values...)
}

// 执行原生sql
func (db *OrmDB) Exec(ctx context.Context, sql string, values ...interface{}) *gorm.DB {
	return db.Context(ctx).Exec(sql, values...)
}

// 开启事务
func (db *OrmDB) Transaction(ctx context.Context, transactionFunc OrmTransactionFunction) (err error) {
	transactionCtx, _ := context.WithTimeout(ctx, time.Duration(db.conf.TranTimeout))

	tx := db.BeginTx(transactionCtx, &sql.TxOptions{})
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
	}()

	txDB := db.Clone(tx, []*gorm.DB{tx})

	err = transactionFunc(transactionCtx, txDB)

	return err
}

func concatConnectURI(dsnConfig *DSNConfig) string {
	uri := fmt.Sprintf("%s:%s@(%s:%d)/%s",
		dsnConfig.UserName, dsnConfig.Password,
		dsnConfig.Endpoint.Address, dsnConfig.Endpoint.Port,
		dsnConfig.DBName)
	if len(dsnConfig.Options) != 0 {
		uri = fmt.Sprintf("%s?%s", uri, strings.Join(dsnConfig.Options, "&"))
	}

	return uri
}

// 空logger
type nopLogger struct{}

func (nopLogger) Print(values ...interface{}) {}

// 建立连接
func connectGORM(c *Config, dsnConfig *DSNConfig) (*gorm.DB, error) {
	d, err := gorm.Open("mysql", concatConnectURI(dsnConfig))
	if err != nil {
		err = errors.WithStack(err)
		return nil, err
	}
	d.LogMode(false)
	d.SetLogger(nopLogger{})
	d.DB().SetMaxOpenConns(c.Active)
	d.DB().SetMaxIdleConns(c.Idle)
	// todo:超时时间暂时没有配置
	d.DB().SetConnMaxLifetime(time.Duration(c.IdleTimeout))

	RegisterCustomCallbacks(d, NewHookManager(c.Config, dsnConfig))

	return d, nil
}

// 获取只读索引
func (db *OrmDB) readIndex() int {
	if len(db.read) == 0 {
		return 0
	}
	v := atomic.AddInt64(&db.idx, 1) % int64(len(db.read))
	atomic.StoreInt64(&db.idx, v)
	return int(v)
}

// Ping指定连接
func (db *OrmDB) ping(c context.Context, now *gorm.DB) (err error) {
	_, c, cancel := db.conf.ExecTimeout.Shrink(c)
	err = now.DB().PingContext(c)
	cancel()
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

// 获取当前连接
func (db *OrmDB) getCurrentDB(isReadOnly bool) *gorm.DB {
	if isReadOnly {
		return db.ReadOnly()
	} else {
		return db.DB
	}
}

// 获取gorm的标签
func getGORMTags(field reflect.StructField) map[string]interface{} {
	tagString := field.Tag.Get("gorm")

	tags := make(map[string]interface{})

	for _, tag := range strings.Split(tagString, ";") {
		tagSlice := strings.SplitN(tag, ":", 2)

		switch len(tagSlice) {
		case 1:
			tags[tagSlice[0]] = ""
		case 2:
			tags[strings.TrimSpace(tagSlice[0])] = strings.TrimSpace(tagSlice[1])
		}
	}

	return tags
}

// 结构体转gorm可读取的map结构
// 如果需要转换空值，则转换的字段必须有convertible的标签
// todo:由于无法判断空值，所以只要包含convertible标签，都会转换
// Deprecated
// 	方法已废弃。如需更新零值，实体及request模型请使用 null 包中的对应类型
func StructToGORMMap(obj interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		var (
			vField = v.Field(i)
			tField = v.Type().Field(i)
			tagMap = getGORMTags(tField)
		)

		if tField.Type.Kind() == reflect.Struct && tField.Anonymous {
			childMap := StructToGORMMap(vField.Interface())
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

// sql是否可以打印
func ormIsPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

// 生成完整sql
func generateFullSQL(originSql string, params []interface{}) string {
	formattedValues := make([]string, 0)
	var fullSql string

	for _, value := range params {
		indirectValue := reflect.Indirect(reflect.ValueOf(value))
		if indirectValue.IsValid() {
			value = indirectValue.Interface()
			if t, ok := value.(time.Time); ok {
				formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
			} else if b, ok := value.([]byte); ok {
				if str := string(b); ormIsPrintable(str) {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
				} else {
					formattedValues = append(formattedValues, "'<binary>'")
				}
			} else if r, ok := value.(driver.Valuer); ok {
				if value, err := r.Value(); err == nil && value != nil {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				} else {
					formattedValues = append(formattedValues, "NULL")
				}
			} else {
				switch value.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
					formattedValues = append(formattedValues, fmt.Sprintf("%v", value))
				default:
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				}
			}
		} else {
			formattedValues = append(formattedValues, "NULL")
		}
	}

	// differentiate between $n placeholders or else treat like ?
	if regexp.MustCompile(`\$\d+`).MatchString(originSql) {
		fullSql = originSql
		for index, value := range formattedValues {
			placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
			fullSql = regexp.MustCompile(placeholder).ReplaceAllString(fullSql, value+"$1")
		}
	} else {
		formattedValuesLength := len(formattedValues)
		for index, value := range regexp.MustCompile(`\?`).Split(originSql, -1) {
			fullSql += value
			if index < formattedValuesLength {
				fullSql += formattedValues[index]
			}
		}
	}

	return fullSql
}

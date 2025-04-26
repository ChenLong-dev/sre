package sql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/base/runtime"
)

const (
	// 信息存储键
	HookStoreKey    = "hook"
	ContextStoreKey = "scope_context"
)

// 自定义回调
type qtCallBack struct {
	// 钩子管理器
	manager *hook.Manager
}

// 新建自定义回调
func newQTCallbacks(manager *hook.Manager) *qtCallBack {
	return &qtCallBack{
		manager: manager,
	}
}

// 创建前回调
func (c *qtCallBack) beforeCreate(scope *gorm.Scope) { c.before(scope, "INSERT") }

// 创建后回调
func (c *qtCallBack) afterCreate(scope *gorm.Scope) { c.after(scope, "INSERT") }

// 查询前回调
func (c *qtCallBack) beforeQuery(scope *gorm.Scope) { c.before(scope, "SELECT") }

// 查询后回调
func (c *qtCallBack) afterQuery(scope *gorm.Scope) { c.after(scope, "SELECT") }

// 更新前回调
func (c *qtCallBack) beforeUpdate(scope *gorm.Scope) { c.before(scope, "UPDATE") }

// 更新后回调
func (c *qtCallBack) afterUpdate(scope *gorm.Scope) { c.after(scope, "UPDATE") }

// 删除前回调
func (c *qtCallBack) beforeDelete(scope *gorm.Scope) { c.before(scope, "DELETE") }

// 删除后回调
func (c *qtCallBack) afterDelete(scope *gorm.Scope) { c.after(scope, "DELETE") }

// 行数查询前回调
func (c *qtCallBack) beforeRowQuery(scope *gorm.Scope) {
	operation := strings.ToUpper(strings.Split(scope.SQL, " ")[0])
	c.before(scope, operation)
}

// 行数查询后回调
func (c *qtCallBack) afterRowQuery(scope *gorm.Scope) {
	operation := strings.ToUpper(strings.Split(scope.SQL, " ")[0])
	c.after(scope, operation)
}

// 操作前回调
func (c *qtCallBack) before(scope *gorm.Scope, operation string) {
	ctxValue, ok := scope.DB().Get(ContextStoreKey)
	if !ok {
		return
	}
	ctx := ctxValue.(context.Context)

	hk := c.manager.CreateHook(ctx).
		AddArg(render.StartTimeArgKey, time.Now()).
		AddArg(render.SourceArgKey, runtime.GetDefaultFilterCallers()).
		AddArg("level", "sql").
		AddArg("operation", operation).
		ProcessPreHook()

	scope.Set(HookStoreKey, hk)
	scope.Set(ContextStoreKey, hk.Context())
}

// 操作后回调
func (c *qtCallBack) after(scope *gorm.Scope, operation string) {
	hkValue, ok := scope.DB().Get(HookStoreKey)
	if !ok {
		return
	}
	hk := hkValue.(*hook.Hook)

	endTime := time.Now()
	duration := endTime.Sub(hk.Arg(render.StartTimeArgKey).(time.Time))
	hk.AddArg(render.EndTimeArgKey, endTime).
		AddArg(render.DurationArgKey, duration).
		AddArg("table", scope.TableName()).
		AddArg("rows", int(scope.DB().RowsAffected)).
		AddArg("sql", generateFullSQL(scope.SQL, scope.SQLVars)).
		AddArg(render.ErrorArgKey, scope.DB().Error).
		ProcessAfterHook()
}

// 注册自定义回调
func RegisterCustomCallbacks(db *gorm.DB, manager *hook.Manager) {
	callbacks := newQTCallbacks(manager)

	registerQTCallbacks(db, "create", callbacks)
	registerQTCallbacks(db, "query", callbacks)
	registerQTCallbacks(db, "update", callbacks)
	registerQTCallbacks(db, "delete", callbacks)
	registerQTCallbacks(db, "row_query", callbacks)
}

// 注册自定义回调
func registerQTCallbacks(db *gorm.DB, name string, c *qtCallBack) {
	beforeName := fmt.Sprintf("qt:%v_before", name)
	afterName := fmt.Sprintf("qt:%v_after", name)

	gormCallbackName := fmt.Sprintf("gorm:%v", name)

	switch name {
	case "create":
		db.Callback().Create().Before(gormCallbackName).Register(beforeName, c.beforeCreate)
		db.Callback().Create().After(gormCallbackName).Register(afterName, c.afterCreate)
	case "query":
		db.Callback().Query().Before(gormCallbackName).Register(beforeName, c.beforeQuery)
		db.Callback().Query().After(gormCallbackName).Register(afterName, c.afterQuery)
	case "update":
		db.Callback().Update().Before(gormCallbackName).Register(beforeName, c.beforeUpdate)
		db.Callback().Update().After(gormCallbackName).Register(afterName, c.afterUpdate)
	case "delete":
		db.Callback().Delete().Before(gormCallbackName).Register(beforeName, c.beforeDelete)
		db.Callback().Delete().After(gormCallbackName).Register(afterName, c.afterDelete)
	case "row_query":
		db.Callback().RowQuery().Before(gormCallbackName).Register(beforeName, c.beforeRowQuery)
		db.Callback().RowQuery().After(gormCallbackName).Register(afterName, c.afterRowQuery)
	}
}

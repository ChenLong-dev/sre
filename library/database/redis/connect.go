package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/base/runtime"
)

// 连接
type Conn struct {
	// 连接
	redis.Conn
	// 钩子管理器
	manager *hook.Manager
	// 连接池
	pool *Pool
}

func (c *Conn) GetOriginConnect() redis.Conn {
	return c.Conn
}

// 执行命令
func (c *Conn) Do(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	ctx, hk := c.before(ctx, "Do", commandName, args)

	reply, err = c.Conn.Do(commandName, args...)

	c.after(hk, err, reply)

	return
}

// 刷新输出缓存
func (c *Conn) Flush(ctx context.Context) error {
	ctx, hk := c.before(ctx, "Flush", "pipeline::flush", nil)

	err := c.Conn.Flush()

	c.after(hk, err, nil)

	return err
}

// 发送写命令
func (c *Conn) Send(ctx context.Context, commandName string, args ...interface{}) error {
	ctx, hk := c.before(ctx, "Send", fmt.Sprintf("pipeline::send::%s", commandName), args)

	err := c.Conn.Send(commandName, args...)

	c.after(hk, err, nil)

	return err
}

// 接受单个回复
func (c *Conn) Receive(ctx context.Context) (reply interface{}, err error) {
	ctx, hk := c.before(ctx, "Receive", "", nil)

	reply, err = c.Conn.Receive()

	c.after(hk, err, reply)

	return
}

// Ping操作
func (c *Conn) ping(ctx context.Context) (reply interface{}, err error) {
	reply, err = c.Conn.Do("PING")
	if err != nil {
		return nil, err
	}

	return reply, nil
}

// 操作前注入
func (c *Conn) before(ctx context.Context, funcName, commandName string,
	commandArgs []interface{}) (context.Context, *hook.Hook) {

	hk := c.manager.CreateHook(ctx).
		AddArg("endpoint", c.pool.config.GetEndpoint()).
		AddArg(render.StartTimeArgKey, time.Now()).
		AddArg(render.SourceArgKey, runtime.GetDefaultFilterCallers()).
		AddArg("func_name", funcName).
		AddArg("command_name", commandName).
		AddArg("command_args", commandArgs).
		ProcessPreHook()

	return hk.Context(), hk
}

// 操作后注入
func (c *Conn) after(hk *hook.Hook, err error, reply interface{}) {
	endTime := time.Now()
	duration := endTime.Sub(hk.Arg(render.StartTimeArgKey).(time.Time))

	hk = hk.AddArg(render.EndTimeArgKey, endTime).
		AddArg(render.DurationArgKey, duration).
		AddArg("replay", reply).
		AddArg(render.ErrorArgKey, err).
		ProcessAfterHook()
}

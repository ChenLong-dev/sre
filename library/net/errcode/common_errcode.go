package errcode

import "net/http"

var (
	// 0-1999999 为保留错误码
	OK = add(http.StatusOK, 0, "success")

	BadRequest   = add(http.StatusBadRequest, 1040400, "请求错误")
	Unauthorized = add(http.StatusUnauthorized, 1040401, "验证未通过")
	Forbidden    = add(http.StatusForbidden, 1040403, "服务器拒绝")
	NotFound     = add(http.StatusNotFound, 1040404, "没有找到路由")

	InternalError      = add(http.StatusInternalServerError, 1050500, "系统错误,请稍后重试")
	ServiceUnavailable = add(http.StatusServiceUnavailable, 1050503, "服务暂不可用")

	UnknownError         = add(http.StatusInternalServerError, 1060000, "未知错误")
	MysqlError           = add(http.StatusInternalServerError, 1060001, "Mysql数据库错误")
	MongoError           = add(http.StatusInternalServerError, 1060002, "Mongodb数据库错误")
	InvalidParams        = add(http.StatusBadRequest, 1060003, "参数错误")
	EncryptEncodeError   = add(http.StatusInternalServerError, 1060004, "加密错误")
	EncryptDecodeError   = add(http.StatusInternalServerError, 1060005, "解密错误")
	NoRowsFoundError     = add(http.StatusInternalServerError, 1060006, "没有找到任何记录")
	RedisError           = add(http.StatusInternalServerError, 1060007, "Redis数据库错误")
	RedisEmptyKeyError   = add(http.StatusInternalServerError, 1060008, "Redis键为空")
	RabbitMQError        = add(http.StatusInternalServerError, 1060009, "RabbitMQ错误")
	KafkaError           = add(http.StatusInternalServerError, 1060010, "Kafka错误")
	RedLockError         = add(http.StatusInternalServerError, 1060011, "RedLock错误")
	RedLockLockError     = add(http.StatusInternalServerError, 1060012, "RedLock加锁错误")
	RedLockUnLockError   = add(http.StatusInternalServerError, 1060013, "RedLock解锁错误")
	EtcdError            = add(http.StatusInternalServerError, 1060014, "Etcd数据库错误")
	BreakerOpenError     = add(http.StatusInternalServerError, 1060015, "断路器已开启")
	BreakerTimeoutError  = add(http.StatusInternalServerError, 1060016, "断路器超时错误")
	BreakerDegradedError = add(http.StatusInternalServerError, 1060017, "断路器已降级")
	SentryError          = add(http.StatusInternalServerError, 1060018, "Sentry错误")
)

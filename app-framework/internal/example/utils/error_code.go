package utils

import "gitlab.shanhai.int/sre/library/net/errcode"

var (
	// ==========================
	// 声明自定义错误代码时
	// 应当先查看Library中errcode包:https://gitlab.shanhai.int/sre/library/tree/master/net/errcode
	// 仔细阅读文档后并申请相应码段，在固定码段内新建错误
	// ==========================
	// 加锁错误
	OwnerGithubAggregationLockError = errcode.New(9999990, "分布式锁加锁失败")
	// 解锁错误
	OwnerGithubAggregationUnlockError = errcode.New(9999991, "分布式锁解锁失败")
)

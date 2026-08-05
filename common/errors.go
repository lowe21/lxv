package common

import (
	"github.com/lowe21/lxv/util"
)

var (
	GatewayError   = util.Error("GATEWAY_ERROR", "网关错误")
	SystemError    = util.Error("SYSTEM_ERROR", "系统异常")
	SystemBusy     = util.Error("SYSTEM_BUSY", "系统繁忙")
	InternalError  = util.Error("INTERNAL_ERROR", "内部错误")
	InvalidRequest = util.Error("INVALID_REQUEST", "无效请求")
	InvalidParam   = util.Error("INVALID_PARAM", "参数错误")
	InvalidSign    = util.Error("INVALID_SIGN", "验签失败")
	InvalidAuth    = util.Error("INVALID_AUTH", "鉴权失败")
	InvalidToken   = util.Error("INVALID_TOKEN", "令牌失效")
	BusinessError  = util.Error("BUSINESS_ERROR", "业务异常")
	OperationError = util.Error("OPERATION_ERROR", "操作失败")
)

package errcode

var (
	GatewayError   = New("GATEWAY_ERROR", "网关错误")
	SystemError    = New("SYSTEM_ERROR", "系统异常")
	SystemBusy     = New("SYSTEM_BUSY", "系统繁忙")
	InternalError  = New("INTERNAL_ERROR", "内部错误")
	InvalidRequest = New("INVALID_REQUEST", "无效请求")
	InvalidEvent   = New("INVALID_EVENT", "无效事件")
	InvalidParam   = New("INVALID_PARAM", "参数错误")
	InvalidSign    = New("INVALID_SIGN", "验签失败")
	InvalidAuth    = New("INVALID_AUTH", "鉴权失败")
	InvalidToken   = New("INVALID_TOKEN", "令牌失效")
	BusinessError  = New("BUSINESS_ERROR", "业务异常")
	OperationError = New("OPERATION_ERROR", "操作失败")
	UnknownError   = New("UNKNOWN_ERROR", "未知错误")
)

type ErrCode struct {
	code    int
	subCode string
	message string
	detail  any
}

func (e ErrCode) Code() int {
	return e.code
}

func (e ErrCode) SubCode() string {
	return e.subCode
}

func (e ErrCode) Message() string {
	return e.message
}

func (e ErrCode) Detail() any {
	return e.detail
}

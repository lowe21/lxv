package errcode

var (
	ErrGateway         = New("GATEWAY_ERROR", "网关错误")
	ErrSystem          = New("SYSTEM_ERROR", "系统错误")
	ErrSystemBusy      = New("SYSTEM_BUSY", "系统繁忙")
	ErrInternal        = New("INTERNAL_ERROR", "内部错误")
	ErrInvalidRequest  = New("INVALID_REQUEST", "无效请求")
	ErrInvalidEvent    = New("INVALID_EVENT", "无效事件")
	ErrInvalidParam    = New("INVALID_PARAM", "无效参数")
	ErrInvalidSign     = New("INVALID_SIGN", "无效签名")
	ErrInvalidToken    = New("INVALID_TOKEN", "无效令牌")
	ErrAuthFailed      = New("AUTH_FAILED", "鉴权失败")
	ErrBusinessFailed  = New("BUSINESS_FAILED", "业务失败")
	ErrOperationFailed = New("OPERATION_FAILED", "操作失败")
	ErrUnknown         = New("UNKNOWN_ERROR", "未知错误")
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

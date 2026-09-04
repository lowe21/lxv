package common

type (
	APIReq struct {
		RequestID string `json:"requestID" valid:"required"      description:"请求ID"`
		Content   string `json:"content"   valid:"required|json" description:"请求内容"`
		Timestamp string `json:"timestamp" valid:"required"      description:"时间戳，单位：毫秒"`
		Sign      string `json:"sign"      valid:""              description:"签名"`
	}
	APIRes struct {
		Code    string `json:"code"    description:"返回码"`
		Message string `json:"message" description:"返回信息"`
		Data    any    `json:"data"    description:"返回数据"`
	}
)

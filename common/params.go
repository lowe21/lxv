package common

type (
	Req struct {
		AppId     string `json:"appId"     valid:"required"      description:"应用ID"`
		RequestId string `json:"requestId" valid:"required"      description:"请求ID"`
		Content   string `json:"content"   valid:"required|json" description:"请求内容"`
		Timestamp string `json:"timestamp" valid:"required"      description:"时间戳，单位：毫秒"`
		Sign      string `json:"sign"      valid:"required"      description:"签名"`
	}
	Ret struct {
		Content string `json:"content"        description:"响应内容"`
		Sign    string `json:"sign,omitempty" description:"签名"`
	}
	Res struct {
		Code    string `json:"code"    description:"返回码"`
		Message string `json:"message" description:"返回信息"`
		Data    any    `json:"data"    description:"返回数据"`
	}
)

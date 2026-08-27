package socket

import (
	"github.com/gogf/gf/v2/encoding/gjson"

	"github.com/lowe21/lxv/pkg/errcode"
)

type (
	Input struct {
		Id    string `json:"id"    valid:"required"`
		Event string `json:"event" valid:"required"`
		Data  any    `json:"data"`
	}
	Output struct {
		Id      string `json:"id"`
		Event   string `json:"event"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
)

func Message(id, event string, args ...any) []byte {
	var (
		subCode string
		message string
		data    any
	)

	if argsLen := len(args); argsLen > 0 {
		switch arg := args[0].(type) {
		case error:
			subCode, message = errcode.Parse(arg)
		default:
			data = arg
		}
	}

	return gjson.MustEncode(&Output{
		Id:      id,
		Event:   event,
		Code:    subCode,
		Message: message,
		Data:    data,
	})
}

package errcode

import (
	"errors"

	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/hessian2"
	"dubbo.apache.org/dubbo-go/v3/protocol/triple/triple_protocol"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
)

func New(args ...any) error {
	var (
		code    int
		subCode string
		message string
		detail  any
	)

	if argsLen := len(args); argsLen > 0 {
		switch arg := args[0].(type) {
		case error:
			if exception, ok := errors.AsType[*hessian2.GenericException](arg); ok {
				desc := gstr.StrEx(exception.Error(), "desc = ")
				code = gcode.CodeNil.Code()
				subCode = gstr.StrTillEx(desc, "@")
				message = gstr.StrEx(desc, "@")
				detail = desc
			} else if err, ok := errors.AsType[*triple_protocol.Error](arg); ok {
				code = gcode.CodeNil.Code()
				subCode = gstr.StrTillEx(err.Message(), "@")
				message = gstr.StrEx(err.Message(), "@")
			} else {
				switch errorCode := gerror.Code(arg).(type) {
				case ErrCode:
					code = errorCode.Code()
					subCode = errorCode.SubCode()
					detail = errorCode.Detail()
				case gcode.Code:
					code = errorCode.Code()
					subCode = errorCode.Message()
					detail = errorCode.Detail()
				}
				message = arg.Error()
			}
		case ErrCode:
			code = arg.Code()
			subCode = arg.SubCode()
			message = arg.Message()
			detail = arg.Detail()
		case gcode.Code:
			code = arg.Code()
			subCode = arg.Message()
			message = arg.Message()
			detail = arg.Detail()
		case string:
			if argsLen > 1 {
				code = gcode.CodeNil.Code()
				subCode = arg
			} else {
				subCode = gcode.CodeInternalError.Message()
				message = arg
			}
		}

		if argsLen > 1 {
			if arg, ok := args[1].(string); ok {
				message = arg
			}
		}
		if argsLen > 2 {
			if arg, ok := args[2].(string); ok {
				detail = arg
			}
		}
	}

	if code == 0 {
		code = gcode.CodeInternalError.Code()
	}
	if subCode == "" {
		subCode = gcode.CodeInternalError.Message()
	}
	if message == "" {
		message = gcode.CodeUnknown.Message()
	}

	return gerror.NewCode(ErrCode{
		code:    code,
		subCode: gstr.CaseSnakeScreaming(subCode),
		message: message,
		detail:  detail,
	}, message)
}

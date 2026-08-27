package errcode

import (
	"errors"

	"github.com/go-redsync/redsync/v4"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func Parse(err error) (subCode, message string) {
	if err == nil {
		err = UnknownError
	}

	errorCode := gerror.Code(err)
	if _, ok := errorCode.(ErrCode); !ok {
		switch errorCode.Code() {
		case gcode.CodeInternalPanic.Code():
			err = SystemError
		case gcode.CodeInternalError.Code():
			err = InternalError
		case gcode.CodeDbOperationError.Code():
			err = OperationError
		case gcode.CodeValidationFailed.Code():
			err = New(InvalidParam, err.Error())
		default:
			if _, ok = errors.AsType[*redsync.ErrTaken](err); ok {
				err = SystemBusy
			} else if _, ok = errors.AsType[*redsync.ErrNodeTaken](err); ok {
				err = SystemBusy
			} else {
				lastErr := err
				for {
					nextErr := gerror.Unwrap(lastErr)
					if nextErr == nil {
						break
					}
					lastErr = nextErr
				}
				err = New(errorCode, lastErr.Error())
			}
		}
	}

	errCode := gerror.Code(err).(ErrCode)
	subCode = errCode.SubCode()
	message = errCode.Message()

	return
}

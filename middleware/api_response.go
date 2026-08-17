package middleware

import (
	"errors"
	"fmt"
	"mime"
	"net/http"

	"github.com/go-redsync/redsync/v4"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/lowe21/lxv/common"
	"github.com/lowe21/lxv/pkg/error_code"
)

var streamType = gset.NewFrom([]string{
	"text/event-stream",
	"application/octet-stream",
	"multipart/x-mixed-replace",
})

func ApiResponse(request *ghttp.Request) {
	if request.GetError() == nil {
		request.Middleware.Next()
	}

	if request.IsExited() || request.Response.IsHeaderWrote() {
		return
	}

	mediaType, _, _ := mime.ParseMediaType(request.Response.Header().Get("Content-Type"))
	if streamType.Contains(mediaType) {
		return
	}

	var (
		ctx     = request.GetCtx()
		err     = request.GetError()
		subCode string
		message string
		data    any
	)

	if err != nil {
		request.Response.ClearBuffer()

		g.Log().Error(ctx, err, request.RequestURI, request.GetBodyString())

		errCode := gerror.Code(err)
		switch errCode.Code() {
		case gcode.CodeInternalPanic.Code():
			err = error_code.SystemError
		case gcode.CodeInternalError.Code():
			err = error_code.InternalError
		case gcode.CodeDbOperationError.Code():
			err = error_code.OperationError
		case gcode.CodeValidationFailed.Code():
			err = error_code.New(error_code.InvalidParam, err.Error())
		default:
			if _, ok := errCode.(error_code.ErrorCode); !ok {
				if _, ok = errors.AsType[*redsync.ErrTaken](err); ok {
					err = error_code.SystemBusy
				} else if _, ok = errors.AsType[*redsync.ErrNodeTaken](err); ok {
					err = error_code.SystemBusy
				} else {
					if lastErr := gerror.Unwrap(err); lastErr != nil {
						err = lastErr
					}
					err = error_code.New(errCode, err.Error())
				}
			}
		}

		errorCode := gerror.Code(err).(error_code.ErrorCode)
		subCode = errorCode.SubCode()
		message = errorCode.Message()

		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.SetStatus(codes.Error, message)
		}
	} else {
		if request.Response.Status >= http.StatusBadRequest {
			request.Response.ClearBuffer()
			subCode = gerror.Code(error_code.GatewayError).(error_code.ErrorCode).SubCode()
			message = fmt.Sprintf("HTTP %d %s", request.Response.Status, http.StatusText(request.Response.Status))
		} else {
			if request.Response.Status >= http.StatusMultipleChoices {
				return
			}
			if request.Response.BufferLength() > 0 || request.Response.BytesWritten() > 0 {
				return
			}

			data = request.GetHandlerResponse()
		}
	}

	request.Response.WriteJson(&common.ApiRes{
		Code:    subCode,
		Message: message,
		Data:    data,
	})
}

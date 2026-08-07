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
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/common"
	"github.com/lowe21/lxv/util"
)

var streamType = gset.NewFrom([]string{
	"text/event-stream",
	"application/octet-stream",
	"multipart/x-mixed-replace",
})

func Response(request *ghttp.Request) {
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
		if request.Response.BufferLength() > 0 {
			request.Response.ClearBuffer()
		}

		g.Log().Error(ctx, err, request.RequestURI, request.GetBodyString())

		errCode := gerror.Code(err)
		switch errCode.Code() {
		case gcode.CodeInternalPanic.Code():
			err = common.SystemError
		case gcode.CodeInternalError.Code():
			err = common.InternalError
		case gcode.CodeDbOperationError.Code():
			err = common.OperationError
		case gcode.CodeValidationFailed.Code():
			err = util.Error(common.InvalidParam, err.Error())
		default:
			if _, ok := errCode.(util.ErrorCode); !ok {
				if _, ok = errors.AsType[*redsync.ErrNodeTaken](err); ok {
					err = common.SystemBusy
				} else {
					if lastErr := gerror.Unwrap(err); lastErr != nil {
						err = lastErr
					}
					err = util.Error(errCode, err.Error())
				}
			}
		}

		errorCode := gerror.Code(err).(util.ErrorCode)
		subCode = errorCode.SubCode()
		message = errorCode.Message()

		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.SetStatus(codes.Error, message)
		}
	} else {
		if request.Response.Status > 0 && request.Response.Status != http.StatusOK {
			subCode = gerror.Code(common.GatewayError).(util.ErrorCode).SubCode()
			if request.Response.BufferLength() > 0 {
				message = request.Response.BufferString()
				request.Response.ClearBuffer()
			} else {
				message = http.StatusText(request.Response.Status)
			}
			message = fmt.Sprintf("HTTP %d %s", request.Response.Status, message)
		} else {
			if request.Response.BufferLength() > 0 || request.Response.BytesWritten() > 0 {
				return
			}

			data = request.GetHandlerResponse()
		}
	}

	ret := &common.Ret{
		Content: gjson.New(&common.Res{
			Code:    subCode,
			Message: message,
			Data:    data,
		}).MustToJsonString(),
	}

	req := &common.Req{}
	if err = gconv.Scan(request.GetRequestMap(), req); err != nil {
		g.Log().Error(ctx, err)
	}
	if req.AppId != "" {
		privateKey := g.Config().MustGet(ctx, gstr.Join([]string{"certificate", req.AppId, "privateKey"}, ".")).String()
		sign, err := util.Rsa2Sign(privateKey, ret.Content)
		if err != nil {
			g.Log().Error(ctx, err)
		}
		ret.Sign = sign
	}

	request.Response.WriteJson(ret)
}

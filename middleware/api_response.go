package middleware

import (
	"fmt"
	"mime"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/lowe21/lxv/common"
	"github.com/lowe21/lxv/pkg/errcode"
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
		subCode, message = errcode.Parse(err)
		g.Log().Error(ctx, err, request.RequestURI, request.GetBodyString())

		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.SetStatus(codes.Error, message)
		}
	} else {
		if request.Response.Status >= http.StatusBadRequest {
			request.Response.ClearBuffer()
			subCode, _ = errcode.Parse(errcode.ErrGateway)
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

package middleware

import (
	"fmt"
	"mime"
	"net/http"
	"slices"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/lowe21/lxv/common"
	"github.com/lowe21/lxv/pkg/errcode"
)

var (
	streamTypes = []string{
		"text/event-stream",
		"application/octet-stream",
		"multipart/x-mixed-replace",
	}
	statusMap = func() map[string]int {
		data := make(map[string]int)
		for err, status := range map[error]int{
			errcode.ErrGateway:          http.StatusBadGateway,
			errcode.ErrSystemBusy:       http.StatusServiceUnavailable,
			errcode.ErrInvalidRequest:   http.StatusBadRequest,
			errcode.ErrInvalidEvent:     http.StatusBadRequest,
			errcode.ErrInvalidParam:     http.StatusBadRequest,
			errcode.ErrInvalidSign:      http.StatusBadRequest,
			errcode.ErrInvalidToken:     http.StatusUnauthorized,
			errcode.ErrAuthFailed:       http.StatusUnauthorized,
			errcode.ErrValidationFailed: http.StatusBadRequest,
			errcode.ErrPermissionDenied: http.StatusForbidden,
		} {
			subCode, _ := errcode.Parse(err)
			data[subCode] = status
		}

		return data
	}()
)

func APIResponse(request *ghttp.Request) {
	if request.GetError() == nil {
		request.Middleware.Next()
	}

	if request.IsExited() || request.Response.IsHeaderWrote() {
		return
	}

	mediaType, _, _ := mime.ParseMediaType(request.Response.Header().Get("Content-Type"))
	if slices.Contains(streamTypes, mediaType) {
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
		g.Log().Error(ctx, err, request.RequestURI, request.GetBodyString())
		request.Response.ClearBuffer()
		subCode, message = errcode.Parse(err)
		status, ok := statusMap[subCode]
		if !ok {
			status = http.StatusInternalServerError
		}
		request.Response.WriteHeader(status)
	} else {
		if request.Response.Status >= http.StatusMultipleChoices {
			if request.Response.Status >= http.StatusBadRequest {
				request.Response.ClearBuffer()
				subCode, _ = errcode.Parse(errcode.ErrGateway)
				message = fmt.Sprintf("HTTP %d %s", request.Response.Status, http.StatusText(request.Response.Status))
			} else {
				return
			}
		} else {
			if request.Response.BufferLength() > 0 || request.Response.BytesWritten() > 0 {
				return
			}
			data = request.GetHandlerResponse()
		}
	}

	request.Response.WriteJson(&common.APIRes{
		Code:    subCode,
		Message: message,
		Data:    data,
	})
}

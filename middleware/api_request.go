package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"slices"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/common"
	"github.com/lowe21/lxv/pkg/errcode"
	"github.com/lowe21/lxv/pkg/jwt"
	"github.com/lowe21/lxv/pkg/validation"
)

type (
	AuthHandler func(ctx context.Context, token string, leeway bool) (*jwt.Payload, error)
	PreHandler  func(ctx context.Context, req *common.APIReq) error
)

func APIRequest(authHandler AuthHandler, preHandler PreHandler) ghttp.HandlerFunc {
	return func(request *ghttp.Request) {
		err := request.GetError()
		if err != nil {
			request.Middleware.Next()
			return
		}
		defer func() {
			request.SetError(err)
			request.Middleware.Next()
		}()

		handler := request.GetServeHandler()
		if handler == nil || !handler.Handler.Info.IsStrictRoute {
			return
		}

		ctx := request.GetCtx()
		payload := &jwt.Payload{}

		if gconv.Bool(handler.GetMetaTag("notify")) {
			return
		}
		if gconv.Bool(handler.GetMetaTag("auth")) {
			authorization := request.Header.Get("Authorization")
			if authorization == "" {
				err = errcode.New(errcode.ErrAuthFailed, "Authorization header is empty")
				return
			}

			fields := gstr.Fields(authorization)
			if len(fields) != 2 || !gstr.Equal(fields[0], "Bearer") {
				err = errcode.New(errcode.ErrAuthFailed, "Authorization header is invalid")
				return
			}
			if fields[1] == "" {
				err = errcode.New(errcode.ErrAuthFailed, "Bearer token is empty")
				return
			}

			if authHandler == nil {
				err = errcode.New(errcode.ErrAuthFailed, "authHandler is nil")
				return
			}

			payload, err = authHandler(ctx, fields[1], gconv.Bool(handler.GetMetaTag("leeway")))
			if err != nil {
				return
			}
			if payload == nil {
				err = errcode.New(errcode.ErrAuthFailed, "payload is nil")
				return
			}
			if payload.SessionKey == "" {
				err = errcode.New(errcode.ErrAuthFailed, "sessionKey is empty")
				return
			}
		}
		if gconv.Bool(handler.GetMetaTag("upload")) {
			return
		}

		req := &common.APIReq{}
		if err = validation.Validator(ctx, req); err != nil {
			err = errcode.New(errcode.ErrInvalidParam, err.Error())
			return
		}

		if payload.SessionKey != "" {
			reqMap := gconv.MapStrStr(req)
			reqValues := make([]string, 0, len(reqMap))
			for _, key := range slices.Sorted(maps.Keys(reqMap)) {
				if key != "" && key != "sign" {
					reqValues = append(reqValues, key+"="+reqMap[key])
				}
			}
			hash := hmac.New(sha256.New, []byte(payload.SessionKey))
			if _, err = hash.Write([]byte(gstr.Join(reqValues, "&"))); err != nil {
				return
			}
			sign, decodeErr := hex.DecodeString(req.Sign)
			if decodeErr != nil || !hmac.Equal(hash.Sum(nil), sign) {
				err = errcode.ErrInvalidSign
				return
			}
		}

		if preHandler != nil {
			if err = preHandler(ctx, req); err != nil {
				return
			}
		}

		setParam := func(data any) {
			dataMap := gconv.Map(data)
			for _, key := range slices.Sorted(maps.Keys(dataMap)) {
				if newKey := gstr.CaseCamelLower(key); newKey != key {
					if _, ok := dataMap[newKey]; !ok {
						dataMap[newKey] = dataMap[key]
					}
					delete(dataMap, key)
				}
			}
			request.SetParamMap(dataMap)
		}

		setParam(req.Content)
		if payload.SessionKey != "" {
			setParam(payload)
		}
	}
}

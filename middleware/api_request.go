package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/common"
	"github.com/lowe21/lxv/pkg/errcode"
	"github.com/lowe21/lxv/pkg/jwt"
	"github.com/lowe21/lxv/pkg/validation"
)

type (
	AuthFunc func(ctx context.Context, token string, refresh bool) (payload *jwt.Payload, sessionKey string, err error)
	IdemFunc func(ctx context.Context, requestId, timestamp string) (err error)
)

func ApiRequest(authFunc AuthFunc, idemFunc IdemFunc) ghttp.HandlerFunc {
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

		var (
			ctx        = request.GetCtx()
			sessionKey string
		)

		handler := request.GetServeHandler()
		if handler != nil {
			if !handler.Handler.Info.IsStrictRoute {
				return
			}
			if handler.GetMetaTag("notify") != "" {
				return
			}
			if handler.GetMetaTag("auth") != "" {
				authorization := request.Header.Get("Authorization")
				if authorization == "" {
					err = errcode.New(errcode.InvalidAuth, "Authorization header is empty")
					return
				}

				parts := gstr.Split(authorization, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					err = errcode.New(errcode.InvalidAuth, "Authorization header is invalid")
					return
				}

				if authFunc == nil {
					err = errcode.New(errcode.InvalidAuth, "authFunc is nil")
					return
				}

				payload := &jwt.Payload{}
				payload, sessionKey, err = authFunc(ctx, parts[1], handler.GetMetaTag("refresh") != "")
				if err != nil {
					return
				}
				if sessionKey == "" {
					err = errcode.New(errcode.InvalidAuth, "sessionKey is empty")
					return
				}

				defer func() {
					if err == nil && payload != nil {
						request.SetParamMap(gconv.Map(payload))
					}
				}()
			}
			if handler.GetMetaTag("upload") != "" {
				multipartForm := request.GetMultipartForm()
				if multipartForm != nil && len(multipartForm.File) > 0 {
					return
				}
			}
		} else {
			return
		}

		req := &common.ApiReq{}
		if err = validation.Validator(ctx, req); err != nil {
			err = errcode.New(errcode.InvalidParam, err.Error())
			return
		}

		if sessionKey != "" {
			reqMap := gconv.MapStrStr(req)
			reqKeys := make([]string, 0, len(reqMap))
			for key := range reqMap {
				if key == "sign" {
					continue
				}
				reqKeys = append(reqKeys, key)
			}
			sort.Strings(reqKeys)
			reqValues := make([]string, 0, len(reqKeys))
			for _, key := range reqKeys {
				reqValues = append(reqValues, gstr.Join([]string{key, reqMap[key]}, "="))
			}
			reqString := gstr.Join(reqValues, "&")

			hash := hmac.New(sha256.New, []byte(sessionKey))
			if _, err = hash.Write([]byte(reqString)); err != nil {
				return
			}
			sign, _ := hex.DecodeString(req.Sign)
			if !hmac.Equal(hash.Sum(nil), sign) {
				err = errcode.InvalidSign
				return
			}
		}

		if idemFunc != nil {
			if err = idemFunc(ctx, req.RequestId, req.Timestamp); err != nil {
				return
			}
		}

		content := gconv.Map(req.Content)
		for key, value := range content {
			newKey := gstr.CaseCamelLower(key)
			if newKey != key {
				content[newKey] = value
				delete(content, key)
			}
		}
		request.SetParamMap(content)
	}
}

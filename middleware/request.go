package middleware

import (
	"sort"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/lowe21/lxv/common"
	"github.com/lowe21/lxv/util"
)

func Request(request *ghttp.Request) {
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
	if handler != nil {
		if !handler.Handler.Info.IsStrictRoute || handler.GetMetaTag("notify") != "" {
			return
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

	ctx := request.GetCtx()

	req := &common.Req{}
	if err = util.Validator(ctx, req); err != nil {
		err = util.Error(common.InvalidParam, err.Error())
		return
	}

	if !gjson.Valid(req.Content) {
		err = common.InvalidParam
		return
	}

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

	publicKey := g.Config().MustGet(ctx, gstr.Join([]string{"certificate", req.AppId, "publicKey"}, ".")).String()
	if util.Rsa2Verify(publicKey, reqString, req.Sign) != nil {
		err = common.InvalidSign
		return
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

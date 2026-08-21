package validation

import (
	"context"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gvalid"
)

func Validator(ctx context.Context, pointer any, args ...any) (err error) {
	data := gmap.New()

	if len(args) > 0 {
		for _, arg := range args {
			for key, value := range gconv.Map(arg) {
				data.Set(key, value)
			}
		}
	} else {
		if request := ghttp.RequestFromCtx(ctx); request != nil {
			for key, value := range request.GetRequestMap() {
				data.Set(key, value)
			}
		} else {
			for key, value := range gconv.Map(pointer) {
				data.Set(key, value)
			}
		}
	}

	if err = gvalid.New().Bail().Data(pointer).Assoc(data).Run(ctx); err != nil {
		return
	}

	return gconv.Scan(data, pointer)
}

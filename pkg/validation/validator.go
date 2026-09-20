package validation

import (
	"context"
	"maps"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gvalid"
)

func Validator(ctx context.Context, pointer any, values ...any) (err error) {
	data := make(map[string]any)

	if len(values) > 0 {
		for _, value := range values {
			maps.Copy(data, gconv.Map(value))
		}
	} else {
		if request := ghttp.RequestFromCtx(ctx); request != nil {
			data = request.GetRequestMap()
		} else {
			data = gconv.Map(pointer)
		}
	}

	if err = gvalid.New().Bail().Data(pointer).Assoc(data).Run(ctx); err != nil {
		return
	}

	return gconv.Scan(data, pointer)
}

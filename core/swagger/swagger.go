package swagger

import (
	"fmt"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"

	"github.com/lowe21/lxv/common"
)

var once sync.Once

func Init() {
	once.Do(func() {
		openapiPath := g.Config().MustGet(nil, "server.openapiPath").String()
		swaggerPath := g.Config().MustGet(nil, "server.swaggerPath").String()
		if openapiPath != "" && swaggerPath != "" {
			server := g.Server()
			server.SetSwaggerUITemplate(fmt.Sprintf(`
<!doctype html>
<html>
	<head>
		<title>%s API Reference</title>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<div id="app"></div>
		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
		<script>
			Scalar.createApiReference('#app', {
				url: '{SwaggerUIDocUrl}',
				proxyUrl: 'https://proxy.scalar.com'
			})
		</script>
	</body>
</html>
`, server.GetName()))

			swaggerUser := g.Config().MustGet(nil, "server.swaggerUser").String()
			swaggerPass := g.Config().MustGet(nil, "server.swaggerPass").String()
			if swaggerUser != "" && swaggerPass != "" {
				server.BindHookHandler(server.GetOpenApiPath(), ghttp.HookBeforeServe, func(request *ghttp.Request) {
					if !request.BasicAuth(swaggerUser, swaggerPass) {
						request.ExitAll()
					}
				})
			}

			openApi := server.GetOpenApi()
			openApi.Config.CommonRequest = &common.ApiReq{}
			openApi.Config.CommonRequestDataField = "Content"
			openApi.Config.CommonResponse = &common.ApiRes{}
			openApi.Config.CommonResponseDataField = "Data"
			openApi.Info = goai.Info{
				Title:       fmt.Sprintf("%s API Reference", server.GetName()),
				Description: "Special note: The `content` in the request parameter needs to be passed in as a JSON string, not an object",
			}
			openApi.Components = goai.Components{
				SecuritySchemes: goai.SecuritySchemes{
					"jwt": goai.SecuritySchemeRef{
						Value: &goai.SecurityScheme{
							Type:         "http",
							Scheme:       "bearer",
							BearerFormat: "Bearer Token",
						},
					},
				},
			}
		}
	})
}

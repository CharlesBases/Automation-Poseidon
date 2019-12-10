package template

const controllerTemplate = `// this file is generated for {{.Name}}
package {{package}}

import (
	"net/http"

	"gitlab.ifchange.com/bot/gokitcommon/log"
	"gitlab.ifchange.com/bot/gokitcommon/web"
	"gopkg.in/go-playground/validator.v9"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	{{imports}}
)

/**
@api {post} url title
@apiVersion 1.0.0
@apiGroup {{.Group}}
@apiDescription 描述

@apiParam {Object} p 请求参数
{{parseRequestParams .Params}}
@apiParamExample {json} Request-Example:
{{jsonDemo .Params}}

@apiSuccess {Object} results 返回结果
{{parseResponseParams .Results}}
@apiSuccessExample {json} Response-Example:
{{jsonDemo .Results}}
*/
func (*{{service}}) {{.Name}}({{requestParse}}) ({{responseParse}}) {
	defer func() {
		if response.ErrNo != 0 {
			response.Results = nil
		}
	}()
	// response
	{{responseAssignment}}
	// validator
	if err := validator.New().Struct(request); err != nil {
		log.Error("json validator error: ", err)
		response.ErrNo, response.ErrMsg = web.WebError(web.ParamsErr)
		return
	}
	// Do something
	{{business}}

	return 
}

func {{.Name}}() http.Handler {
	return httptransport.NewServer(
		MakeEndpoint{{.Name}}({{service}}service),
		DecodeRequest{{.Name}},
		EncodeResponse{{.Name}},
	)
}

func MakeEndpoint{{.Name}}(svc {{interface}}) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		resp := svc.{{.Name}}({{requestCoercive}})
		return resp, nil
	}
}

func DecodeRequest{{.Name}}(ctx context.Context, r *http.Request) (interface{}, error) { {{requestDecode}} }

func EncodeResponse{{.Name}}(ctx context.Context, rw http.ResponseWriter, response interface{}) error {
	return web.EncodeResponse(ctx, rw, response)
}
`

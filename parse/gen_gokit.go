package parse

import (
	"fmt"
	"html/template"
	"io"
	"path"
	"strings"

	log "github.com/cihub/seelog"
)

const KitTemplate = `// this file is generated for {{.Name}} {{$Request := parseRequest}} {{$Response := parseResponse}}
package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"gitlab.ifchange.com/bot/gokitcommon/web"
	"gopkg.in/go-playground/validator.v9"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	{{genimports}}
)

/*
@api {post} /interviews/get  获取试题
@apiVersion 0.1.0
@apiGroup Interviews
@apiDescription 获取试题 一套题（uuid）包含多张试卷（interview_uuid），interview_id为试卷的类型

@apiParam {Object} p 请求参数
@apiParam {String} uuid 套题id
@apiParam {Number} interview_id 类型id(1素质测评 2潜力测评)
@apiParam {Object} param 参数

@apiParamExample {json} Request-Example:
{

}

@apiSuccess {Object} results 返回结果
@apiSuccess {Bool} result.interviews_info 全部试卷
@apiSuccess {Bool} result.interviews_info.status 试卷状态 未开始 1已开始 2完成
@apiSuccess {Bool} result.interviews_info.uuid 测评试卷id
@apiSuccess {Bool} result.interviews_info.created_at 测评试卷创建时间，10位时间戳
@apiSuccess {Bool} result.cur_interview_id 当前试卷类型 1性格测评 5潜力测评
@apiSuccess {Bool} result.cur_interview_uuid 当前试卷id
@apiSuccess {Bool} result.data 测评数据

@apiSuccessExample {json} Response-Example:
{

}
*/
func (*{{service}}) {{.Name}}(request *{{$Request}}) (response *{{$Response}}) {
	// response.Results
	results := {{results}}
	// response.Error
	Error := new(web.Error)
	defer func() {
		if Error.ErrNo != 0 {
			response = &{{$Response}}{
				Error: core.Error{
					ErrNo:  Error.ErrNo,
					ErrMsg: Error.ErrMsg,
				},
			}
		} else {
			response = &{{$Response}}{
				Results: results,
			}
		}
	}()
	if err := validator.New().Struct(request); err != nil {
		Error.WebError(web.ParamsErr)
		return
	}

	// Do something

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
		req := request.(*{{$Request}})
		resp := svc.{{.Name}}(req)
		return resp, nil
	}
}

func DecodeRequest{{.Name}}(ctx context.Context, r *http.Request) (interface{}, error) {
	request := new({{$Request}})
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, err
	}
	return request, nil
}

func EncodeResponse{{.Name}}(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return web.Encode(w, response)
}
`

func (file *File) GenKitFile(Interface *Interface, Func *Func, wr io.Writer) {
	log.Info(fmt.Sprintf("generating %s files ...", Func.Name))
	temp := template.New(fmt.Sprintf("%s.go", Func.Name))
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return path.Base(path.Dir(file.GenPath))
		},
		"service": func() string {
			return file.Package
		},
		"interface": func() string {
			return fmt.Sprintf("%s.%s", path.Base(path.Dir(file.GenPath)), Interface.Name)
		},
		"parseRequest": func() string {
			return strings.TrimPrefix(Func.Params[0].GoType,"*")
		},
		"parseResponse": func() string {
			return strings.TrimPrefix(Func.Results[0].GoType, "*")
		},
		"genimports": func() template.HTML {
			imports := strings.Builder{}
			imports.WriteString("\n\t")
			imports.WriteString(fmt.Sprintf(`%s "%s"`, path.Base(path.Dir(file.GenPath)), path.Dir(file.GenPath)))
			for k, v := range file.ImportA {
				imports.WriteString("\n\t")
				imports.WriteString(fmt.Sprintf(`%s "%s"`, k, v))
			}
			return template.HTML(imports.String())
		},
		"results": func() (string) {
			return fmt.Sprintf("new(%s)",strings.ReplaceAll(strings.TrimPrefix(Func.Results[0].GoType,"*"),"Response","Resp") )
		},
	})
	kitTemplate, err := temp.Parse(KitTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	kitTemplate.Execute(wr, Func)
}

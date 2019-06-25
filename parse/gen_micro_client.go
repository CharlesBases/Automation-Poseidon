package parse

import (
	"io"
	"path/filepath"
	"text/template"

	log "github.com/cihub/seelog"
)

const ServiceMicroClient = `package main {{$PackagePB := .Package}} {{$PackageIN := .PkgPath | funcSort}}

import (
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/client/selector"
	"github.com/micro/go-micro/registry"
	wrapper "github.com/micro/go-plugins/wrapper/trace/opentracing"
	opentracing "github.com/opentracing/opentracing-go"

	"{{.PkgPath}}"
	{{$PackagePB}} "{{.GenPath}}"
)

var (
	service micro.Service
	opts    []client.CallOption
)

func init() {
	microClient()
}

func microClient() {
	service = grpc.NewService(
		micro.Name(SERVICE_NAME),
		micro.WrapCall(
			wrapper.NewCallWrapper(
				opentracing.GlobalTracer(),
			),
		),
		micro.WrapClient(
			wrapper.NewClientWrapper(
				opentracing.GlobalTracer(),
			),
		),
		micro.Selector(
			selector.NewSelector(
				selector.Registry(registry.DefaultRegistry),
				selector.SetStrategy(selector.RoundRobin),
			),
		),
	)

	opts = []client.CallOption{
		client.WithSelectOption(
			selector.WithFilter(
				selector.FilterVersion("v1.0"),
			),
		),
	}
}
{{range $index, $iface := .Interfaces}}
func New{{.Name}}() {{$PackageIN}}.{{.Name}} {
	return {{$PackagePB}}.New{{.Name}}Client_(
		{{$PackagePB}}.New{{.Name}}(
			SERVICE_NAME,
			service.Client(),
		),
		opts...,
	)
}
{{end}}
`

func (file *File) GenMicroClient(wr io.Writer) {
	log.Info("generating micro_client file ...")
	t := template.New("client.go")
	t.Funcs(template.FuncMap{
		"funcSort": func(Package string) string {
			return filepath.Base(Package)
		},
	})

	parsed, err := t.Parse(ServiceMicroClient)
	if err != nil {
		log.Error(err)
		return
	}
	parsed.Execute(wr, file)
}

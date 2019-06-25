package parse

import (
	"io"
	"text/template"

	log "github.com/cihub/seelog"
)

const ServiceMicroServer = `package main {{$PackagePB := .Package}}

import (
	log "github.com/cihub/seelog"
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	wrapper "github.com/micro/go-plugins/wrapper/trace/opentracing"
	"github.com/opentracing/opentracing-go"

	"{{.PkgPath}}"
	{{$PackagePB}} "{{.GenPath}}"
)

const SERVICE_NAME = ""

func main() {
	service := grpc.NewService(
		micro.Registry(registry.DefaultRegistry), 
		micro.Name(SERVICE_NAME), 
		micro.Version("v1.0"),
		micro.RegisterTTL(time.Minute), 
		micro.RegisterInterval(time.Second*40),
		micro.WrapHandler(wrapper.NewHandlerWrapper(opentracing.GlobalTracer())),
	)
{{range $index, $iface := .Interfaces}}
	{{$PackagePB}}.Register{{.Name}}Handler(service.Server(), {{$PackagePB}}.New{{.Name}}Server_(A.New{{.Name}}()))
{{end}}
	if err := service.Run(); err != nil {
		log.Error(err)
	}
}
`

func (file *File) GenMicroServer(wr io.Writer) {
	log.Info("generating micro_server file ...")
	t := template.New("server.go")

	parsed, err := t.Parse(ServiceMicroServer)
	if err != nil {
		log.Error(err)
		return
	}
	parsed.Execute(wr, file)
}

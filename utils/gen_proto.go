package utils

import (
	"html/template"
	"io"

	log "github.com/cihub/seelog"
)

const ProtoTemplate = `
`

func (file *File) GenProtoFile(wr io.Writer) {
	log.Info("generating .proto files ...")
	temp := template.New("pb.proto")
	temp.Funcs(template.FuncMap{
		"index": func(i int) int {
			return i + 1
		},
		"unescaped": func(x string) template.HTML {
			return template.HTML(x)
		},
	})
	protoTemplate, err := temp.Parse(ProtoTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	protoTemplate.Execute(wr, file)
}

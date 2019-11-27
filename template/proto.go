package template

import (
	"html/template"
	"io"

	log "github.com/cihub/seelog"
)

func (infor *Infor) GenerateProto(wr io.Writer) {
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
	protoTemplate, err := temp.Parse(protoTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	protoTemplate.Execute(wr, infor.File)
}

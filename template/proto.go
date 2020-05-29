package template

import (
	"html/template"
	"io"

	"charlesbases/Automation-Poseidon/utils"
)

func (*Base) GenerateProto(wr io.Writer) {
	utils.Info("generating .proto files ...")
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
	utils.ThrowCheck(err)
	protoTemplate.Execute(wr, poseidon)
}

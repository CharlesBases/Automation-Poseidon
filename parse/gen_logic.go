package parse

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"

	log "github.com/cihub/seelog"
)

const LogicTemplate = `package {{package}}

type {{.Group}} struct{}
`

func (file *File) GenLogicFile(Func *Func, wr io.Writer) {
	log.Info(fmt.Sprintf("generating logic for %s ...", Func.Group))
	temp := template.New(fmt.Sprintf("%s.go", Func.Group))
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return filepath.Base(file.GenLogicPath)
		},
	})
	logicTemplate, err := temp.Parse(LogicTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	logicTemplate.Execute(wr, Func)
}

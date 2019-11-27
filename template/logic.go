package template

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	log "github.com/cihub/seelog"
)

func (infor *Infor) GenerateLogic(wr io.Writer) {
	log.Info(fmt.Sprintf("generating logic for %s ...", infor.Func.Group))
	temp := template.New(fmt.Sprintf("%s.go", infor.Func.Group))
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return filepath.Base(strings.ToLower(infor.Func.Group))
		},
	})
	logicTemplate, err := temp.Parse(logicTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	logicTemplate.Execute(wr, infor.Func)
}

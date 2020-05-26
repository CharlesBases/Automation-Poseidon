package template

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"charlesbases/Automation-Poseidon/utils"
)

func (infor *Infor) GenerateLogic(wr io.Writer) {
	utils.Info(fmt.Sprintf("generating logic for %s ...", infor.Func.Group))
	temp := template.New(fmt.Sprintf("%s.go", infor.Func.Group))
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return filepath.Base(strings.ToLower(infor.Func.Group))
		},
	})
	logicTemplate, err := temp.Parse(logicTemplate)
	utils.ThrowCheck(err)
	logicTemplate.Execute(wr, infor.Func)
}

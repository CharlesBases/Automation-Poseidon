package template

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"charlesbases/Automation-Poseidon/utils"
)

func (logic *LogicInfor) GenerateLogic(wr io.Writer) {
	utils.Info(fmt.Sprintf("generating logic for %s ...", logic.Group))
	temp := template.New(fmt.Sprintf("%s.go", logic.Group))
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return filepath.Base(strings.ToLower(logic.Group))
		},
	})
	logicTemplate, err := temp.Parse(logicTemplate)
	utils.ThrowCheck(err)
	logicTemplate.Execute(wr, logic)
}

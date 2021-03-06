package template

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"charlesbases/Automation-Poseidon/utils"
)

func (*Base) GenerateImplement(wr io.Writer) {
	utils.Info("generating implement files ...")
	temp := template.New("implement.go")
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return filepath.Base(poseidon.GenInterPath)
		},
		"service": func(Interface utils.Interface) string {
			return strings.ToLower(strings.TrimSuffix(Interface.Name, "Service"))
		},
		"response": func(Interface utils.Interface) string {
			return fmt.Sprintf("%s.%s", filepath.Base(poseidon.PackagePath), Interface.Name)
		},
		"imports": func() template.HTML {
			sb := strings.Builder{}
			sb.WriteString(
				fmt.Sprintf(`"%s"`,
					strings.ReplaceAll(
						func() string {
							return fmt.Sprintf("%s/%s", poseidon.ProjectPath, filepath.Base(poseidon.PackagePath))
						}(),
						`\`, `/`,
					),
				),
			)
			return template.HTML(sb.String())
		},
		"interfaceName": func(Interface utils.Interface) string {
			return Interface.Name
		},
	})
	implementTemplate, err := temp.Parse(implementTemplate)
	utils.ThrowCheck(err)
	implementTemplate.Execute(wr, poseidon)
}

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"charlesbases/Automation-Poseidon/template"
	"charlesbases/Automation-Poseidon/utils"
)

var (
	sourceFile        = flag.String("file", ".", "full path of the interface file")                            // 源文件路径
	projectPath       = flag.String("project", "", "module path")                                              // go.mod 中项目路径
	generateInterPath = flag.String("interP", "../controllers/", "full path of the generate interface folder") // 路由层文件夹
	generateLogicPath = flag.String("logicP", "../logics/", "full path of the generate logics folder")         // 业务层文件夹

	generateProtoPath = flag.String("protoP", "./pb/", "full path of the generate rpc folder") // .proto 文件夹
	protoPackage      = flag.String("package", "pb", "package name in .proto file")            // .proto 文件包名
	generateProto     = flag.Bool("proto", false, "generate proto file or not")                // 是否生成 .proto 文件

	update  = flag.Bool("update", false, "update existing interface or not") // 是否更新接口
	context = flag.Bool("ctx", true, "import context or not")                // 是否导入 context
)

var (
	err     error
	src     string // $GOPATH/src
	astFile *ast.File
	config  *utils.File

	swg               = sync.WaitGroup{}
	astFileChannel    = make(chan struct{})
	sourceFileChannel = make(chan struct{})
)

func main() {
	flag.Parse()

	src = func() string {
		list := filepath.SplitList(os.Getenv("GOPATH"))
		for index := range list {
			return filepath.Join(list[index], "src")
		}
		return ""
	}()

	*sourceFile, _ = filepath.Abs(*sourceFile)
	utils.Info("parsing files for go: ", *sourceFile)

	go func() {
		astFile, err = parser.ParseFile(token.NewFileSet(), *sourceFile, nil, 0) // 获取文件信息
		utils.ThrowCheck(err)
		astFileChannel <- struct{}{}
	}()

	swg.Add(1)
	go func() {
		defer swg.Done()

		config = &utils.File{
			Name:        filepath.Base(*sourceFile),
			PackagePath: strings.TrimPrefix(filepath.Dir(*sourceFile), src)[1:],
			ProjectPath: func() string {
				if *projectPath != "" {
					return fmt.Sprintf("%s", *projectPath)
				}
				return filepath.Dir(strings.TrimPrefix(filepath.Dir(*sourceFile), src)[1:])
			}(),
			ProtoPackage: *protoPackage,
			GenProtoPath: func() string {
				abspath, err := filepath.Abs(*generateProtoPath)
				utils.ThrowCheck(err)
				if *generateProto {
					os.MkdirAll(abspath, 0755)
				}
				return strings.TrimPrefix(abspath, src)[1:]
			}(),
			GenInterPath: func() string {
				abspath, err := filepath.Abs(*generateInterPath)
				utils.ThrowCheck(err)
				os.MkdirAll(abspath, 0755)
				return strings.TrimPrefix(abspath, src)[1:]
			}(),
			GenLogicPath: func() string {
				abspath, err := filepath.Abs(*generateLogicPath)
				utils.ThrowCheck(err)
				os.MkdirAll(abspath, 0755)
				return strings.TrimPrefix(abspath, src)[1:]
			}(),

			ImportsA: make(map[string]string),
			Structs:  make(map[string]map[string][]utils.Field, 0),
		}

		sourceFileChannel <- struct{}{}
	}()

	<-sourceFileChannel
	<-astFileChannel

	if *context {
		config.ImportsA["context"] = "context"
	}

	config.ParseFile(astFile)
	if len(config.Interfaces) == 0 {
		utils.ThrowCheck(fmt.Errorf("no interface found"))
	}

	config.GoTypeConfig()

	// generate proto file
	swg.Add(1)
	go func() {
		defer swg.Done()

		if *generateProto {
			protofile := filepath.Join(config.GenProtoPath, fmt.Sprintf("%s.proto", *protoPackage))
			protoFile, err := createFile(protofile)
			utils.ThrowCheck(err)
			defer protoFile.Close()

			infor := &template.Infor{
				File: config,
			}
			infor.GenerateProto(protoFile)

			absprotofile := filepath.Join(src, protofile)
			dir := filepath.Dir(absprotofile)

			// run protoc
			utils.Info("run the protoc command ...")
			_, err = exec.Command(
				"protoc",
				fmt.Sprintf("--proto_path=%s", dir),
				fmt.Sprintf("--gogofaster_out=plugins=grpc:%s", dir),
				absprotofile,
			).CombinedOutput()
			utils.ThrowCheck(err)
			utils.Info("protoc complete !")
		}
	}()

	// gen implement
	swg.Add(1)
	go func() {
		defer swg.Done()

		implementFile, err := createFile(filepath.Join(config.GenInterPath, "implement.go"))
		utils.ThrowCheck(err)
		infor := &template.Infor{File: config}
		infor.GenerateImplement(implementFile)
	}()

	// gen func
	for _, Interface := range config.Interfaces {
		for _, Func := range Interface.Funcs {
			swg.Add(2)

			// logic
			go func(f utils.Func) {
				defer swg.Done()

				logicdir := filepath.Join(config.GenLogicPath, strings.ToLower(f.Group))
				currentlogicfile := filepath.Join(logicdir, fmt.Sprintf("%s.go", strings.ToLower(f.Group)))

				os.MkdirAll(filepath.Join(src, logicdir), 0755)

				if !isExit(currentlogicfile) {
					logicfile, err := createFile(currentlogicfile)
					utils.ThrowCheck(err)
					infor := &template.Infor{
						File: config,
						Func: &f,
					}
					infor.GenerateLogic(logicfile)
				}
			}(Func)

			// controllers
			go func(i utils.Interface, f utils.Func) {
				defer swg.Done()

				currentfile := filepath.Join(config.GenInterPath, fmt.Sprintf("%s.go", utils.Snake(f.Name)))

				if !*update && isExit(currentfile) {
					return
				}
				controllerFile, err := createFile(currentfile)
				utils.ThrowCheck(err)
				infor := &template.Infor{
					File:      config,
					Interface: &i,
					Func:      &f,
				}
				infor.GenerateController(controllerFile)

			}(Interface, Func)
		}
	}

	swg.Wait()

	gofmt(filepath.Join(src, config.GenInterPath))
	utils.Info("complete!")
}

func isExit(filename string) bool {
	if _, err := os.Stat(filepath.Join(src, filename)); err == nil {
		return true
	}
	return false
}

func createFile(fileName string) (*os.File, error) {
	utils.Info("create file: " + fileName)
	fileName = filepath.Join(src, fileName)
	os.RemoveAll(fileName)
	return os.Create(fileName)
}

func gofmt(filepath string) {
	exec.Command("gofmt", "-l", "-w", "-s", filepath).Run()
}

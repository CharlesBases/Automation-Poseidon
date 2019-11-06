package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/cihub/seelog"

	"github.com/CharlesBases/proto/parse"
)

var (
	sourceFile   = flag.String("file", "", "full path of the file")
	generatePath = flag.String("path", "./pb/", "full path of the generate folder")
	protoPackage = flag.String("package", "", "package name in .proto file")
	genProto     = flag.Bool("proto", false, "generate proto file or not")
)

var (
	proFile       = "proto"
	controllerPkg = "controllers"
)

func init() {
	logger, _ := log.LoggerFromConfigAsString(`
			<?xml version="1.0" encoding="utf-8" ?>
			<seelog levels="info,error">
				<outputs formatid="main">
					<filter levels="warn,info">
						<console formatid="main"/>
					</filter>
					<filter levels="error,critical">
						<console formatid="error"/>
					</filter>
				</outputs>
				<formats>
					<format id="main" format="[%Date(2006-01-02 15:04:05.000)][%LEV] ==&gt; %Msg%n"/>
					<format id="error" format="%EscM(31)[%Date(2006-01-02 15:04:05.000)][%LEV] ==&gt; %Msg%n%EscM(0)"/>
				</formats>
			</seelog>`)
	log.ReplaceLogger(logger)
}

func main() {
	defer log.Flush()
	flag.Parse()

	if *protoPackage == "" {
		*protoPackage = filepath.Base(*generatePath)
	}

	proFile = path.Join(*generatePath, fmt.Sprintf("%s.%s", *protoPackage, proFile))

	log.Info("parsing files for go: ", *sourceFile)

	astFile, err := parser.ParseFile(token.NewFileSet(), *sourceFile, nil, 0) // 获取文件信息
	if err != nil {
		log.Error(err)
		return
	}
	gofile := parse.NewFile(func() (string, string, string) {
		slice := make([]string, 3)
		slice[0] = *protoPackage
		genPath, _ := filepath.Abs(*generatePath)
		pkgPath := filepath.Dir(*sourceFile)
		absPath, _ := filepath.Abs(".")
		list := filepath.SplitList(os.Getenv("GOPATH"))
		for _, val := range list {
			if strings.Contains(pkgPath, fmt.Sprintf("%s%s", val, "/src/")) {
				slice[1] = pkgPath[len(val)+5:]
			}
			if strings.Contains(absPath, fmt.Sprintf("%s%s", val, "/src/")) {
				slice[1] = absPath[len(val)+5:]
			}
			if strings.Contains(genPath, fmt.Sprintf("%s%s", val, "/src/")) {
				slice[2] = genPath[len(val)+5:]
			}
		}
		return slice[0], slice[1], slice[2]
	}())
	gofile.ParseFile(astFile)
	if len(gofile.Interfaces) == 0 {
		return
	}
	gofile.ParsePkgStruct(&parse.Package{PkgPath: gofile.PkgPath})

	gofile.GoTypeConfig()

	if *genProto {
		os.MkdirAll(*generatePath, 0755)

		// generate proto file
		profile, err := createFile(proFile)
		if err != nil {
			log.Error(err)
			return
		}
		defer profile.Close()
		gofile.GenProtoFile(profile)

		// run protoc
		log.Info("run the protoc command ...")
		dir := filepath.Dir(proFile)
		out, err := exec.Command("protoc", "--proto_path="+dir+"/", "--gogofaster_out=plugins=grpc:"+dir+"/", proFile).CombinedOutput()
		if err != nil {
			log.Error("protoc error: ", string(out))
			return
		}
		log.Info("protoc complete !")
	}

	// gen mplement
	controllerPkg = path.Join("../", controllerPkg)
	os.MkdirAll(controllerPkg, 0755)

	implementFile, err := createKitFile(path.Join(controllerPkg, "implement.go"))
	if err != nil {
		log.Error(err)
		return
	}
	gofile.GenImplFile(&parse.Package{
		Name: "controllers",
		Path: path.Dir(gofile.GenPath),
	}, implementFile)

	// gen func
	for _, Interface := range gofile.Interfaces {
		for _, Func := range Interface.Funcs {
			log.Info("create file: " + Func.Name)
			kitfile, err := createKitFile(path.Join(controllerPkg, fmt.Sprintf("%s.go", Func.Name)))
			if err != nil {
				log.Error(err)
				return
			}
			gofile.GenKitFile(&Interface, &Func, kitfile)
		}
	}

	log.Info("complete!")
}

func createFile(fileName string) (*os.File, error) {
	os.RemoveAll(fileName)
	log.Info("create file: " + fileName)
	return os.Create(fileName)
}

func createKitFile(filepath string) (*os.File, error) {
	os.RemoveAll(filepath)
	return os.Create(filepath)
}

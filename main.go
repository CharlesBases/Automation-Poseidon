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

	"github.com/CharlesBases/Automation-Poseidon/parse"
)

var (
	sourceFile        = flag.String("file", ".", "full path of the interface file")
	generateProtoPath = flag.String("protoP", "./pb/", "full path of the generate rpc folder")
	generateInterPath = flag.String("interP", "../controllers/", "full path of the generate interface folder")
	protoPackage      = flag.String("package", "pb", "package name in .proto file")
	generateProto     = flag.Bool("proto", false, "generate proto file or not")
	update            = flag.Bool("update", false, "update existing interface or not")
)

var src string // $GOPATH/src

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

	// swg := sync.WaitGroup{}
	// swg.Add(4)

	protofile := path.Join(*generateProtoPath, fmt.Sprintf("%s.proto", *protoPackage))

	*sourceFile, _ = filepath.Abs(*sourceFile)
	log.Info("parsing files for go: ", *sourceFile)

	astFile, err := parser.ParseFile(token.NewFileSet(), *sourceFile, nil, 0) // 获取文件信息
	if err != nil {
		log.Error(err)
		return
	}

	gofile := new(parse.File)
	// go func() {
	// 	defer swg.Done()

	src = func() string {
		list := filepath.SplitList(os.Getenv("GOPATH"))
		for _, gopath := range list {
			return filepath.Join(gopath, "src")
		}
		return "src"
	}()

	*gofile = parse.File{
		Name:         filepath.Base(*sourceFile),
		PackagePath:  strings.TrimPrefix(filepath.Dir(*sourceFile), src)[1:],
		ProtoPackage: *protoPackage,
		GenProtoPath: func() string {
			abspath, err := filepath.Abs(*generateProtoPath)
			if err != nil {
				log.Error("parse generate proto path error: ", err)
				os.Exit(1)
			}
			if *generateProto {
				os.MkdirAll(abspath, 0755)
			}
			return strings.TrimPrefix(abspath, src)[1:]
		}(),
		GenInterPath: func() string {
			abspath, err := filepath.Abs(*generateInterPath)
			if err != nil {
				log.Error("parse generate interface path error: ", err)
				os.Exit(1)
			}
			os.MkdirAll(abspath, 0755)
			return strings.TrimPrefix(abspath, src)[1:]
		}(),
	}
	// }()

	gofile.ParseFile(astFile)
	if len(gofile.Interfaces) == 0 {
		log.Error("no interface found")
		return
	}
	// 解析源文件包下结构体
	gofile.ParsePkgStruct(&parse.Package{PackagePath: gofile.PackagePath})

	gofile.GoTypeConfig()

	if *generateProto {
		// generate proto file
		profile, err := createFile(protofile)
		if err != nil {
			log.Error(err)
			return
		}
		defer profile.Close()
		gofile.GenProtoFile(profile)

		// run protoc
		log.Info("run the protoc command ...")
		dir := filepath.Dir(protofile)
		out, err := exec.Command("protoc", "--proto_path="+dir+"/", "--gogofaster_out=plugins=grpc:"+dir+"/", protofile).CombinedOutput()
		if err != nil {
			log.Error("protoc error: ", string(out))
			return
		}
		log.Info("protoc complete !")
	}

	// gen implement
	implementFile, err := createFile(filepath.Join(gofile.GenInterPath, "implement.go"))
	if err != nil {
		log.Error(err)
		return
	}
	gofile.GenImplFile(implementFile)

	// gen func
	for _, Interface := range gofile.Interfaces {
		for _, Func := range Interface.Funcs {
			if !*update && isexit(filepath.Join(gofile.GenInterPath, fmt.Sprintf("%s.go", Func.Name))) {
				continue
			}
			log.Info("create file: " + Func.Name)
			kitfile, err := createFile(filepath.Join(gofile.GenInterPath, fmt.Sprintf("%s.go", Func.Name)))
			if err != nil {
				log.Error(err)
				return
			}
			gofile.GenKitFile(&Interface, &Func, kitfile)
		}
	}

	log.Info("complete!")
}

func isexit(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

func createFile(fileName string) (*os.File, error) {
	log.Info("create file: " + fileName)
	fileName = filepath.Join(src, fileName)
	os.RemoveAll(fileName)
	return os.Create(fileName)
}

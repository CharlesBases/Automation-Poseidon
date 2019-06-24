package main

import (
	"bytes"
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
	"golang.org/x/tools/imports"

	"github.com/CharlesBases/proto/parse"
)

var (
	goFile       = flag.String("file", "", "full path of the file")
	generatePath = flag.String("path", "./pb/", "full path of the generate folder")
	protoPackage = flag.String("package", "auto", "package name in .proto file")
)

var (
	serFile = "server.go"
	cliFile = "client.go"
	proFile = "proto"
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

	serFile = path.Join(*generatePath, fmt.Sprintf("%s.%s", *protoPackage, serFile))
	cliFile = path.Join(*generatePath, fmt.Sprintf("%s.%s", *protoPackage, cliFile))
	proFile = path.Join(*generatePath, fmt.Sprintf("%s.%s", *protoPackage, proFile))

	os.MkdirAll(*generatePath, 0755)

	log.Info("parsing files for go: ", *goFile)

	astFile, err := parser.ParseFile(token.NewFileSet(), *goFile, nil, 0) // 获取文件信息
	if err != nil {
		log.Error(err)
		return
	}
	gofile := parse.NewFile(*protoPackage, func() string {
		list := filepath.SplitList(os.Getenv("GOPATH"))
		packagePath := filepath.Dir(*goFile)
		absPath, _ := filepath.Abs(".")
		for i := range list {
			if strings.Contains(packagePath, list[i]) {
				return packagePath[len(list[i])+5:]
			}
			if strings.Contains(absPath, list[i]) {
				return absPath[len(list[i])+5:]
			}
		}
		return ""
	}())
	gofile.ParseFile(astFile)
	if len(gofile.Interfaces) == 0 {
		return
	}
	gofile.ParsePkgStruct(&parse.Package{PkgPath: gofile.PkgPath})

	// generate proto file
	profile, err := createFile(proFile)
	if err != nil {
		log.Error(err)
		return
	}
	defer profile.Close()
	gofile.GenProtoFile(profile)

	log.Info("run the protoc command ...")
	dir := filepath.Dir(proFile)
	out, err := exec.Command("protoc", "--proto_path="+dir+"/", "--gogo_out=plugins=grpc:"+dir+"/", proFile).CombinedOutput()
	if err != nil {
		log.Error("protoc error: ", string(out))
		return
	}
	log.Info("protoc complete !")

	gofile.GoTypeConfig()

	// generate server file
	serfile, err := createFile(serFile)
	if err != nil {
		log.Error(err)
		return
	}
	defer serfile.Close()
	bufferSer := bytes.NewBuffer([]byte{})
	gofile.GenServer(bufferSer)
	serfile.Write(bufferSer.Bytes())
	byteSlice, e1 := imports.Process("", bufferSer.Bytes(), nil)
	if e1 != nil {
		log.Error(e1)
		return
	}
	serfile.Truncate(0)
	serfile.Seek(0, 0)
	serfile.Write(byteSlice)

	// generate client file
	clifile, err := createFile(cliFile)
	if err != nil {
		log.Error(err)
		return
	}
	defer clifile.Close()
	bufferCli := bytes.NewBuffer([]byte{})
	gofile.GenClient(bufferCli)
	clifile.Write(bufferCli.Bytes())
	byteSlice, e := imports.Process("", bufferCli.Bytes(), nil)
	if e != nil {
		log.Error(e)
		return
	}
	clifile.Truncate(0)
	clifile.Seek(0, 0)
	clifile.Write(byteSlice)

	log.Info("complete!")
}

func createFile(fileName string) (*os.File, error) {
	os.RemoveAll(fileName)
	log.Info("create file: " + fileName)
	file, err := os.Create(fileName)
	if err != nil {
		return file, err
	}
	return file, nil
}

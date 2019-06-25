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
	protoPackage = flag.String("package", "", "package name in .proto file")
	isCS         = flag.Bool("cs", true, "whether to generate C/S code")
)

var (
	serFile = "server.go"
	cliFile = "client.go"
	proFile = "proto"

	microServerFile = "micro/server.go"
	microClientFile = "micro/client.go"
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

	microServerFile = path.Join(*generatePath, fmt.Sprintf("%s", microServerFile))
	microClientFile = path.Join(*generatePath, fmt.Sprintf("%s", microClientFile))

	os.MkdirAll(*generatePath, 0755)

	log.Info("parsing files for go: ", *goFile)

	astFile, err := parser.ParseFile(token.NewFileSet(), *goFile, nil, 0) // 获取文件信息
	if err != nil {
		log.Error(err)
		return
	}
	gofile := parse.NewFile(func() (string, string, string) {
		slice := make([]string, 3)
		slice[0] = *protoPackage
		genPath, _ := filepath.Abs(*generatePath)
		pkgPath := filepath.Dir(*goFile)
		list := filepath.SplitList(os.Getenv("GOPATH"))
		for _, val := range list {
			if strings.Contains(pkgPath, fmt.Sprintf("%s%s", val, "/src/")) {
				slice[1] = pkgPath[len(val)+5:]
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
	out, err := exec.Command("protoc", "--proto_path="+dir+"/", "--gogofaster_out=plugins=grpc:"+dir+"/", "--micro_out="+dir+"/", proFile).CombinedOutput()
	if err != nil {
		log.Error("protoc error: ", string(out))
		return
	}
	log.Info("protoc complete !")

	gofile.GoTypeConfig()

	if !*isCS {
		log.Info("complete!")
		return
	}

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
	byteSlice, serErr := imports.Process("", bufferSer.Bytes(), nil)
	if serErr != nil {
		log.Error(serErr)
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
	byteSlice, cliErr := imports.Process("", bufferCli.Bytes(), nil)
	if cliErr != nil {
		log.Error(cliErr)
		return
	}
	clifile.Truncate(0)
	clifile.Seek(0, 0)
	clifile.Write(byteSlice)

	os.MkdirAll(fmt.Sprintf("%s%s", *generatePath, "/micro"), 0755)

	// generate micro server
	microServer, err := createFile(microServerFile)
	if err != nil {
		log.Error(err)
		return
	}
	defer microServer.Close()
	bufferMicroServer := bytes.NewBuffer([]byte{})
	gofile.GenMicroServer(bufferMicroServer)
	microServer.Write(bufferMicroServer.Bytes())
	byteSlice, microServerErr := imports.Process("", bufferMicroServer.Bytes(), nil)
	if microServerErr != nil {
		log.Error(microServerErr)
		return
	}
	microServer.Truncate(0)
	microServer.Seek(0, 0)
	microServer.Write(byteSlice)

	// generate micro client
	microClient, err := createFile(microClientFile)
	if err != nil {
		log.Error(err)
		return
	}
	defer microClient.Close()
	bufferMicroClient := bytes.NewBuffer([]byte{})
	gofile.GenMicroClient(bufferMicroClient)
	microClient.Write(bufferMicroClient.Bytes())
	byteSlice, microClientErr := imports.Process("", bufferMicroClient.Bytes(), nil)
	if microClientErr != nil {
		log.Error(microClientErr)
		return
	}
	microClient.Truncate(0)
	microClient.Seek(0, 0)
	microClient.Write(byteSlice)

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

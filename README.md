## 说明
	根据接口自动生成rpc文件
###### 备注
* C/S代码基于micro_out生成
* 使用 protoc-gen-gogofaster 插件生成pb.go文件
```
protoc-gen-gofast   (5-7 times faster than gogo)
protoc-gen-gogofast (same as gofast, but imports gogoprotobuf)
protoc-gen-gogofaster   (same as gogofast, without XXX_unrecognized, less pointer fields)
protoc-gen-gogoslick    (same as gogofaster, but with generated string, gostring and equal methods)
```

## 安装
* go get github.com/gogo/protobuf/proto
* go get github.com/micro/micro
* go get github.com/micro/protoc-gen-micro
* go install github.com/CharlesBases/proto
* brew install protobuf

## 用法
```
proto --help
Usage of proto:
  -file string
        full path of the file
  -package string
        package name in .proto file (default "auto")
  -path string
        full path of the generate folder (default "./pb/")
```
* proto -file="./bll.go" -package=auto
* //go:generate proto -file=$GOFILE -package=auto

```
[2019-06-24 14:27:58.381][INF] ==> parsing files for go: /Users/sun/go/SourceCode/src/github.com/CharlesBases/proto/test/bll.go
[2019-06-24 14:27:58.381][INF] ==> find interface: AutoService
[2019-06-24 14:27:58.381][INF] ==> parse structs in package: github.com/CharlesBases/proto/test
[2019-06-24 14:27:58.833][INF] ==> parse structs in package: github.com/CharlesBases/proto/test/model
[2019-06-24 14:27:58.833][INF] ==> find struct: Request
[2019-06-24 14:27:58.833][INF] ==> find struct: Response
[2019-06-24 14:27:58.833][INF] ==> parse structs in package: github.com/CharlesBases/proto/test/model
[2019-06-24 14:27:58.834][INF] ==> find struct: Point
[2019-06-24 14:27:58.834][INF] ==> create file: pb/auto.proto
[2019-06-24 14:27:58.834][INF] ==> generating .proto files ...
[2019-06-24 14:27:58.837][INF] ==> run the protoc command ...
[2019-06-24 14:27:58.859][INF] ==> protoc complete !
[2019-06-24 14:27:58.859][INF] ==> create file: pb/auto.server.go
[2019-06-24 14:27:58.859][INF] ==> generating server file ...
[2019-06-24 14:27:58.864][INF] ==> create file: pb/auto.client.go
[2019-06-24 14:27:58.864][INF] ==> generating client file ...
[2019-06-24 14:27:59.002][INF] ==> complete!
```

## 支持类型
* golang基础类型
* error
* interface{}
* map[string]interface{}

## 不支持类型
* byte
* map slice
* slice point
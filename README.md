## 说明
	根据接口自动生成rpc文件
###### 备注
* 包含 go-kit 框架所使用的 DecodeRequest、 EncodeResponse 和 Endpoint
* 使用 protoc-gen-gogofaster 插件生成pb.go文件
```
protoc-gen-gofast       (5-7 times faster than gogo)
protoc-gen-gogofast     (same as gofast, but imports gogoprotobuf)
protoc-gen-gogofaster   (same as gogofast, without XXX_unrecognized, less pointer fields)
protoc-gen-gogoslick    (same as gogofaster, but with generated string, gostring and equal methods)
```

## 安装
* go get github.com/golang/protobuf/{proto,protoc-gen-go}
* go get github.com/gogo/protobuf/proto
* go install github.com/gogo/protobuf/protoc-gen-gogofaster
* go install github.com/CharlesBases/proto
* brew install protobuf

## 用法
```
proto --help
Usage of proto:
  -file string
        full path of the interface file (default ".")
  -interP string
        full path of the generate interface folder (default "../controllers/")
  -package string
        package name in .proto file (default "pb")
  -proto
        generate proto file or not
  -protoP string
        full path of the generate rpc folder (default "./pb/")
  -update
        update existing interface or not

```
* proto -file="./bll.go" -package=auto
* //go:generate proto -file=$GOFILE -package=auto

```
[2019-11-17 12:49:47.502][INF] ==> parsing files for go: D:\Program\Go\src\src\github.com\CharlesBases\Automation-Poseidon\bll\bll.go
[2019-11-17 12:49:47.502][INF] ==> find interface: ProtoService
[2019-11-17 12:49:47.510][INF] ==> parse structs in package: github.com/CharlesBases/Automation-Poseidon/core
[2019-11-17 12:49:47.544][INF] ==> find struct: ProtoDemoRequest
[2019-11-17 12:49:47.577][INF] ==> find struct: Error
[2019-11-17 12:49:47.577][INF] ==> find struct: ProtoDemoResponse
[2019-11-17 12:49:47.577][INF] ==> parse structs in package: github.com/CharlesBases/Automation-Poseidon/core
[2019-11-17 12:49:47.612][INF] ==> find struct: Error
[2019-11-17 12:49:47.612][INF] ==> find struct: Util
[2019-11-17 12:49:47.612][INF] ==> parse structs in package: github.com/CharlesBases/Automation-Poseidon/core
[2019-11-17 12:49:47.646][INF] ==> find struct: A
[2019-11-17 12:49:47.646][INF] ==> create file: github.com\CharlesBases\Automation-Poseidon\controllers\implement.go
[2019-11-17 12:49:47.646][INF] ==> create file: github.com\CharlesBases\Automation-Poseidon\controllers\ProtoDemoD.go
[2019-11-17 12:49:47.646][INF] ==> create file: github.com\CharlesBases\Automation-Poseidon\controllers\ProtoDemoB.go
[2019-11-17 12:49:47.646][INF] ==> create file: github.com\CharlesBases\Automation-Poseidon\controllers\ProtoDemoC.go
[2019-11-17 12:49:47.646][INF] ==> create file: github.com\CharlesBases\Automation-Poseidon\controllers\ProtoDemoA.go
[2019-11-17 12:49:47.647][INF] ==> generating implement files ...
[2019-11-17 12:49:47.647][INF] ==> generating ProtoDemoD files ...
[2019-11-17 12:49:47.647][INF] ==> generating ProtoDemoB files ...
[2019-11-17 12:49:47.647][INF] ==> generating ProtoDemoC files ...
[2019-11-17 12:49:47.647][INF] ==> generating ProtoDemoA files ...
[2019-11-17 12:49:47.648][INF] ==> complete!

```

## 支持类型
* 匿名参数
* golang基础类型
* error
* interface{}
* map[string]interface{}

## 不支持类型
* 匿名结构体
* byte
* map slice
* slice point
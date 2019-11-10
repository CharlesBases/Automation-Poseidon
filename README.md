## 说明
	根据接口自动生成rpc文件
###### 备注
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
        full path of the file
  -package string
        package name in .proto file
  -path string
        full path of the generate folder (default "./pb/")

```
* proto -file="./bll.go" -package=auto
* //go:generate proto -file=$GOFILE -package=auto

```
[2019-06-25 17:14:50.104][INF] ==> parsing files for go: bll.go
[2019-06-25 17:14:50.105][INF] ==> find interface: AutoService
[2019-06-25 17:14:50.105][INF] ==> parse structs in package: github.com/CharlesBases/proto/test/model
[2019-06-25 17:14:50.108][INF] ==> find struct: Request
[2019-06-25 17:14:50.109][INF] ==> find struct: Response
[2019-06-25 17:14:50.109][INF] ==> parse structs in package: github.com/CharlesBases/proto/test/model
[2019-06-25 17:14:50.110][INF] ==> find struct: Point
[2019-06-25 17:14:50.110][INF] ==> parse structs in package: github.com/CharlesBases/proto/test
[2019-06-25 17:14:50.489][INF] ==> create file: pb/auto.proto
[2019-06-25 17:14:50.489][INF] ==> generating .proto files ...
[2019-06-25 17:14:50.492][INF] ==> run the protoc command ...
[2019-06-25 17:14:50.538][INF] ==> protoc complete !
[2019-06-25 17:14:50.538][INF] ==> create file: pb/auto.server.go
[2019-06-25 17:14:50.538][INF] ==> generating server file ...
[2019-06-25 17:14:50.543][INF] ==> create file: pb/auto.client.go
[2019-06-25 17:14:50.543][INF] ==> generating client file ...
[2019-06-25 17:14:50.685][INF] ==> create file: pb/micro/server.go
[2019-06-25 17:14:50.685][INF] ==> generating micro_server file ...
[2019-06-25 17:14:50.686][INF] ==> create file: pb/micro/client.go
[2019-06-25 17:14:50.686][INF] ==> generating micro_client file ...
[2019-06-25 17:14:50.687][INF] ==> complete!
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
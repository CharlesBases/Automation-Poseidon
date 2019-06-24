## 说明
	根据接口自动生成rpc文件
## 安装
	* go get github.com/golang/protobuf/{proto,protoc-gen-go}
	* go install github.com/CharlesBases/proto
	* brew install protobuf
## 用法
	* proto -file="./bll.go" -package=auto
	* //go:generate proto -file=$GOFILE -package=auto
```
[2019-06-24 14:27:58.381][INF] ==> parsing files for go: /Users/sun/go/SourceCode/src/github.com/CharlesBases/proto/A/bll.go
[2019-06-24 14:27:58.381][INF] ==> find interface: AutoService
[2019-06-24 14:27:58.381][INF] ==> parse structs in package: github.com/CharlesBases/proto/A
[2019-06-24 14:27:58.833][INF] ==> parse structs in package: github.com/CharlesBases/proto/A/model
[2019-06-24 14:27:58.833][INF] ==> find struct: Request
[2019-06-24 14:27:58.833][INF] ==> find struct: Response
[2019-06-24 14:27:58.833][INF] ==> parse structs in package: github.com/CharlesBases/proto/A/model
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
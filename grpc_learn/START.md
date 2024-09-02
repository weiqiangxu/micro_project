# GRPC

### 定义服务 Service

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

export PATH="$PATH:$(go env GOPATH)/bin"

protoc-gen-go --version
protoc-gen-go-grpc --version
```

```
/home/dofun/protoc/bin/protoc
```

```
# 调用后生成 *.pb.go

/home/dofun/protoc/bin/protoc --go_out=./ --go-grpc_out=./ proto/hello.proto
```

```
--go_out用于指定生成源码的保存路径

--go_out=.会针对hello.proto文件里的message生成相关代码

--go-grpc_out=.会针对hello.proto文件里的service生成相关代码
```

```
go run main.go -name xuweiqiang
```

```
# 生成目录与service同级别
/home/dofun/protoc/bin/protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. ./service/logic.proto
```
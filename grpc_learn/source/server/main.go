package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "grpc_learn/source/hello"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// 实现接口 Greeter
type server struct {
	pb.UnimplementedGreeterServer
}

// Greeter.SayHello 具体的实现
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	flag.Parse()

	// 创建一个TCP监听
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 启动 GRPC 服务
	s := grpc.NewServer()

	// 注册 Greeter.SayHello 的具体实现
	pb.RegisterGreeterServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	// 监听每一个inconming request
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "grpc_learn/source/hello"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")

	// go run main.go -name xuweiqiang
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {

	flag.Parse()

	// 设置与服务器的连接
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// 建立 Client
	c := pb.NewGreeterClient(conn)

	// 连接服务器并打印出它的响应
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 调用 client.SayHello
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}

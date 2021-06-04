package main

import (
	"context"
	"io"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/tencentyun/tsf-go/log"
	"github.com/tencentyun/tsf-go/pkg/grpc/server"
	"github.com/tencentyun/tsf-go/pkg/util"
	pb "github.com/tencentyun/tsf-go/testdata"

	"google.golang.org/grpc"
)

func main() {
	util.ParseFlag()

	server := server.NewServer(&server.Config{ServerName: "provider-demo"})
	pb.RegisterGreeterServer(server.GrpcServer(), &Service{})
	server.Use(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		log.DefaultLog.Infow("msg", "enter grpc handler!", "method", info.FullMethod, "dur", time.Since(start))
		return
	})

	err := server.Start()
	if err != nil {
		panic(err)
	}
}

// Service is gRPC service
type Service struct {
}

// SayHello is service method of SayHello
func (s *Service) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	if req.Name == "Error" {
		// 返回400 badrequest
		return &pb.HelloReply{Message: "hi " + req.Name}, errors.BadRequest(errors.UnknownReason, req.Name)
	}
	return &pb.HelloReply{Message: "hi " + req.Name}, nil
}

// SayHello is service method of SayHelloStream
func (s *Service) SayHelloStream(stream pb.Greeter_SayHelloStreamServer) error {
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		err = stream.Send(&pb.HelloReply{Message: "welcome :" + r.Name})
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

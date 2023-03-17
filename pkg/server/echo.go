package server

import (
	"context"
	"io"
	"strings"

	pb "github.com/frauniki/echo-server/gen/proto/echo"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type EchoService struct {
	pb.UnimplementedEchoServiceServer
}

func (s *EchoService) Get(ctx context.Context, _ *pb.Empty) (*pb.Response, error) {
	return genResponseFromContext(ctx, "Hello!"), nil
}

func (s *EchoService) GetRoute1(ctx context.Context, _ *pb.Empty) (*pb.Response, error) {
	return genResponseFromContext(ctx, "Hello!"), nil
}

func (s *EchoService) GetRoute2(ctx context.Context, _ *pb.Empty) (*pb.Response, error) {
	return genResponseFromContext(ctx, "Hello!"), nil
}

func (s *EchoService) Stream(ss pb.EchoService_StreamServer) error {
	for {
		if _, err := ss.Recv(); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := ss.Send(genResponseFromContext(ss.Context(), "Hello Stream!")); err != nil {
			return err
		}
	}

	return nil
}

func newEchoService() pb.EchoServiceServer {
	return &EchoService{}
}

func genResponseFromContext(ctx context.Context, msg string) *pb.Response {
	resp := &pb.Response{Message: msg}

	if pr, ok := peer.FromContext(ctx); ok {
		resp.ClientAddress = pr.Addr.String()
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		resp.Metadata = make(map[string]string)
		for k, v := range md {
			resp.Metadata[k] = strings.Join(v, ",")
		}
	}

	return resp
}

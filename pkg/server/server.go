package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	pb "github.com/frauniki/echo-server/gen/proto/echo"
	"github.com/frauniki/echo-server/pkg/logger"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	network = "tcp"
)

var (
	log *logrus.Logger
)

func init() {
	log = logger.NewLogger()
}

func GRPCServerRun(ctx context.Context, addr string) error {
	l, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	entry := logrus.NewEntry(log)
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		grpc_logrus.UnaryServerInterceptor(entry),
	))
	grpc_logrus.ReplaceGrpcLogger(entry)

	pb.RegisterEchoServiceServer(s, newEchoService())

	log.Infof("GRPC Server Listen on %s", addr)

	go func() {
		defer func() {
			if err := l.Close(); err != nil {
				log.Errorf("failed to grpc close %s: %v", addr, err)
			}
		}()
		defer s.GracefulStop()
		<-ctx.Done()
	}()

	return s.Serve(l)
}

func HTTPServerRun(ctx context.Context, addr string, grpcEndpoint string) error {
	mux := runtime.NewServeMux(runtime.WithMetadata(func(_ context.Context, req *http.Request) metadata.MD {
		return metadata.New(map[string]string{
			"grpcgateway-http-path": req.URL.Path,
		})
	}))

	if err := pb.RegisterEchoServiceHandlerServer(ctx, mux, newEchoService()); err != nil {
		return err
	}

	if err := pb.RegisterEchoServiceHandlerFromEndpoint(
		ctx,
		mux,
		grpcEndpoint,
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	); err != nil {
		return err
	}

	log.Infof("HTTP Server Listen on %s", addr)

	server := &http.Server{Addr: addr, Handler: mux}
	go func() {
		defer func() {
			if err := server.Shutdown(ctx); err != nil {
				log.Errorf("failed to close http server %s: %v", addr, err)
			}
		}()
		<-ctx.Done()
	}()

	return server.ListenAndServe()
}

func Run(host string, grpcPort int, httpPort int) error {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer cancel()
		<-sigCh
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	grpcAddr := fmt.Sprintf("%s:%d", host, grpcPort)
	httpAddr := fmt.Sprintf("%s:%d", host, httpPort)

	go func() {
		if err := GRPCServerRun(ctx, grpcAddr); err != nil {
			log.Fatal(err.Error())
		}
	}()
	go func() {
		if err := HTTPServerRun(ctx, httpAddr, grpcAddr); err != nil {
			log.Fatal(err.Error())
		}
	}()

	wg.Wait()

	return nil
}

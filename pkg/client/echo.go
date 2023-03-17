package client

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "github.com/frauniki/echo-server/gen/proto/echo"
	"github.com/frauniki/echo-server/pkg/logger"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	log *logrus.Logger
)

func init() {
	log = logger.NewLogger()
}

func sendEcho(ctx context.Context, client pb.EchoServiceClient) error {
	resp, err := client.Get(ctx, &pb.Empty{})
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	log.Infof("response=%s", string(bytes))

	return nil
}

func Run(address string, stream bool, loop bool, tick time.Duration) error {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer cancel()
		<-sigCh
	}()

	var wg sync.WaitGroup

	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil
	}
	defer conn.Close()

	client := pb.NewEchoServiceClient(conn)

	if stream {
		var wg sync.WaitGroup
		wg.Add(1)

		// Reciver
		go func() {
			defer wg.Done()

			for range time.NewTicker(tick).C {
				select {
				case <-ctx.Done():
					return
				default:
					s, err := client.Stream(ctx)
					if err != nil {
						log.Error(err.Error())
						continue
					}

					if err := s.Send(&pb.Empty{}); err != nil && err != io.EOF {
						log.Error(err.Error())
					}
					log.Debug("send stream echo")

					resp, err := s.Recv()
					if err == io.EOF {
						return
					} else if err != nil {
						log.Error(err.Error())
					}

					bytes, err := json.Marshal(resp)
					if err != nil {
						log.Error(err.Error())
					}
					log.Infof("response=%s", string(bytes))
				}
			}
		}()

		wg.Wait()
	} else {
		if loop {
			wg.Add(1)
			go func() {
				for range time.NewTicker(tick).C {
					select {
					case <-ctx.Done():
						wg.Done()
						break
					default:
						if err := sendEcho(ctx, client); err != nil {
							log.Error(err.Error())
						}
					}
				}
			}()
			wg.Wait()
		} else {
			if err := sendEcho(ctx, client); err != nil {
				return err
			}
		}
	}

	return nil
}

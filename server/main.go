package main

import (
	"context"
	"net"
	"strconv"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/encoding/gzip" // Install the gzip compressor
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"examples/helloworld/proto/productservice"
)

var (
	ports = []string{
		"localhost:30001",
		"localhost:30002",
	}
)

type server struct {
	productservice.UnimplementedProductServer
}

func (s *server) Import(ctx context.Context, request *productservice.ImportRequest) (*productservice.ImportResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "UnaryEcho: failed to get metadata")
	}
	if t, ok := md["timestamp"]; ok {
		timestampInt, _ := strconv.ParseInt(t[0], 10, 64)
		requestedAt := time.Unix(timestampInt, 0)
		log.Info("latecy", zap.Int64("time_ms",time.Now().Sub(requestedAt).Milliseconds()))
	}
	return &productservice.ImportResponse{Message: "Hello " + request.GetAppID()}, nil
}

var log *zap.Logger

func main() {
	log, _ = zap.NewProduction()
	defer log.Sync()

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_zap.StreamServerInterceptor(log),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(log),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
	productservice.RegisterProductServer(s, &server{})

	var lis net.Listener
	var err error
	for index, port := range ports {
		lis, err = net.Listen("tcp", port)
		if err == nil {
			break
		}
		if index == len(ports) - 1 {
			log.Fatal("no free port to use")
		}
	}
	log.Info("start", zap.String("port", lis.Addr().String()))
	if err := s.Serve(lis); err != nil {
		log.Fatal("error", zap.Error(err))
	}
}

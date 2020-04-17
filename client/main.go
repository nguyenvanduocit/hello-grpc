package main

import (
	"context"
	"strconv"
	"time"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"

	"examples/helloworld/proto/productservice"
)
var log *zap.Logger
func main() {
	log, _ = zap.NewProduction()
	defer log.Sync()

	ctx := context.Background()
	response, err := importProduct(ctx)
	if err != nil {
		log.Fatal("did not connect", zap.Error(err))
	}
	log.Info(response.GetMessage())
}

func importProduct (ctx context.Context) (*productservice.ImportResponse, error) {
	// Make another ClientConn with round_robin policy.
	conn, err := grpc.Dial(
		Scheme + ":///product.core",
		grpc.WithBalancerName("round_robin"), // This sets the initial balancing policy.
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(grpc_zap.UnaryClientInterceptor(log)),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer conn.Close()

	client := productservice.NewProductClient(conn)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	md := metadata.Pairs("timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := client.Import(ctx, &productservice.ImportRequest{
		AppID: "test-app",
		Shop: &productservice.Shop{
			ShopID: 111,
			AccessToken: "abc",
			MyshopifyDomain: "abc",
		},
		Database: &productservice.Database{
			URI:               "abc",
			Database:          "abc",
			ProductCollection: "abc",
		},
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return response, nil
}

func init() {
	resolver.Register(&serviceResolverBuilder{})
}

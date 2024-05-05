package internal

import (
	"context"
	"flag"
	"time"

	pb "github.com/akolpakov-somehash/crispy-spoon/proto/catalog/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

type CatalogClient struct {
	conn *grpc.ClientConn
	c    pb.ProductInfoClient
}

func NewCatalogClient() (*CatalogClient, error) {
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}
	c := pb.NewProductInfoClient(conn)
	return &CatalogClient{conn, c}, nil
}

func (c *CatalogClient) GetProductList() (*pb.ProductList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return c.c.GetProductList(ctx, &pb.Empty{})
}

func (c *CatalogClient) GetProductInfo(id uint64) (*pb.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return c.c.GetProductInfo(ctx, &pb.ProductId{Id: id})
}

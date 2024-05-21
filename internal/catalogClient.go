package internal

import (
	"context"
	"os"
	"time"

	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/catalog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CatalogClient struct {
	conn *grpc.ClientConn
	c    pb.ProductInfoClient
}

type CatalogClientInterface interface {
	GetProductList() (*pb.ProductList, error)
	GetProductInfo(id uint64) (*pb.Product, error)
}

func NewCatalogClient() (*CatalogClient, error) {
	addr := os.Getenv("CATALOG_GRPC_SERVER")
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

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

package internal

import (
	"context"

	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/catalog"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockCatalogClient is a mock implementation of the CatalogClient
type MockCatalogClient struct {
	mock.Mock
}

// NewMockCatalogClient returns a new instance of MockCatalogClient
func NewMockCatalogClient() *MockCatalogClient {
	return &MockCatalogClient{}
}

// GetProductList simulates fetching the product list
func (m *MockCatalogClient) GetProductList() (*pb.ProductList, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*pb.ProductList), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetProductInfo simulates fetching product information by product ID
func (m *MockCatalogClient) GetProductInfo(id uint64) (*pb.Product, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*pb.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

// Mock implementation of the ProductInfoClient
type MockProductInfoClient struct {
	mock.Mock
}

func (m *MockProductInfoClient) GetProductList(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ProductList, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ProductList), args.Error(1)
}

func (m *MockProductInfoClient) GetProductInfo(ctx context.Context, in *pb.ProductId, opts ...grpc.CallOption) (*pb.Product, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.Product), args.Error(1)
}

func (m *MockProductInfoClient) AddProduct(ctx context.Context, in *pb.Product, opts ...grpc.CallOption) (*pb.ProductId, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ProductId), args.Error(1)
}

func (m *MockProductInfoClient) UpdateProduct(ctx context.Context, in *pb.Product, opts ...grpc.CallOption) (*pb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.Empty), args.Error(1)
}

func (m *MockProductInfoClient) DeleteProduct(ctx context.Context, in *pb.ProductId, opts ...grpc.CallOption) (*pb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.Empty), args.Error(1)
}

package internal

import (
	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/catalog"
	"github.com/stretchr/testify/mock"
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

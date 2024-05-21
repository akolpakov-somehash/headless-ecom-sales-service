package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/catalog"
)

func TestCatalogClient_GetProductList(t *testing.T) {
	mockClient := &MockProductInfoClient{}
	catalogClient := &CatalogClient{
		conn: nil,
		c:    mockClient,
	}

	expectedProductList := &pb.ProductList{
		Products: map[uint64]*pb.Product{
			0: {Id: 1, Name: "Product1", Price: 100.0},
			1: {Id: 2, Name: "Product2", Price: 200.0},
		},
	}

	mockClient.On("GetProductList", mock.Anything, &pb.Empty{}).Return(expectedProductList, nil)

	productList, err := catalogClient.GetProductList()
	assert.Nil(t, err)
	assert.Equal(t, expectedProductList, productList)
	mockClient.AssertExpectations(t)
}

func TestCatalogClient_GetProductInfo(t *testing.T) {
	mockClient := &MockProductInfoClient{}
	catalogClient := &CatalogClient{
		conn: nil,
		c:    mockClient,
	}

	expectedProduct := &pb.Product{Id: 1, Name: "Product1", Price: 100.0}
	mockClient.On("GetProductInfo", mock.Anything, &pb.ProductId{Id: 1}).Return(expectedProduct, nil)

	product, err := catalogClient.GetProductInfo(1)
	assert.Nil(t, err)
	assert.Equal(t, expectedProduct, product)
	mockClient.AssertExpectations(t)
}

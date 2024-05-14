package internal

import (
	"context"
	"sync"
	"testing"

	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/sale"
	"github.com/stretchr/testify/assert"
)

func TestNewQuoteServer(t *testing.T) {
	quoteServer, quoteStorage := NewQuoteServer()
	assert.NotNil(t, quoteServer)
	assert.NotNil(t, quoteStorage)
}

func TestQuoteStorageImpl_GetQuote(t *testing.T) {
	tests := []struct {
		name       string
		customerId int32
		quotes     map[int32]*Quote
	}{
		{"New customer", 1, make(map[int32]*Quote)},
		{"Existing customer", 1, map[int32]*Quote{1: {CustomerId: 1, Items: make(map[int32]*QuoteItem)}}},
	}

	for _, test := range tests {
		quoteStorage := &QuoteStorageImpl{
			quotes:    test.quotes,
			qouteLock: sync.RWMutex{},
		}
		t.Run(test.name, func(t *testing.T) {
			quote := quoteStorage.GetQuote(test.customerId)
			assert.Equal(t, test.customerId, quote.CustomerId)
		})
	}
}

func TestQuoteStorageImpl_AddProduct(t *testing.T) {
	tests := []struct {
		name          string
		customerId    int32
		productId     int32
		quantity      int32
		expectedQty   int32
		expectedItems int
	}{
		{"Add new product", 1, 101, 2, 2, 1},
		{"Increase quantity", 1, 101, 3, 5, 1},
		{"Add another product", 1, 102, 1, 1, 2},
	}

	quoteStorage := &QuoteStorageImpl{
		quotes:    make(map[int32]*Quote),
		qouteLock: sync.RWMutex{},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quoteStorage.AddProduct(test.customerId, test.productId, test.quantity)
			quote := quoteStorage.GetQuote(test.customerId)
			assert.Equal(t, test.expectedItems, len(quote.Items))
			assert.Equal(t, test.expectedQty, quote.Items[test.productId].Quantity)
		})
	}
}

func TestQuoteStorageImpl_RemoveProduct(t *testing.T) {
	tests := []struct {
		name          string
		customerId    int32
		productId     int32
		expectedItems int
		expectError   bool
	}{
		{"Remove existing product", 1, 101, 0, false},
		{"Remove non-existing product", 1, 102, 1, false},
		{"Remove from non-existing quote", 2, 101, 0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quoteStorage := &QuoteStorageImpl{
				quotes: map[int32]*Quote{
					1: {
						CustomerId: 1,
						Items: map[int32]*QuoteItem{
							101: {ProductID: 101, Quantity: 2},
						},
					},
				},
				qouteLock: sync.RWMutex{},
			}
			_, err := quoteStorage.RemoveProduct(test.customerId, test.productId)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				quote := quoteStorage.GetQuote(test.customerId)
				assert.Equal(t, test.expectedItems, len(quote.Items))
			}
		})
	}
}

func TestQuoteStorageImpl_UpdateQuantity(t *testing.T) {
	tests := []struct {
		name        string
		customerId  int32
		productId   int32
		newQuantity int32
		expectedQty int32
		expectError bool
	}{
		{"Update existing product", 1, 101, 5, 5, false},
		{"Update non-existing product", 1, 102, 3, 3, false},
		{"Update in non-existing quote", 2, 101, 1, 0, true},
	}

	for _, test := range tests {
		quoteStorage := &QuoteStorageImpl{
			quotes: map[int32]*Quote{
				1: {
					CustomerId: 1,
					Items: map[int32]*QuoteItem{
						101: {ProductID: 101, Quantity: 2},
					},
				},
			},
			qouteLock: sync.RWMutex{},
		}
		t.Run(test.name, func(t *testing.T) {
			_, err := quoteStorage.UpdateQuantity(test.customerId, test.productId, test.newQuantity)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				quote := quoteStorage.GetQuote(test.customerId)
				assert.Equal(t, test.expectedQty, quote.Items[test.productId].Quantity)
			}
		})
	}
}

func TestQuoteStorageImpl_ClearQuote(t *testing.T) {
	tests := []struct {
		name          string
		customerId    int32
		expectedItems int
		isLocked      bool
	}{
		{"Clear existing quote", 1, 0, false},
		{"Clear non-existing quote", 2, 0, false},
	}

	for _, test := range tests {
		quoteStorage := &QuoteStorageImpl{
			quotes: map[int32]*Quote{
				1: {
					CustomerId: 1,
					Items: map[int32]*QuoteItem{
						101: {ProductID: 101, Quantity: 2},
					},
				},
			},
			qouteLock: sync.RWMutex{},
		}
		t.Run(test.name, func(t *testing.T) {
			quoteStorage.ClearQuote(test.customerId, test.isLocked)
			quote := quoteStorage.GetQuote(test.customerId)
			assert.Equal(t, test.expectedItems, len(quote.Items))
		})
	}
}

func TestQuoteServer_AddProduct(t *testing.T) {

	tests := []struct {
		name        string
		customerId  int32
		productId   int32
		quantity    int32
		initStorage map[int32]*Quote
		expected    *pb.Quote
	}{
		{
			"Add product to new quote",
			1, 101, 2,
			map[int32]*Quote{},
			&pb.Quote{CustomerId: 1, Items: []*pb.QuoteItem{{ProductId: 101, Quantity: 2}}},
		},
		{
			"Add product to existing quote",
			1, 102, 3,
			map[int32]*Quote{
				1: {
					CustomerId: 1,
					Items: map[int32]*QuoteItem{
						101: {ProductID: 101, Quantity: 2},
					},
				},
			},
			&pb.Quote{CustomerId: 1, Items: []*pb.QuoteItem{{ProductId: 101, Quantity: 2}, {ProductId: 102, Quantity: 3}}},
		},
	}

	for _, test := range tests {
		quoteStorage := &QuoteStorageImpl{
			test.initStorage,
			sync.RWMutex{},
		}
		quoteServer := QuoteServer{
			qouteStorage: quoteStorage,
		}
		t.Run(test.name, func(t *testing.T) {
			req := &pb.ProductRequest{
				CustomerId: test.customerId,
				ProductId:  test.productId,
				Quantity:   test.quantity,
			}
			result, err := quoteServer.AddProduct(context.Background(), req)

			assert.Equal(t, test.expected.CustomerId, result.CustomerId)
			assert.Equal(t, len(test.expected.Items), len(result.Items))
			assert.NoError(t, err)
		})
	}
}

func TestQuoteServer_GetQuote(t *testing.T) {

	tests := []struct {
		name        string
		customerId  int32
		initStorage map[int32]*Quote
		expected    *pb.Quote
	}{
		{
			"Get existing quote",
			1,
			map[int32]*Quote{
				1: {
					CustomerId: 1,
					Items: map[int32]*QuoteItem{
						101: {ProductID: 101, Quantity: 2},
					},
				},
			},
			&pb.Quote{CustomerId: 1, Items: []*pb.QuoteItem{{ProductId: 101, Quantity: 2}}},
		},
		{
			"Get new quote",
			2,
			map[int32]*Quote{},
			&pb.Quote{CustomerId: 2, Items: []*pb.QuoteItem{}},
		},
	}

	for _, test := range tests {
		quoteStorage := &QuoteStorageImpl{
			test.initStorage,
			sync.RWMutex{},
		}
		quoteServer := QuoteServer{
			qouteStorage: quoteStorage,
		}
		t.Run(test.name, func(t *testing.T) {
			req := &pb.CustomerId{Id: test.customerId}
			resp, err := quoteServer.GetQuote(context.Background(), req)
			assert.NoError(t, err)
			assert.Equal(t, test.expected.CustomerId, resp.CustomerId)
			assert.Equal(t, len(test.expected.Items), len(resp.Items))
		})
	}
}

func TestQuoteServer_RemoveProduct(t *testing.T) {
	tests := []struct {
		name        string
		customerId  int32
		productId   int32
		initStorage map[int32]*Quote
		expected    *pb.Quote
		expectError bool
	}{
		{
			"Remove existing product",
			1, 101,
			map[int32]*Quote{
				1: {
					CustomerId: 1,
					Items: map[int32]*QuoteItem{
						101: {ProductID: 101, Quantity: 2},
					},
				},
			},
			&pb.Quote{CustomerId: 1, Items: []*pb.QuoteItem{}},
			false,
		},
		{
			"Remove non-existing product",
			1, 102,
			map[int32]*Quote{
				1: {
					CustomerId: 1,
					Items: map[int32]*QuoteItem{
						101: {ProductID: 101, Quantity: 2},
					},
				},
			},
			&pb.Quote{CustomerId: 1, Items: []*pb.QuoteItem{{ProductId: 101, Quantity: 2}}},
			false,
		},
		{
			"Remove from non-existing quote",
			2, 101,
			map[int32]*Quote{},
			nil,
			true,
		},
	}

	for _, test := range tests {
		quoteStorage := &QuoteStorageImpl{
			test.initStorage,
			sync.RWMutex{},
		}
		quoteServer := QuoteServer{
			qouteStorage: quoteStorage,
		}
		t.Run(test.name, func(t *testing.T) {
			req := &pb.ProductRequest{
				CustomerId: test.customerId,
				ProductId:  test.productId,
			}
			_, err := quoteServer.RemoveProduct(context.Background(), req)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQuoteServer_UpdateQuantity(t *testing.T) {

	tests := []struct {
		name        string
		customerId  int32
		productId   int32
		newQuantity int32
		initStorage map[int32]*Quote
		expected    *pb.Quote
		expectError bool
	}{
		{
			"Update existing product quantity",
			1, 101, 5,
			map[int32]*Quote{
				1: {
					CustomerId: 1,
					Items: map[int32]*QuoteItem{
						101: {ProductID: 101, Quantity: 5},
					},
				},
			},
			&pb.Quote{CustomerId: 1, Items: []*pb.QuoteItem{{ProductId: 101, Quantity: 5}}},
			false,
		},
		{
			"Update non-existing product",
			1, 102, 3,
			map[int32]*Quote{
				1: {
					CustomerId: 1,
					Items: map[int32]*QuoteItem{
						101: {ProductID: 101, Quantity: 2},
					},
				},
			},
			&pb.Quote{CustomerId: 1, Items: []*pb.QuoteItem{{ProductId: 101, Quantity: 2}}},
			false,
		},
		{
			"Update in non-existing quote",
			2, 101, 1,
			map[int32]*Quote{},
			nil,
			true,
		},
	}

	for _, test := range tests {
		quoteStorage := &QuoteStorageImpl{
			test.initStorage,
			sync.RWMutex{},
		}
		quoteServer := QuoteServer{
			qouteStorage: quoteStorage,
		}
		t.Run(test.name, func(t *testing.T) {
			req := &pb.ProductRequest{
				CustomerId: test.customerId,
				ProductId:  test.productId,
				Quantity:   test.newQuantity,
			}
			_, err := quoteServer.UpdateQuantity(context.Background(), req)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

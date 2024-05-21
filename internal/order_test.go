package internal

import (
	"context"
	"fmt"
	"sync"
	"testing"

	pbc "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/catalog"
	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/sale"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewOrderServer(t *testing.T) {
	orderServer := NewOrderServer(nil, nil)
	assert.IsType(t, &OrderServer{}, orderServer)
}

func TestOrderToProto(t *testing.T) {
	order := &Order{
		ID:         1,
		CustomerId: 1,
		Items: map[int32]*OrderItem{
			1: {
				ProductID: 1,
				Quantity:  1,
				Price:     100.0,
			},
		},
	}

	protoOrder := orderToProto(order)
	assert.Equal(t, order.ID, protoOrder.Id)
	assert.Equal(t, order.CustomerId, protoOrder.CustomerId)
	assert.Len(t, protoOrder.Items, 1)
	assert.Equal(t, order.Items[1].ProductID, protoOrder.Items[0].ProductId)
	assert.Equal(t, order.Items[1].Quantity, protoOrder.Items[0].Quantity)
	assert.Equal(t, order.Items[1].Price, protoOrder.Items[0].Price)
}

func TestOrderServer_GetOrders(t *testing.T) {
	tests := []struct {
		name             string
		orders           map[int32]map[int32]*Order
		customerOrderMap map[int32]int32
		customerId       int32
		wantErr          bool
		err              error
	}{
		{
			name: "success",
			orders: map[int32]map[int32]*Order{
				1: {
					1: {
						ID:         1,
						CustomerId: 1,
						Items: map[int32]*OrderItem{
							1: {
								ProductID: 1,
								Quantity:  1,
								Price:     100.0,
							},
						},
					},
				},
			},
			customerOrderMap: map[int32]int32{
				1: 1,
			},
			customerId: 1,
			wantErr:    false,
			err:        nil,
		},
		{
			name:             "failure",
			orders:           nil,
			customerOrderMap: nil,
			customerId:       1,
			wantErr:          true,
			err:              fmt.Errorf("no orders found for customer %d", 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderServer := &OrderServer{
				orders:           tt.orders,
				customerOrderMap: tt.customerOrderMap,
				orderLock:        sync.RWMutex{},
				quoteStorage:     nil,
				catalogClient:    nil,
			}
			got, err := orderServer.GetOrders(context.Background(), &pb.CustomerId{Id: tt.customerId})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Equal(t, tt.err.Error(), err.Error())
				return
			}
			assert.Equal(t, len(tt.orders), len(got.Orders))
		})
	}
}

func TestOrderServer_GetOrder(t *testing.T) {
	tests := []struct {
		name             string
		orders           map[int32]map[int32]*Order
		customerOrderMap map[int32]int32
		customerId       int32
		orderId          int32
		wantErr          bool
		err              error
	}{
		{
			name: "success",
			orders: map[int32]map[int32]*Order{
				1: {
					1: {
						ID:         1,
						CustomerId: 1,
						Items: map[int32]*OrderItem{
							1: {
								ProductID: 1,
								Quantity:  1,
								Price:     100.0,
							},
						},
					},
				},
			},
			customerOrderMap: map[int32]int32{
				1: 1,
			},
			customerId: 1,
			orderId:    1,
			wantErr:    false,
			err:        nil,
		},
		{
			name:             "failure",
			orders:           nil,
			customerOrderMap: nil,
			customerId:       1,
			orderId:          1,
			wantErr:          true,
			err:              fmt.Errorf("order with id %d not found", 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderServer := &OrderServer{
				orders:           tt.orders,
				customerOrderMap: tt.customerOrderMap,
				orderLock:        sync.RWMutex{},
				quoteStorage:     nil,
				catalogClient:    nil,
			}
			got, err := orderServer.GetOrder(context.Background(), &pb.OrderId{Id: tt.orderId})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Equal(t, tt.err.Error(), err.Error())
				return
			}
			assert.Equal(t, tt.orders[tt.customerId][tt.orderId].ID, got.Id)
		})
	}
}

func TestOrderServer_PlaceOrder(t *testing.T) {
	tests := []struct {
		name       string
		quote      *Quote
		customerId int32
		wantErr    bool
		err        error
	}{
		{
			name: "success",
			quote: &Quote{
				CustomerId: 1,
				Items: map[int32]*QuoteItem{
					1: {
						ProductID: 1,
						Quantity:  1,
					},
				},
			},
			customerId: 1,
			wantErr:    false,
			err:        nil,
		},
		{
			name:       "empty quote",
			quote:      &Quote{CustomerId: 1, Items: map[int32]*QuoteItem{}},
			customerId: 1,
			wantErr:    true,
			err:        fmt.Errorf("quote is empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCatalogClient := &MockCatalogClient{}
			mockCatalogClient.On("GetProductInfo", uint64(1)).Return(&pbc.Product{Id: 1, Price: 100}, nil)

			orderServer := &OrderServer{
				orders:           map[int32]map[int32]*Order{},
				customerOrderMap: nil,
				orderLock:        sync.RWMutex{},
				quoteStorage: &QuoteStorage{
					quotes: map[int32]*Quote{
						tt.customerId: tt.quote,
					},
				},
				catalogClient: mockCatalogClient,
			}
			stream := &MockOrderService_PlaceOrderServer{}
			stream.On("Send", mock.Anything).Return(nil)
			err := orderServer.PlaceOrder(&pb.CustomerId{Id: tt.customerId}, stream)
			if (err != nil) != tt.wantErr {
				t.Errorf("PlaceOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				assert.Equal(t, tt.err.Error(), err.Error())
				return
			}
			assert.Equal(t, 1, len(orderServer.orders[tt.customerId]))
		})
	}
}

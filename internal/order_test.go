package internal

import (
	"context"
	"fmt"
	"sync"
	"testing"

	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/sale"
	"github.com/stretchr/testify/assert"
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

func TestGetOrders(t *testing.T) {
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
			orderService := &OrderServer{
				orders:           tt.orders,
				customerOrderMap: tt.customerOrderMap,
				orderLock:        sync.RWMutex{},
				quoteStorage:     nil,
				catalogClient:    nil,
			}
			got, err := orderService.GetOrders(context.Background(), &pb.CustomerId{Id: tt.customerId})
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

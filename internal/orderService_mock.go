package internal

import (
	"context"

	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/sale"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

// MockOrderService_PlaceOrderServer is a mock implementation of the pb.OrderService_PlaceOrderServer interface
type MockOrderService_PlaceOrderServer struct {
	mock.Mock
}

// Send is a mock implementation of the Send method
func (m *MockOrderService_PlaceOrderServer) Send(status *pb.ProcessStatus) error {
	args := m.Called(status)
	return args.Error(0)
}

// Recv is a mock implementation of the Recv method
func (m *MockOrderService_PlaceOrderServer) Recv() (*pb.Order, error) {
	args := m.Called()
	return args.Get(0).(*pb.Order), args.Error(1)
}

// SetHeader is a mock implementation of the SetHeader method
func (m *MockOrderService_PlaceOrderServer) SetHeader(md metadata.MD) error {
	args := m.Called(md)
	return args.Error(0)
}

// SendHeader is a mock implementation of the SendHeader method
func (m *MockOrderService_PlaceOrderServer) SendHeader(md metadata.MD) error {
	args := m.Called(md)
	return args.Error(0)
}

// SetTrailer is a mock implementation of the SetTrailer method
func (m *MockOrderService_PlaceOrderServer) SetTrailer(md metadata.MD) {
	m.Called(md)
}

// Context is a mock implementation of the Context method
func (m *MockOrderService_PlaceOrderServer) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

// RecvMsg is a mock implementation of the RecvMsg method
func (m *MockOrderService_PlaceOrderServer) RecvMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

// SendMsg is a mock implementation of the SendMsg method
func (m *MockOrderService_PlaceOrderServer) SendMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

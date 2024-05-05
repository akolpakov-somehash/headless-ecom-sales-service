package internal

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/akolpakov-somehash/go-microservices/proto/sale/order"
)

type OrderItem struct {
	ProductID int32
	Quantity  int32
	Price     float32
}

type Order struct {
	ID         int32
	Items      map[int32]*OrderItem
	CustomerId int32
}

type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	orders           map[int32]map[int32]*Order
	customerOrderMap map[int32]int32
	orderLock        sync.RWMutex
	quoteStorage     QuoteStorage
	catalogClient    CatalogClient
}

func NewOrderServer(quoteStorage QuoteStorage) *OrderServer {
	catalogClient, err := NewCatalogClient()
	if err != nil {
		panic(fmt.Sprintf("failed to create catalog client: %v", err))
	}
	return &OrderServer{
		orders:           make(map[int32]map[int32]*Order),
		customerOrderMap: make(map[int32]int32),
		orderLock:        sync.RWMutex{},
		quoteStorage:     quoteStorage,
		catalogClient:    *catalogClient,
	}
}

func orderToProto(order *Order, protoOrder *pb.Order) *pb.Order {
	protoOrder.CustomerId = order.CustomerId
	protoOrder.Items = make([]*pb.OrderItem, 0)

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, item := range order.Items {
		wg.Add(1)
		go func(item *OrderItem) {
			defer wg.Done()

			mu.Lock()
			protoOrder.Items = append(protoOrder.Items, &pb.OrderItem{
				ProductId: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price,
			})
			mu.Unlock()
		}(item)
	}
	wg.Wait()

	return protoOrder
}

func (s *OrderServer) lockOrderRead() {
	s.orderLock.RLock()
}

func (s *OrderServer) unlockOrderRead() {
	s.orderLock.RUnlock()
}

func (s *OrderServer) lockOrderWrite() {
	s.orderLock.Lock()
}

func (s *OrderServer) unlockOrderWrite() {
	s.orderLock.Unlock()
}

func (s *OrderServer) GetOrders(ctx context.Context, in *pb.CustomerId) (*pb.OrderList, error) {
	s.lockOrderRead()
	defer s.unlockOrderRead()

	orders, exists := s.orders[in.Id]
	if !exists {
		return nil, fmt.Errorf("no orders found for customer %d", in.Id)
	}

	orderList := make([]*pb.Order, 0, len(orders))
	orderChan := make(chan *pb.Order)
	var wg sync.WaitGroup

	// Spawn goroutines to convert orders to protobuf format
	for _, order := range orders {
		wg.Add(1)
		go func(order *Order) {
			defer wg.Done()
			protoOrder := &pb.Order{}
			orderToProto(order, protoOrder)
			orderChan <- protoOrder // Send the proto order to the channel
		}(order)
	}

	// Close the channel once all goroutines have finished
	go func() {
		wg.Wait()
		close(orderChan)
	}()

	// Collect results from the channel
	for protoOrder := range orderChan {
		orderList = append(orderList, protoOrder)
	}

	return &pb.OrderList{Orders: orderList}, nil
}

func (s *OrderServer) GetOrder(ctx context.Context, in *pb.OrderId) (*pb.Order, error) {
	s.lockOrderRead()
	defer s.unlockOrderRead()

	customerId, exists := s.customerOrderMap[in.Id]
	if !exists {
		return nil, fmt.Errorf("order with id %d not found", in.Id)
	}
	pbOrder := &pb.Order{}
	orderToProto(s.orders[customerId][in.Id], pbOrder)
	return pbOrder, nil
}

func (s *OrderServer) PlaceOrder(in *pb.CustomerId, stream pb.OrderService_PlaceOrderServer) error {
	s.lockOrderWrite()
	defer s.unlockOrderWrite()

	s.quoteStorage.LockQuoteWrite()
	defer s.quoteStorage.UnlockQuoteWrite()

	quote := s.quoteStorage.GetQuote(in.Id)
	if len(quote.Items) == 0 {
		return sendError(stream, 0, "quote is empty")
	}

	orderItems := make(map[int32]*OrderItem, 0)
	for _, item := range quote.Items {
		product, err := s.catalogClient.GetProductInfo(uint64(item.ProductID))
		if err != nil {
			fmt.Println(err)
			return sendError(stream, 0, fmt.Sprintf("failed to get product info for product %d", item.ProductID))
		}
		orderItems[item.ProductID] = &OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		}
	}

	order := &Order{
		Items:      orderItems,
		CustomerId: in.Id,
	}

	if s.orders[in.Id] == nil {
		s.orders[in.Id] = make(map[int32]*Order)
	}
	orderId := int32(len(s.orders[in.Id]) + 1)
	order.ID = orderId
	s.orders[in.Id][orderId] = order
	s.quoteStorage.ClearQuote(in.Id, true)

	// Simulate order processing steps
	orderSteps := []pb.ProcessStatus{
		{OrderId: 1, Status: pb.OrderStatus_STARTED, Message: "Order processing started."},
		{OrderId: 1, Status: pb.OrderStatus_PROCESSED, Message: "Collecting shipping information."},
		{OrderId: 1, Status: pb.OrderStatus_PROCESSED, Message: "Collecting payment details."},
		{OrderId: 1, Status: pb.OrderStatus_COMPLETED, Message: "Order has been completed."},
	}

	// Stream each step back to client
	for i := range orderSteps {
		// Simulating delay between steps
		time.Sleep(2 * time.Second)
		fmt.Printf("Sending order process status: %s\n", orderSteps[i].Message)
		if err := stream.Send(&orderSteps[i]); err != nil {
			return fmt.Errorf("failed to send order process status: %v", err)
		}
	}

	return nil
}

func sendError(stream pb.OrderService_PlaceOrderServer, orderId int32, message string) error {
	stream.Send(&pb.ProcessStatus{
		OrderId: orderId,
		Status:  pb.OrderStatus_ERROR,
		Message: message,
	})
	return fmt.Errorf(message)
}

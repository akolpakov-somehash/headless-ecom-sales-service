package internal

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/akolpakov-somehash/go-microservices/proto/sale/quote"
)

type QuoteItem struct {
	ProductID int32
	Quantity  int32
}

type Quote struct {
	Items      map[int32]*QuoteItem
	CustomerId int32
}

type QuoteServer struct {
	pb.UnimplementedQuoteServiceServer
	qouteStorage QuoteStorage
}

func NewQuoteServer() (*QuoteServer, QuoteStorage) {
	quoteStorage := &QuoteStorageImpl{
		make(map[int32]*Quote),
		sync.RWMutex{},
	}

	return &QuoteServer{
		qouteStorage: quoteStorage,
	}, quoteStorage
}

type QuoteStorage interface {
	GetQuote(int32) *Quote
	AddProduct(customerId int32, productId int32, quantity int32) (*Quote, error)
	RemoveProduct(customerId int32, productId int32) (*Quote, error)
	UpdateQuantity(customerId int32, productId int32, quantity int32) (*Quote, error)
	ClearQuote(customerId int32, isLocked bool)
	LockQuoteRead()
	UnlockQuoteRead()
	LockQuoteWrite()
	UnlockQuoteWrite()
}

type QuoteStorageImpl struct {
	quotes    map[int32]*Quote
	qouteLock sync.RWMutex
}

/**
 * QuoteStorageImpl
 */

func (s *QuoteStorageImpl) GetQuote(customerId int32) *Quote {
	quote, exists := s.quotes[customerId]
	if !exists {
		quote = &Quote{
			CustomerId: customerId,
			Items:      make(map[int32]*QuoteItem),
		}
		s.quotes[customerId] = quote
	}
	return quote
}

func (s *QuoteStorageImpl) ClearQuote(customerId int32, isLocked bool) {
	if !isLocked {
		s.LockQuoteWrite()
		defer s.UnlockQuoteWrite()
	}

	delete(s.quotes, customerId)
}

func (s *QuoteStorageImpl) LockQuoteRead() {
	s.qouteLock.RLock()
}

func (s *QuoteStorageImpl) UnlockQuoteRead() {
	s.qouteLock.RUnlock()
}

func (s *QuoteStorageImpl) LockQuoteWrite() {
	s.qouteLock.Lock()
}

func (s *QuoteStorageImpl) UnlockQuoteWrite() {
	s.qouteLock.Unlock()
}

func (s *QuoteStorageImpl) AddProduct(customerId int32, productId int32, quantity int32) (*Quote, error) {
	s.LockQuoteWrite()
	defer s.UnlockQuoteWrite()

	quote := s.GetQuote(customerId)
	item, exexists := quote.Items[productId]
	if exexists {
		item.Quantity += quantity
	} else {
		quote.Items[productId] = &QuoteItem{
			ProductID: productId,
			Quantity:  quantity,
		}
	}
	return quote, nil
}

func (s *QuoteStorageImpl) RemoveProduct(customerId int32, productId int32) (*Quote, error) {
	s.LockQuoteWrite()
	defer s.UnlockQuoteWrite()

	quote, exists := s.quotes[customerId]
	if !exists {
		return nil, fmt.Errorf("quote not found")
	}
	delete(quote.Items, productId)
	return quote, nil
}

func (s *QuoteStorageImpl) UpdateQuantity(customerId int32, productId int32, quantity int32) (*Quote, error) {
	s.LockQuoteWrite()
	defer s.UnlockQuoteWrite()

	quote, exists := s.quotes[customerId]
	if !exists {
		return nil, fmt.Errorf("quote not found")
	}

	quote.Items[productId].Quantity = quantity
	return quote, nil
}

/**
 * QuoteServer
 */

func (s *QuoteServer) AddProduct(ctx context.Context, in *pb.ProductRequest) (*pb.Quote, error) {
	quote, err := s.qouteStorage.AddProduct(in.CustomerId, in.ProductId, in.Quantity)
	if err != nil {
		return nil, err
	}
	protoQuote := &pb.Quote{}
	quoteToProto(quote, protoQuote)
	return protoQuote, nil
}

func (s *QuoteServer) GetQuote(ctx context.Context, in *pb.CustomerId) (*pb.Quote, error) {

	protoQuote := &pb.Quote{}
	quoteToProto(s.qouteStorage.GetQuote(in.Id), protoQuote)
	return protoQuote, nil
}

func (s *QuoteServer) RemoveProduct(ctx context.Context, in *pb.ProductRequest) (*pb.Quote, error) {
	quote, err := s.qouteStorage.RemoveProduct(in.CustomerId, in.ProductId)
	if err != nil {
		return nil, err
	}
	protoQuote := &pb.Quote{}
	quoteToProto(quote, protoQuote)
	return protoQuote, nil
}

func (s *QuoteServer) UpdateQuantity(ctx context.Context, in *pb.ProductRequest) (*pb.Quote, error) {
	quote, err := s.qouteStorage.UpdateQuantity(in.CustomerId, in.ProductId, in.Quantity)
	if err != nil {
		return nil, err
	}
	protoQuote := &pb.Quote{}
	quoteToProto(quote, protoQuote)
	return protoQuote, nil
}

func quoteToProto(quote *Quote, protoQuote *pb.Quote) {
	protoQuote.CustomerId = quote.CustomerId
	protoQuote.Items = make([]*pb.QuoteItem, 0)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, item := range quote.Items {
		wg.Add(1)
		go func(item *QuoteItem) {
			defer wg.Done()

			mu.Lock()
			defer mu.Unlock()
			protoQuote.Items = append(protoQuote.Items, &pb.QuoteItem{ProductId: item.ProductID, Quantity: item.Quantity})
		}(item)
	}
	wg.Wait()
}

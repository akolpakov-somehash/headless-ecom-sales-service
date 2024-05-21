package internal

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/sale"
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
	qouteStorage QuoteStorageInterface
}

func NewQuoteServer() (*QuoteServer, QuoteStorageInterface) {
	quoteStorage := &QuoteStorage{
		make(map[int32]*Quote),
		sync.RWMutex{},
	}

	return &QuoteServer{
		qouteStorage: quoteStorage,
	}, quoteStorage
}

type QuoteStorageInterface interface {
	GetQuote(int32) *Quote
	AddProduct(customerId int32, productId int32, quantity int32) *Quote
	RemoveProduct(customerId int32, productId int32) (*Quote, error)
	UpdateQuantity(customerId int32, productId int32, quantity int32) (*Quote, error)
	ClearQuote(customerId int32, isLocked bool)
	LockQuoteRead()
	UnlockQuoteRead()
	LockQuoteWrite()
	UnlockQuoteWrite()
}

type QuoteStorage struct {
	quotes    map[int32]*Quote
	qouteLock sync.RWMutex
}

/**
 * QuoteStorageImpl
 */

func (s *QuoteStorage) GetQuote(customerId int32) *Quote {
	s.LockQuoteRead()
	quote, exists := s.quotes[customerId]
	s.UnlockQuoteRead()

	if !exists {
		s.LockQuoteWrite()
		defer s.UnlockQuoteWrite()

		quote = &Quote{
			CustomerId: customerId,
			Items:      make(map[int32]*QuoteItem),
		}
		s.quotes[customerId] = quote
	}
	return quote
}

func (s *QuoteStorage) ClearQuote(customerId int32, isLocked bool) {
	if !isLocked {
		s.LockQuoteWrite()
		defer s.UnlockQuoteWrite()
	}

	delete(s.quotes, customerId)
}

func (s *QuoteStorage) LockQuoteRead() {
	s.qouteLock.RLock()
}

func (s *QuoteStorage) UnlockQuoteRead() {
	s.qouteLock.RUnlock()
}

func (s *QuoteStorage) LockQuoteWrite() {
	s.qouteLock.Lock()
}

func (s *QuoteStorage) UnlockQuoteWrite() {
	s.qouteLock.Unlock()
}

func (s *QuoteStorage) AddProduct(customerId int32, productId int32, quantity int32) *Quote {
	quote := s.GetQuote(customerId)

	s.LockQuoteWrite()
	defer s.UnlockQuoteWrite()
	item, exexists := quote.Items[productId]
	if exexists {
		item.Quantity += quantity
	} else {
		quote.Items[productId] = &QuoteItem{
			ProductID: productId,
			Quantity:  quantity,
		}
	}
	return quote
}

func (s *QuoteStorage) RemoveProduct(customerId int32, productId int32) (*Quote, error) {
	s.LockQuoteWrite()
	defer s.UnlockQuoteWrite()

	quote, exists := s.quotes[customerId]
	if !exists {
		return nil, fmt.Errorf("quote not found")
	}
	delete(quote.Items, productId)
	return quote, nil
}

func (s *QuoteStorage) UpdateQuantity(customerId int32, productId int32, quantity int32) (*Quote, error) {
	s.LockQuoteWrite()
	defer s.UnlockQuoteWrite()

	quote, exists := s.quotes[customerId]
	if !exists {
		return nil, fmt.Errorf("quote not found")
	}

	item, exexists := quote.Items[productId]
	if exexists {
		item.Quantity = quantity
	} else {
		quote.Items[productId] = &QuoteItem{
			ProductID: productId,
			Quantity:  quantity,
		}
	}
	return quote, nil
}

/**
 * QuoteServer
 */

func (s *QuoteServer) AddProduct(ctx context.Context, in *pb.ProductRequest) (*pb.Quote, error) {
	quote := s.qouteStorage.AddProduct(in.CustomerId, in.ProductId, in.Quantity)
	protoQuote := quoteToProto(quote)
	return protoQuote, nil
}

func (s *QuoteServer) GetQuote(ctx context.Context, in *pb.CustomerId) (*pb.Quote, error) {
	protoQuote := quoteToProto(s.qouteStorage.GetQuote(in.Id))
	return protoQuote, nil
}

func (s *QuoteServer) RemoveProduct(ctx context.Context, in *pb.ProductRequest) (*pb.Quote, error) {
	quote, err := s.qouteStorage.RemoveProduct(in.CustomerId, in.ProductId)
	if err != nil {
		return nil, err
	}
	protoQuote := quoteToProto(quote)
	return protoQuote, nil
}

func (s *QuoteServer) UpdateQuantity(ctx context.Context, in *pb.ProductRequest) (*pb.Quote, error) {
	quote, err := s.qouteStorage.UpdateQuantity(in.CustomerId, in.ProductId, in.Quantity)
	if err != nil {
		return nil, err
	}
	protoQuote := quoteToProto(quote)
	return protoQuote, nil
}

func quoteToProto(quote *Quote) *pb.Quote {
	protoQuote := &pb.Quote{}
	protoQuote.CustomerId = quote.CustomerId
	protoQuote.Items = make([]*pb.QuoteItem, 0)
	for _, item := range quote.Items {
		protoQuote.Items = append(protoQuote.Items, &pb.QuoteItem{ProductId: item.ProductID, Quantity: item.Quantity})
	}
	return protoQuote
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sale/internal"

	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/sale"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50052, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	qouteServer, quoteStorage := internal.NewQuoteServer()
	pb.RegisterQuoteServiceServer(s, qouteServer)
	orderServer := internal.NewOrderServer(quoteStorage)
	pb.RegisterOrderServiceServer(s, orderServer)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

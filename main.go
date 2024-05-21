package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sale/internal"

	pb "github.com/akolpakov-somehash/headless-ecom-protos/gen/go/sale"
	"github.com/joho/godotenv"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50052, "The server port")
)

func loadEnv() error {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { //For docker run we don't have the file
		return fmt.Errorf("error loading .env file: %v", err)
	}
	return nil
}

func main() {
	err := loadEnv()
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
	}
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	qouteServer, quoteStorage := internal.NewQuoteServer()
	catalogClient, err := internal.NewCatalogClient()
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to create a new catalog client: %v", err)
	}
	pb.RegisterQuoteServiceServer(s, qouteServer)
	orderServer := internal.NewOrderServer(quoteStorage, catalogClient)
	pb.RegisterOrderServiceServer(s, orderServer)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

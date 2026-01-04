package catalog

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func startE2EServer(t *testing.T) (string, func()) {
	url := os.Getenv("ELASTICSEARCH_URL_FOR_TEST")
	if url == "" {
		url = "http://localhost:9202"
	}

	repo, err := NewElasticRepository(url)
	if err != nil {
		t.Fatalf("failed to connect to Elasticsearch: %v", err)
	}

	svc := NewSerivce(repo)

	lis, err := net.Listen("tcp", ":9092")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCatalogServiceServer(s, &grpcServer{service: svc})

	go s.Serve(lis)

	cleanup := func() {
		s.Stop()
		repo.Close()
		lis.Close()
	}

	return "localhost:9092", cleanup
}

func TestE2E_PostAndGetProduct(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewCatalogServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	postRes, err := client.PostProduct(ctx, &pb.PostProductRequest{
		Name:        "E2E Test Laptop",
		Description: "A laptop for end-to-end testing",
		Price:       1299.99,
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}

	if postRes.Product.Id == "" {
		t.Fatal("expected generated id")
	}

	getRes, err := client.GetProduct(ctx, &pb.GetProductRequest{Id: postRes.Product.Id})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if getRes.Product.Name != "E2E Test Laptop" {
		t.Errorf("unexpected name, got %s", getRes.Product.Name)
	}

	if getRes.Product.Price != 1299.99 {
		t.Errorf("unexpected price, got %f", getRes.Product.Price)
	}
}

func TestE2E_GetProducts(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewCatalogServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	_, err = client.PostProduct(ctx, &pb.PostProductRequest{
		Name:        "E2E Product Alpha",
		Description: "First product",
		Price:       10.00,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.PostProduct(ctx, &pb.PostProductRequest{
		Name:        "E2E Product Beta",
		Description: "Second product",
		Price:       20.00,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.PostProduct(ctx, &pb.PostProductRequest{
		Name:        "E2E Product Gamma",
		Description: "Third product",
		Price:       30.00,
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.GetProducts(ctx, &pb.GetProductsRequest{Skip: 0, Take: 2})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Products) < 2 {
		t.Errorf("expected at least 2 products, got %d", len(res.Products))
	}
}

func TestE2E_SearchProducts(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewCatalogServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.PostProduct(ctx, &pb.PostProductRequest{
		Name:        "Wireless Mechanical Keyboard",
		Description: "RGB backlit mechanical keyboard",
		Price:       89.99,
	})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	res, err := client.GetProducts(ctx, &pb.GetProductsRequest{
		Query: "Mechanical",
		Skip:  0,
		Take:  10,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Products) == 0 {
		t.Error("expected at least one product in search results")
	}

	found := false
	for _, p := range res.Products {
		if p.Name == "Wireless Mechanical Keyboard" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find 'Wireless Mechanical Keyboard' in search results")
	}
}

func TestE2E_GetProductsByIds(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewCatalogServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res1, err := client.PostProduct(ctx, &pb.PostProductRequest{
		Name:        "Monitor",
		Description: "4K monitor",
		Price:       399.99,
	})
	if err != nil {
		t.Fatal(err)
	}

	res2, err := client.PostProduct(ctx, &pb.PostProductRequest{
		Name:        "Mouse",
		Description: "Wireless mouse",
		Price:       29.99,
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.GetProducts(ctx, &pb.GetProductsRequest{
		Ids:  []string{res1.Product.Id, res2.Product.Id},
		Skip: 0,
		Take: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(res.Products))
	}

	foundMonitor := false
	foundMouse := false
	for _, p := range res.Products {
		if p.Name == "Monitor" {
			foundMonitor = true
		}
		if p.Name == "Mouse" {
			foundMouse = true
		}
	}

	if !foundMonitor || !foundMouse {
		t.Error("expected both Monitor and Mouse in results")
	}
}

func TestE2E_GetProductNotFound(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewCatalogServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.GetProduct(ctx, &pb.GetProductRequest{Id: "non-existent-id-12345"})
	if err == nil {
		t.Error("expected error for non-existent product, got nil")
	}
}

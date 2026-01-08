package catalog

import (
	"context"
	"net"
	"testing"

	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog/pb"
	"github.com/RathodViraj/go-microservice-graphql-grpc/inventory"
	inventorypb "github.com/RathodViraj/go-microservice-graphql-grpc/inventory/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func startFakeInventoryServerIntegration(t *testing.T) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(srv, &fakeInventoryServerIntegration{})
	go srv.Serve(lis)
	stop := func() {
		srv.Stop()
		_ = lis.Close()
	}
	return lis.Addr().String(), stop
}

type fakeInventoryServerIntegration struct {
	inventorypb.UnimplementedInventoryServiceServer
}

func (s *fakeInventoryServerIntegration) UpdateStock(ctx context.Context, r *inventorypb.UpdateStockRequest) (*inventorypb.UpdateStockResponse, error) {
	return &inventorypb.UpdateStockResponse{OutOfStock: []string{}}, nil
}

func (s *fakeInventoryServerIntegration) CheckStock(ctx context.Context, r *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	in := make([]int32, len(r.Pids))
	for i := range in {
		in[i] = 100
	}
	return &inventorypb.CheckStockResponse{InStock: in}, nil
}

func startServerIntegrationTest(t *testing.T) (*grpc.ClientConn, func()) {
	svc := NewSerivce(testRepo)

	// Setup inventory client
	invAddr, stopInv := startFakeInventoryServerIntegration(t)
	invClient, err := inventory.NewClient(invAddr)
	if err != nil {
		t.Fatalf("failed to create inventory client: %v", err)
	}

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	pb.RegisterCatalogServiceServer(s, &grpcServer{service: svc, inventoryClient: invClient})

	go s.Serve(lis)

	conn, _ := grpc.DialContext(
		context.Background(),
		"buf",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)

	cleanup := func() {
		conn.Close()
		invClient.Close()
		stopInv()
		s.Stop()
	}

	return conn, cleanup
}

func TestGRPC_PostProduct_Integration(t *testing.T) {
	conn, cleanup := startServerIntegrationTest(t)
	defer cleanup()

	client := pb.NewCatalogServiceClient(conn)

	res, err := client.PostProduct(
		context.Background(),
		&pb.PostProductRequest{
			Name:        "Integration Test Product",
			Description: "This is a test product",
			Price:       29.99,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if res.Product.Name != "Integration Test Product" {
		t.Errorf("expected Integration Test Product, got %s", res.Product.Name)
	}

	if res.Product.Id == "" {
		t.Error("expected non-empty product ID")
	}

	if res.Product.Price != 29.99 {
		t.Errorf("expected price 29.99, got %f", res.Product.Price)
	}
}

func TestGRPC_GetProduct_Integration(t *testing.T) {
	conn, cleanup := startServerIntegrationTest(t)
	defer cleanup()

	client := pb.NewCatalogServiceClient(conn)

	// First create a product
	postRes, err := client.PostProduct(
		context.Background(),
		&pb.PostProductRequest{
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       999.99,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Now get the product
	getRes, err := client.GetProduct(
		context.Background(),
		&pb.GetProductRequest{Id: postRes.Product.Id},
	)
	if err != nil {
		t.Fatal(err)
	}

	if getRes.Product.Product.Name != "Laptop" {
		t.Errorf("expected Laptop, got %s", getRes.Product.Product.Name)
	}

	if getRes.Product.Product.Description != "High-performance laptop" {
		t.Errorf("expected High-performance laptop, got %s", getRes.Product.Product.Description)
	}
}

func TestGRPC_GetProducts_Integration(t *testing.T) {
	conn, cleanup := startServerIntegrationTest(t)
	defer cleanup()

	client := pb.NewCatalogServiceClient(conn)

	// Create multiple products
	_, err := client.PostProduct(
		context.Background(),
		&pb.PostProductRequest{Name: "Product A", Description: "Description A", Price: 10.00},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.PostProduct(
		context.Background(),
		&pb.PostProductRequest{Name: "Product B", Description: "Description B", Price: 20.00},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.PostProduct(
		context.Background(),
		&pb.PostProductRequest{Name: "Product C", Description: "Description C", Price: 30.00},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Get products with pagination
	res, err := client.GetProducts(
		context.Background(),
		&pb.GetProductsRequest{Skip: 0, Take: 2},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Products) < 2 {
		t.Errorf("expected at least 2 products, got %d", len(res.Products))
	}
}

func TestGRPC_SearchProducts_Integration(t *testing.T) {
	conn, cleanup := startServerIntegrationTest(t)
	defer cleanup()

	client := pb.NewCatalogServiceClient(conn)

	// Create a product with specific name
	_, err := client.PostProduct(
		context.Background(),
		&pb.PostProductRequest{
			Name:        "Special Gaming Mouse",
			Description: "RGB gaming mouse",
			Price:       49.99,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Search for the product
	res, err := client.GetProducts(
		context.Background(),
		&pb.GetProductsRequest{Query: "Gaming", Skip: 0, Take: 10, Ids: nil},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Products) == 0 {
		t.Error("expected at least one product in search results")
	}

	found := false
	for _, p := range res.Products {
		if p.Product.Name == "Special Gaming Mouse" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find 'Special Gaming Mouse' in search results")
	}
}

func TestGRPC_GetProductsByIds_Integration(t *testing.T) {
	conn, cleanup := startServerIntegrationTest(t)
	defer cleanup()

	client := pb.NewCatalogServiceClient(conn)

	// Create products and collect IDs
	res1, err := client.PostProduct(
		context.Background(),
		&pb.PostProductRequest{Name: "Product X", Description: "Desc X", Price: 15.00},
	)
	if err != nil {
		t.Fatal(err)
	}

	res2, err := client.PostProduct(
		context.Background(),
		&pb.PostProductRequest{Name: "Product Y", Description: "Desc Y", Price: 25.00},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Get products by IDs
	res, err := client.GetProducts(
		context.Background(),
		&pb.GetProductsRequest{Ids: []string{res1.Product.Id, res2.Product.Id}, Skip: 0, Take: 2},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(res.Products))
	}
}

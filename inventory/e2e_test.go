package inventory

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/inventory/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func startE2EServer(t *testing.T) (string, func()) {
	url := os.Getenv("REDIS_URL_FOR_TEST")
	if url == "" {
		url = "redis://localhost:6379"
	}

	repo, err := NewRepository(url)
	if err != nil {
		t.Fatalf("failed to connect to Elasticsearch: %v", err)
	}

	svc := NewService(repo)

	lis, err := net.Listen("tcp", ":9094")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterInventoryServiceServer(s, &grpcServer{service: svc})

	go s.Serve(lis)

	cleanup := func() {
		s.Stop()
		repo.Close()
		lis.Close()
	}

	return "localhost:9094", cleanup
}

func TestE2E_UpdateInventory_Add_Success(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewInventoryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.UpdateStock(
		ctx,
		&pb.UpdateStockRequest{
			Pids:   []string{"p1", "p2"},
			Deltas: []int32{2, 2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.OutOfStock) != 0 {
		t.Errorf("exepected 0 out of stock got %d", len(res.OutOfStock))
	}
}

func TestE2E_UpdateInventory_Add_InvalidInput(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewInventoryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.UpdateStock(
		ctx,
		&pb.UpdateStockRequest{
			Pids:   []string{"p1"},
			Deltas: []int32{2, 2},
		},
	)
	if err == nil {
		t.Errorf("expected error got nil")
	}
}

func TestE2E_UpdateInventory_Remove_Success(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewInventoryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.UpdateStock(
		ctx,
		&pb.UpdateStockRequest{
			Pids:   []string{"p3"},
			Deltas: []int32{2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.UpdateStock(
		ctx,
		&pb.UpdateStockRequest{
			Pids:   []string{"p3"},
			Deltas: []int32{-1},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.OutOfStock) != 0 {
		t.Errorf("exepected 0 out of stock got %d", len(res.OutOfStock))
	}
}

func TestE2E_UpdateInventory_Remove_OutOfStock(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewInventoryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.UpdateStock(
		ctx,
		&pb.UpdateStockRequest{
			Pids:   []string{"p4"},
			Deltas: []int32{-1},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.OutOfStock) == 0 {
		t.Error("exepected out of stock.")
	}

	if res.OutOfStock[0] != "p4" {
		t.Errorf("expected p4 id got %s", res.OutOfStock[0])
	}
}

func TestE2E_CheckInevtory(t *testing.T) {
	addr, cleanup := startE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewInventoryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.UpdateStock(
		ctx,
		&pb.UpdateStockRequest{
			Pids:   []string{"p5"},
			Deltas: []int32{2},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.CheckStock(
		ctx,
		&pb.CheckStockRequest{
			Pids: []string{"p5", "p6"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if res.InStock[0] != 2 && res.InStock[1] != 0 {
		t.Errorf("Invalid output: %#v", res.InStock)
	}
}

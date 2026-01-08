package catalog

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog/pb"
	"github.com/RathodViraj/go-microservice-graphql-grpc/inventory"
	inventorypb "github.com/RathodViraj/go-microservice-graphql-grpc/inventory/pb"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// startFakeInventoryServer spins up a simple in-memory inventory gRPC server for testing.
func startFakeInventoryServer(t *testing.T) (addr string, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(srv, &fakeInventoryServer{})
	go srv.Serve(lis)
	stop = func() {
		srv.Stop()
		_ = lis.Close()
	}
	return lis.Addr().String(), stop
}

type fakeInventoryServer struct {
	inventorypb.UnimplementedInventoryServiceServer
}

func (s *fakeInventoryServer) UpdateStock(ctx context.Context, r *inventorypb.UpdateStockRequest) (*inventorypb.UpdateStockResponse, error) {
	return &inventorypb.UpdateStockResponse{OutOfStock: []string{}}, nil
}

func (s *fakeInventoryServer) CheckStock(ctx context.Context, r *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	in := make([]int32, len(r.Pids))
	for i := range in {
		in[i] = 100
	}
	return &inventorypb.CheckStockResponse{InStock: in}, nil
}

func startTestServer(t *testing.T, svc Service) (*grpc.ClientConn, func()) {
	// Setup inventory client
	invAddr, stopInv := startFakeInventoryServer(t)
	invClient, err := inventory.NewClient(invAddr)
	if err != nil {
		t.Fatalf("failed to create inventory client: %v", err)
	}

	lis := bufconn.Listen(bufSize)

	s := grpc.NewServer()
	pb.RegisterCatalogServiceServer(s, &grpcServer{service: svc, inventoryClient: invClient})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("server exited: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		conn.Close()
		invClient.Close()
		stopInv()
		s.Stop()
	}

	return conn, cleanup
}

func TestServer_PostProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	mockSvc.EXPECT().
		PostProduct(gomock.Any(), "Pen", "Blue ink", 4.99).
		Return(&Product{ID: "p1", Name: "Pen", Description: "Blue ink", Price: 4.99}, nil)

	conn, cleanup := startTestServer(t, mockSvc)
	defer cleanup()
	client := pb.NewCatalogServiceClient(conn)

	_, err := client.PostProduct(context.Background(), &pb.PostProductRequest{Name: "Pen", Description: "Blue ink", Price: 4.99})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServer_GetProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	mockSvc.EXPECT().
		GetProduct(gomock.Any(), "p1").
		Return(&Product{ID: "p1", Name: "Pen", Description: "Blue ink", Price: 4.99}, nil)

	conn, cleanup := startTestServer(t, mockSvc)
	defer cleanup()
	client := pb.NewCatalogServiceClient(conn)

	_, err := client.GetProduct(context.Background(), &pb.GetProductRequest{Id: "p1"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServer_ListProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	mockSvc.EXPECT().
		GetProducts(gomock.Any(), uint64(1), uint64(2)).
		Return([]Product{{ID: "p1"}, {ID: "p2"}}, nil)

	conn, cleanup := startTestServer(t, mockSvc)
	defer cleanup()
	client := pb.NewCatalogServiceClient(conn)

	_, err := client.GetProducts(context.Background(), &pb.GetProductsRequest{Skip: 1, Take: 2})
	if err != nil {
		t.Fatal(err)
	}
}

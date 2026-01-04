package catalog

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog/pb"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func startTestServer(t *testing.T, svc Service) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)

	s := grpc.NewServer()
	pb.RegisterCatalogServiceServer(s, &grpcServer{service: svc})

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

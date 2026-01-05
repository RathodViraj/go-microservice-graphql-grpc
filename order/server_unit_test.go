package order

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/account"
	accountpb "github.com/RathodViraj/go-microservice-graphql-grpc/account/pb"
	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog"
	catalogpb "github.com/RathodViraj/go-microservice-graphql-grpc/catalog/pb"
	"github.com/RathodViraj/go-microservice-graphql-grpc/order/pb"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// startMockAccountServer spins up a real gRPC server backed by a gomock implementation.
func startMockAccountServer(t *testing.T, ctrl *gomock.Controller) (addr string, mock *MockAccountServiceServer, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	mock = NewMockAccountServiceServer(ctrl)
	srv := grpc.NewServer()
	accountpb.RegisterAccountServiceServer(srv, mock)
	go srv.Serve(lis)
	stop = func() {
		srv.Stop()
		_ = lis.Close()
	}
	return lis.Addr().String(), mock, stop
}

// startMockCatalogServer spins up a real gRPC server backed by a gomock implementation.
func startMockCatalogServer(t *testing.T, ctrl *gomock.Controller) (addr string, mock *MockCatalogServiceServer, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	mock = NewMockCatalogServiceServer(ctrl)
	srv := grpc.NewServer()
	catalogpb.RegisterCatalogServiceServer(srv, mock)
	go srv.Serve(lis)
	stop = func() {
		srv.Stop()
		_ = lis.Close()
	}
	return lis.Addr().String(), mock, stop
}

func TestUnitServer_PostOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountAddr, accountMock, stopAccount := startMockAccountServer(t, ctrl)
	defer stopAccount()
	accountClient, err := account.NewClient(accountAddr)
	if err != nil {
		t.Fatalf("failed to create account client: %v", err)
	}
	defer accountClient.Close()

	catalogAddr, catalogMock, stopCatalog := startMockCatalogServer(t, ctrl)
	defer stopCatalog()
	catalogClient, err := catalog.NewClient(catalogAddr)
	if err != nil {
		t.Fatalf("failed to create catalog client: %v", err)
	}
	defer catalogClient.Close()

	accountMock.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(&accountpb.GetAccountResponse{
		Account: &accountpb.Account{Id: "acc1", Name: "Alice"},
	}, nil)

	catalogMock.EXPECT().GetProducts(gomock.Any(), gomock.Any()).Return(&catalogpb.GetProductsResponse{
		Products: []*catalogpb.Product{{
			Id:          "p1",
			Name:        "prod",
			Description: "desc",
			Price:       10,
		}},
	}, nil)

	ctrlService := gomock.NewController(t)
	defer ctrlService.Finish()
	mockService := NewMockService(ctrlService)

	expectedProducts := []OrderedProduct{{ID: "p1", Name: "prod", Description: "desc", Price: 10, Quantity: 2}}
	mockService.EXPECT().PostOrder(gomock.Any(), "acc1", expectedProducts).Return(&Order{
		ID:         "o1",
		AccountID:  "acc1",
		TotalPrice: 20,
		Products:   expectedProducts,
		CreatedAt:  time.Now(),
	}, nil)

	srv := grpcServer{service: mockService, accountClient: accountClient, catalogClient: catalogClient}
	req := &pb.PostOrderRequest{AccountId: "acc1", Products: []*pb.PostOrderRequest_OrderProduct{{ProductId: "p1", Quantity: 2}}}

	resp, err := srv.PostOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetOrder() == nil {
		t.Fatalf("expected order in response")
	}
	if resp.GetOrder().TotalPrice != 20 {
		t.Fatalf("expected total price 20, got %v", resp.GetOrder().TotalPrice)
	}
	if len(resp.GetOrder().Products) != 1 || resp.GetOrder().Products[0].Quantity != 2 {
		t.Fatalf("expected product quantity 2, got %#v", resp.GetOrder().Products)
	}
}

func TestUnitServer_PostOrder_AccountNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountAddr, accountMock, stopAccount := startMockAccountServer(t, ctrl)
	defer stopAccount()
	accountClient, err := account.NewClient(accountAddr)
	if err != nil {
		t.Fatalf("failed to create account client: %v", err)
	}
	defer accountClient.Close()

	catalogAddr, catalogMock, stopCatalog := startMockCatalogServer(t, ctrl)
	defer stopCatalog()
	catalogClient, err := catalog.NewClient(catalogAddr)
	if err != nil {
		t.Fatalf("failed to create catalog client: %v", err)
	}
	defer catalogClient.Close()

	accountMock.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.NotFound, "not found"))

	ctrlService := gomock.NewController(t)
	defer ctrlService.Finish()
	mockService := NewMockService(ctrlService)

	srv := grpcServer{service: mockService, accountClient: accountClient, catalogClient: catalogClient}
	req := &pb.PostOrderRequest{AccountId: "missing", Products: []*pb.PostOrderRequest_OrderProduct{{ProductId: "p1", Quantity: 1}}}

	_, err = srv.PostOrder(context.Background(), req)
	if err == nil || err.Error() != "account not found" {
		t.Fatalf("expected account not found error, got %v", err)
	}

	// Ensure catalog not hit when account fails
	catalogMock.EXPECT().GetProducts(gomock.Any(), gomock.Any()).Times(0)
}

func TestUnitServer_PostOrder_ProductsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountAddr, accountMock, stopAccount := startMockAccountServer(t, ctrl)
	defer stopAccount()
	accountClient, err := account.NewClient(accountAddr)
	if err != nil {
		t.Fatalf("failed to create account client: %v", err)
	}
	defer accountClient.Close()

	catalogAddr, catalogMock, stopCatalog := startMockCatalogServer(t, ctrl)
	defer stopCatalog()
	catalogClient, err := catalog.NewClient(catalogAddr)
	if err != nil {
		t.Fatalf("failed to create catalog client: %v", err)
	}
	defer catalogClient.Close()

	accountMock.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(&accountpb.GetAccountResponse{Account: &accountpb.Account{Id: "acc1", Name: "Alice"}}, nil)
	catalogMock.EXPECT().GetProducts(gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.NotFound, "missing products"))

	ctrlService := gomock.NewController(t)
	defer ctrlService.Finish()
	mockService := NewMockService(ctrlService)

	srv := grpcServer{service: mockService, accountClient: accountClient, catalogClient: catalogClient}
	req := &pb.PostOrderRequest{AccountId: "acc1", Products: []*pb.PostOrderRequest_OrderProduct{{ProductId: "p1", Quantity: 1}}}

	_, err = srv.PostOrder(context.Background(), req)
	if err == nil || err.Error() != "products not found" {
		t.Fatalf("expected products not found error, got %v", err)
	}
}

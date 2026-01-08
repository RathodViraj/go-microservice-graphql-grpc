package order

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/RathodViraj/go-microservice-graphql-grpc/account"
	accountpb "github.com/RathodViraj/go-microservice-graphql-grpc/account/pb"
	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog"
	catalogpb "github.com/RathodViraj/go-microservice-graphql-grpc/catalog/pb"
	"github.com/RathodViraj/go-microservice-graphql-grpc/inventory"
	inventorypb "github.com/RathodViraj/go-microservice-graphql-grpc/inventory/pb"
	"github.com/RathodViraj/go-microservice-graphql-grpc/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var (
	integrationAccountClient   *account.Client
	integrationCatalogClient   *catalog.Client
	integrationInventoryClient *inventory.Client
	integrationOrderClient     pb.OrderServiceClient
)

func setupIntegrationTest(t *testing.T) {
	// Setup real account service with bufconn
	accountURL := os.Getenv("ACCOUNT_DATABASE_URL_FOR_TEST")
	if accountURL == "" {
		accountURL = "postgres://viraj:123456@localhost:5433/viraj?sslmode=disable"
	}
	accountRepo, err := account.NewPostgresRepository(accountURL)
	if err != nil {
		t.Fatalf("failed to create account repo: %v", err)
	}
	accountService := account.NewService(accountRepo)

	accountLis := bufconn.Listen(bufSize)
	accountSrv := grpc.NewServer()
	accountpb.RegisterAccountServiceServer(accountSrv, newAccountGrpcServer(accountService))
	go accountSrv.Serve(accountLis)
	t.Cleanup(func() {
		accountSrv.Stop()
		accountRepo.Close()
		accountLis.Close()
	})

	accountConn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return accountLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create account client: %v", err)
	}
	t.Cleanup(func() { accountConn.Close() })
	integrationAccountClient = newAccountClientWithConn(accountConn)

	// Setup real catalog service with bufconn
	catalogURL := os.Getenv("ELASTICSEARCH_URL_FOR_TEST")
	if catalogURL == "" {
		catalogURL = "http://localhost:9200"
	}
	catalogRepo, err := catalog.NewElasticRepository(catalogURL)
	if err != nil {
		t.Fatalf("failed to create catalog repo: %v", err)
	}
	catalogService := catalog.NewSerivce(catalogRepo)

	catalogLis := bufconn.Listen(bufSize)
	catalogSrv := grpc.NewServer()
	catalogpb.RegisterCatalogServiceServer(catalogSrv, newCatalogGrpcServer(catalogService))
	go catalogSrv.Serve(catalogLis)
	t.Cleanup(func() {
		catalogSrv.Stop()
		catalogRepo.Close()
		catalogLis.Close()
	})

	catalogConn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return catalogLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create catalog client: %v", err)
	}
	t.Cleanup(func() { catalogConn.Close() })
	integrationCatalogClient = newCatalogClientWithConn(catalogConn)

	// Setup fake inventory service with bufconn (in-memory)
	inventoryLis := bufconn.Listen(bufSize)
	inventorySrv := grpc.NewServer()
	invSvc := &fakeInventoryService{}
	inventorypb.RegisterInventoryServiceServer(inventorySrv, newInventoryGrpcServer(invSvc))
	go inventorySrv.Serve(inventoryLis)
	t.Cleanup(func() {
		inventorySrv.Stop()
		inventoryLis.Close()
	})

	inventoryConn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return inventoryLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create inventory client: %v", err)
	}
	t.Cleanup(func() { inventoryConn.Close() })
	integrationInventoryClient = newInventoryClientWithConn(inventoryConn)

	// Setup order service with bufconn
	orderSvc := &orderService{testRepo}
	orderLis := bufconn.Listen(bufSize)
	orderSrv := grpc.NewServer()
	pb.RegisterOrderServiceServer(orderSrv, &grpcServer{
		service:         orderSvc,
		accountClient:   integrationAccountClient,
		catalogClient:   integrationCatalogClient,
		inventoryClient: integrationInventoryClient,
	})
	go orderSrv.Serve(orderLis)
	t.Cleanup(func() {
		orderSrv.Stop()
		orderLis.Close()
	})

	orderConn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return orderLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create order client: %v", err)
	}
	t.Cleanup(func() { orderConn.Close() })

	integrationOrderClient = pb.NewOrderServiceClient(orderConn)
}

// Helper to create inventory gRPC server
func newInventoryGrpcServer(service inventory.Service) inventorypb.InventoryServiceServer {
	return &inventoryGrpcServer{service: service}
}

type inventoryGrpcServer struct {
	inventorypb.UnimplementedInventoryServiceServer
	service inventory.Service
}

func (s *inventoryGrpcServer) UpdateStock(ctx context.Context, r *inventorypb.UpdateStockRequest) (*inventorypb.UpdateStockResponse, error) {
	out, err := s.service.UpdateStock(ctx, r.Pids, r.Deltas)
	if err != nil {
		return nil, err
	}
	return &inventorypb.UpdateStockResponse{OutOfStock: out}, nil
}

func (s *inventoryGrpcServer) CheckStock(ctx context.Context, r *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	inStock, err := s.service.CheckStock(ctx, r.Pids)
	if err != nil {
		return nil, err
	}
	return &inventorypb.CheckStockResponse{InStock: inStock}, nil
}

// Helpers to create clients with bufconn connections

// Helper to create account gRPC server
func newAccountGrpcServer(service account.Service) accountpb.AccountServiceServer {
	return &accountGrpcServer{service: service}
}

type accountGrpcServer struct {
	accountpb.UnimplementedAccountServiceServer
	service account.Service
}

func (s *accountGrpcServer) PostAccount(ctx context.Context, r *accountpb.PostAccountRequest) (*accountpb.PostAccountResponse, error) {
	a, err := s.service.PostAccount(ctx, r.Name)
	if err != nil {
		return nil, err
	}
	return &accountpb.PostAccountResponse{Account: &accountpb.Account{Id: a.ID, Name: a.Name}}, nil
}

func (s *accountGrpcServer) GetAccount(ctx context.Context, r *accountpb.GetAccountRequest) (*accountpb.GetAccountResponse, error) {
	a, err := s.service.GetAccount(ctx, r.Id)
	if err != nil {
		return nil, err
	}
	return &accountpb.GetAccountResponse{Account: &accountpb.Account{Id: a.ID, Name: a.Name}}, nil
}

// Helper to create catalog gRPC server
func newCatalogGrpcServer(service catalog.Service) catalogpb.CatalogServiceServer {
	return &catalogGrpcServer{service: service}
}

type catalogGrpcServer struct {
	catalogpb.UnimplementedCatalogServiceServer
	service catalog.Service
}

func (s *catalogGrpcServer) PostProduct(ctx context.Context, r *catalogpb.PostProductRequest) (*catalogpb.PostProductResponse, error) {
	p, err := s.service.PostProduct(ctx, r.Name, r.Description, r.Price)
	if err != nil {
		return nil, err
	}
	return &catalogpb.PostProductResponse{Product: &catalogpb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}}, nil
}

func (s *catalogGrpcServer) GetProducts(ctx context.Context, r *catalogpb.GetProductsRequest) (*catalogpb.GetProductsResponse, error) {
	var res []catalog.Product
	var err error
	if len(r.Ids) != 0 {
		res, err = s.service.GetProductsById(ctx, r.Ids)
	} else {
		res, err = s.service.GetProducts(ctx, r.Skip, r.Take)
	}
	if err != nil {
		return nil, err
	}
	products := []*catalogpb.ProductInResponse{}
	for _, p := range res {
		products = append(products, &catalogpb.ProductInResponse{
			Product: &catalogpb.Product{
				Id:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			},
			Quntity: 0,
		})
	}
	return &catalogpb.GetProductsResponse{Products: products}, nil
}

// Helpers to create clients with bufconn connections
func newAccountClientWithConn(conn *grpc.ClientConn) *account.Client {
	return &account.Client{
		Conn:    conn,
		Service: accountpb.NewAccountServiceClient(conn),
	}
}

func newCatalogClientWithConn(conn *grpc.ClientConn) *catalog.Client {
	return &catalog.Client{
		Conn:    conn,
		Service: catalogpb.NewCatalogServiceClient(conn),
	}
}

func newInventoryClientWithConn(conn *grpc.ClientConn) *inventory.Client {
	return &inventory.Client{
		Conn:    conn,
		Service: inventorypb.NewInventoryServiceClient(conn),
	}
}

// Simple in-memory inventory service used for tests
type fakeInventoryService struct{}

func (f *fakeInventoryService) UpdateStock(ctx context.Context, pids []string, deltas []int32) ([]string, error) {
	// Always succeed and report no out-of-stock for integration tests
	return []string{}, nil
}

func (f *fakeInventoryService) CheckStock(ctx context.Context, pids []string) ([]int32, error) {
	// Return a large positive stock for all items
	res := make([]int32, len(pids))
	for i := range res {
		res[i] = 100
	}
	return res, nil
}

func TestServer_PostOrder_Success(t *testing.T) {
	setupIntegrationTest(t)

	a, err := integrationAccountClient.PostAccount(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	p, err := integrationCatalogClient.PostProduct(context.Background(), "book", "fiction", 1.42)
	if err != nil {
		t.Fatal(err)
	}

	res, err := integrationOrderClient.PostOrder(
		context.Background(),
		&pb.PostOrderRequest{
			AccountId: a.ID,
			Products: []*pb.PostOrderRequest_OrderProduct{
				{ProductId: p.ID, Quantity: 1},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if res.Order.TotalPrice != 1.42 {
		t.Error("unexpected total price")
	}
}

func TestServer_PostOrder_AccountNotFound(t *testing.T) {
	setupIntegrationTest(t)

	_, err := integrationOrderClient.PostOrder(
		context.Background(),
		&pb.PostOrderRequest{
			AccountId: "non-existing",
			Products: []*pb.PostOrderRequest_OrderProduct{
				{ProductId: "non-existing", Quantity: 1},
			},
		},
	)
	if err == nil {
		t.Error("expected error")
	}
}

func TestServer_PostOrder_ProductNotFound(t *testing.T) {
	setupIntegrationTest(t)

	a, err := integrationAccountClient.PostAccount(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}

	_, err = integrationOrderClient.PostOrder(
		context.Background(),
		&pb.PostOrderRequest{
			AccountId: a.ID,
			Products: []*pb.PostOrderRequest_OrderProduct{
				{ProductId: "non-existing", Quantity: 1},
			},
		},
	)
	if err == nil {
		t.Error("exepected error")
	}
}

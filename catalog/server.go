package catalog

import (
	"context"
	"fmt"
	"net"

	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog/pb"
	"github.com/RathodViraj/go-microservice-graphql-grpc/inventory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedCatalogServiceServer
	service         Service
	inventoryClient *inventory.Client
}

func ListenGRPC(s Service, inventoryURL string, port int) error {
	invetoryClient, err := inventory.NewClient(inventoryURL)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		invetoryClient.Close()
		return err
	}
	srv := grpc.NewServer()
	pb.RegisterCatalogServiceServer(srv, &grpcServer{service: s, inventoryClient: invetoryClient})
	reflection.Register(srv)
	return srv.Serve(lis)
}

func (s *grpcServer) PostProduct(ctx context.Context, r *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	p, err := s.service.PostProduct(ctx, r.Name, r.Description, r.Price)
	if err != nil {
		return nil, err
	}
	return &pb.PostProductResponse{Product: &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}}, nil
}

func (s *grpcServer) GetProduct(ctx context.Context, r *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	p, err := s.service.GetProduct(ctx, r.Id)
	if err != nil {
		return nil, err
	}

	q, err := s.inventoryClient.CheckStock(ctx, []string{r.Id})
	if err != nil {
		return nil, err
	}

	res := &pb.ProductInResponse{
		Product: &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		},
		Quntity: q[0],
	}

	return &pb.GetProductResponse{Product: res}, nil
}

func (s *grpcServer) GetProducts(ctx context.Context, r *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	var (
		res []Product
		err error
	)

	if r.Query != "" {
		res, err = s.service.SearchProduct(ctx, r.Query, r.Skip, r.Take)
	} else if len(r.Ids) != 0 {
		res, err = s.service.GetProductsById(ctx, r.Ids)
	} else {
		res, err = s.service.GetProducts(ctx, r.Skip, r.Take)
	}
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, p := range res {
		ids = append(ids, p.ID)
	}

	quantities, err := s.inventoryClient.CheckStock(ctx, ids)

	products := []*pb.ProductInResponse{}
	for i, p := range res {
		products = append(
			products,
			&pb.ProductInResponse{
				Product: &pb.Product{
					Id:          p.ID,
					Name:        p.Name,
					Description: p.Description,
					Price:       p.Price,
				},
				Quntity: quantities[i],
			},
		)
	}

	return &pb.GetProductsResponse{Products: products}, nil
}

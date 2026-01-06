package inventory

import (
	"context"
	"fmt"
	"net"

	"github.com/RathodViraj/go-microservice-graphql-grpc/inventory/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedInventoryServiceServer
	service Service
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	srv := grpc.NewServer()
	pb.RegisterInventoryServiceServer(srv, &grpcServer{service: s})
	reflection.Register(srv)
	return srv.Serve(lis)
}

func (s *grpcServer) UpdateStock(ctx context.Context, r *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	res, err := s.service.UpdateStock(ctx, r.Pids, r.Deltas)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateStockResponse{OutOfStock: res}, err
}

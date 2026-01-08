package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/RathodViraj/go-microservice-graphql-grpc/account"
	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog"
	"github.com/RathodViraj/go-microservice-graphql-grpc/inventory"
	"github.com/RathodViraj/go-microservice-graphql-grpc/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service         Service
	accountClient   *account.Client
	catalogClient   *catalog.Client
	inventoryClient *inventory.Client
}

func ListenGRPC(s Service, accountURL, catalogURL, inventoryURL string, port int) error {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return err
	}
	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		accountClient.Close()
		return err
	}
	inventroryClient, err := inventory.NewClient(inventoryURL)
	if err != nil {
		accountClient.Close()
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return err
	}
	srv := grpc.NewServer()
	pb.RegisterOrderServiceServer(srv, &grpcServer{service: s, accountClient: accountClient, catalogClient: catalogClient, inventoryClient: inventroryClient})
	reflection.Register(srv)
	return srv.Serve(lis)
}

func (s *grpcServer) PostOrder(ctx context.Context, r *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	_, err := s.accountClient.GetAccount(ctx, r.AccountId)
	if err != nil {
		log.Println("Error getting account:", err)
		return nil, errors.New("account not found")
	}

	productIDs := []string{}
	Quantities := []int32{}
	for _, prd := range r.Products {
		productIDs = append(productIDs, prd.ProductId)
		Quantities = append(Quantities, -1*int32(prd.Quantity))
	}

	outOfStockProducts, err := s.inventoryClient.UpdateStock(ctx, productIDs, Quantities)
	if err != nil {
		log.Println("error checking stock: ", err)
		return nil, errors.New("failed to update stocks")
	}
	if len(outOfStockProducts) != 0 {
		return nil, fmt.Errorf("this product(s) out of stock: %v", outOfStockProducts)
	}

	orderedProducts, err := s.catalogClient.GetProducts(ctx, 0, 0, productIDs, "")
	if err != nil {
		log.Println("Error getting products: ", err)
		return nil, errors.New("products not found")
	}
	products := []OrderedProduct{}
	for _, p := range orderedProducts {
		product := OrderedProduct{
			ID:          p.Product.ID,
			Quantity:    0,
			Price:       p.Product.Price,
			Name:        p.Product.Name,
			Description: p.Product.Description,
		}
		for _, prd := range r.Products {
			if prd.ProductId == p.Product.ID {
				product.Quantity = prd.Quantity
				break
			}
		}
		if product.Quantity != 0 {
			products = append(products, product)
		}
	}

	if len(products) == 0 {
		return nil, errors.New("products not found")
	}

	order, err := s.service.PostOrder(ctx, r.AccountId, products)
	if err != nil {
		log.Println("errors posting err: ", err)
		return nil, errors.New("could not post order")
	}

	orderProto := &pb.Order{
		Id:         order.ID,
		AccountId:  order.AccountID,
		TotalPrice: order.TotalPrice,
		Products:   []*pb.Order_OrderProduct{},
	}

	orderProto.CreatedAt, _ = order.CreatedAt.MarshalBinary()
	for _, p := range order.Products {
		orderProto.Products = append(orderProto.Products, &pb.Order_OrderProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		})
	}

	return &pb.PostOrderResponse{
		Order: orderProto,
	}, nil
}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, r *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {
	accountOrds, err := s.service.GetOrderForAccount(ctx, r.AccountId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	productIDmap := map[string]bool{}
	for _, o := range accountOrds {
		for _, p := range o.Products {
			productIDmap[p.ID] = true
		}
	}

	productIds := []string{}
	for id := range productIDmap {
		productIds = append(productIds, id)
	}

	peoducts, err := s.catalogClient.GetProducts(ctx, 0, 0, productIds, "")
	if err != nil {
		return nil, err
	}

	orders := []*pb.Order{}
	for _, o := range accountOrds {
		op := &pb.Order{
			AccountId:  o.AccountID,
			Id:         o.ID,
			TotalPrice: o.TotalPrice,
			Products:   []*pb.Order_OrderProduct{},
		}
		op.CreatedAt, _ = o.CreatedAt.MarshalBinary()

		for _, product := range o.Products {
			for _, p := range peoducts {
				if p.Product.ID == product.ID {
					product.Name = p.Product.Name
					product.Description = p.Product.Description
					product.Price = p.Product.Price
					break
				}
			}
			op.Products = append(op.Products, &pb.Order_OrderProduct{
				Id:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Quantity:    product.Quantity,
			})
		}
		orders = append(orders, op)
	}

	return &pb.GetOrdersForAccountResponse{Orders: orders}, nil
}

package catalog

import (
	"context"
	"testing"

	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog/pb"
	"go.uber.org/mock/gomock"
)

func TestClient_PostProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPB := NewMockCatalogServiceClient(ctrl)

	mockPB.EXPECT().
		PostProduct(gomock.Any(), &pb.PostProductRequest{Name: "product", Description: "test product", Price: 3.23}).
		Return(&pb.PostProductResponse{Product: &pb.Product{Id: "p1", Name: "product", Description: "test product", Price: 3.23}}, nil)

	c := &Client{
		conn:    nil,
		service: mockPB,
	}

	_, err := c.PostProduct(context.Background(), "product", "test product", 3.23)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_GetProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPB := NewMockCatalogServiceClient(ctrl)
	mockPB.EXPECT().
		GetProduct(gomock.Any(), &pb.GetProductRequest{Id: "p1"}).
		Return(&pb.GetProductResponse{Product: &pb.Product{Id: "p1", Name: "product", Description: "test product", Price: 3.23}}, nil)
	c := &Client{service: mockPB}

	_, err := c.GetProduct(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_ListProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockPB := NewMockCatalogServiceClient(ctrl)
	mockPB.EXPECT().
		GetProducts(gomock.Any(), &pb.GetProductsRequest{Skip: 0, Take: 2}).
		Return(&pb.GetProductsResponse{Products: []*pb.Product{
			{Id: "p1", Name: "product1", Description: "test product1", Price: 3.23},
			{Id: "p2", Name: "product2", Description: "test product2", Price: 4.56},
		}}, nil)
	c := &Client{service: mockPB}

	_, err := c.GetProducts(context.Background(), 0, 2, nil, "")
	if err != nil {
		t.Fatal(err)
	}
}

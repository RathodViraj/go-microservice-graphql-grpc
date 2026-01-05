package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/order/pb"
	"go.uber.org/mock/gomock"
)

func TestClient_PostOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockOrderServiceClient(ctrl)
	createdAt := time.Now()
	createdBytes, _ := createdAt.MarshalBinary()

	mockSvc.EXPECT().PostOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(&pb.PostOrderResponse{
		Order: &pb.Order{
			Id:         "o1",
			AccountId:  "acc1",
			TotalPrice: 20,
			CreatedAt:  createdBytes,
			Products: []*pb.Order_OrderProduct{{
				Id:          "p1",
				Name:        "prod",
				Description: "desc",
				Price:       10,
				Quantity:    2,
			}},
		},
	}, nil)

	client := &Client{service: mockSvc}
	order, err := client.PostOrder(context.Background(), "acc1", []OrderedProduct{{ID: "p1", Quantity: 2, Price: 10, Name: "prod", Description: "desc"}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.ID != "o1" || order.AccountID != "acc1" {
		t.Fatalf("unexpected order: %#v", order)
	}
	if order.TotalPrice != 20 {
		t.Fatalf("expected total 20, got %v", order.TotalPrice)
	}
	if len(order.Products) != 1 || order.Products[0].Quantity != 2 {
		t.Fatalf("unexpected products: %#v", order.Products)
	}
}

func TestClient_PostOrder_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockOrderServiceClient(ctrl)
	mockSvc.EXPECT().PostOrder(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

	client := &Client{service: mockSvc}
	_, err := client.PostOrder(context.Background(), "acc1", []OrderedProduct{{ID: "p1", Quantity: 1}})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}

func TestClient_GetOrdersForAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockOrderServiceClient(ctrl)
	createdAt := time.Now()
	createdBytes, _ := createdAt.MarshalBinary()

	mockSvc.EXPECT().GetOrdersForAccount(gomock.Any(), gomock.Any(), gomock.Any()).Return(&pb.GetOrdersForAccountResponse{
		Orders: []*pb.Order{{
			Id:         "o1",
			AccountId:  "acc1",
			TotalPrice: 30,
			CreatedAt:  createdBytes,
			Products: []*pb.Order_OrderProduct{{
				Id:          "p1",
				Name:        "prod",
				Description: "desc",
				Price:       10,
				Quantity:    3,
			}},
		}},
	}, nil)

	client := &Client{service: mockSvc}
	orders, err := client.GetOrdersForAccount(context.Background(), "acc1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	if orders[0].ID != "o1" || orders[0].TotalPrice != 30 {
		t.Fatalf("unexpected order: %#v", orders[0])
	}
	if len(orders[0].Products) != 1 || orders[0].Products[0].Quantity != 3 {
		t.Fatalf("unexpected products: %#v", orders[0].Products)
	}
}

package order

import (
	"context"
	"testing"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/order/pb"
)

func TestE2E_PostAndGetOrder(t *testing.T) {
	setupIntegrationTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	acc, err := integrationAccountClient.PostAccount(ctx, "E2E Test User")
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	prod, err := integrationCatalogClient.PostProduct(ctx, "E2E Laptop", "High-end laptop", 1500.00)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	postRes, err := integrationOrderClient.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: acc.ID,
		Products: []*pb.PostOrderRequest_OrderProduct{
			{ProductId: prod.ID, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("failed to post order: %v", err)
	}

	if postRes.Order.Id == "" {
		t.Fatal("expected order id")
	}
	if postRes.Order.TotalPrice != 3000.00 {
		t.Errorf("expected total price 3000, got %v", postRes.Order.TotalPrice)
	}
	if len(postRes.Order.Products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(postRes.Order.Products))
	}
	if postRes.Order.Products[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", postRes.Order.Products[0].Quantity)
	}

	// Get orders for account
	getRes, err := integrationOrderClient.GetOrdersForAccount(ctx, &pb.GetOrdersForAccountRequest{AccountId: acc.ID})
	if err != nil {
		t.Fatalf("failed to get orders: %v", err)
	}

	if len(getRes.Orders) == 0 {
		t.Fatal("expected at least 1 order")
	}

	found := false
	for _, order := range getRes.Orders {
		if order.Id == postRes.Order.Id {
			found = true
			if order.TotalPrice != 3000.00 {
				t.Errorf("expected total 3000, got %v", order.TotalPrice)
			}
			break
		}
	}
	if !found {
		t.Error("posted order not found in GetOrdersForAccount")
	}
}

func TestE2E_PostOrderWithMultipleProducts(t *testing.T) {
	setupIntegrationTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	acc, err := integrationAccountClient.PostAccount(ctx, "Multi Product User")
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	prod1, err := integrationCatalogClient.PostProduct(ctx, "Mouse", "Gaming mouse", 50.00)
	if err != nil {
		t.Fatalf("failed to create product 1: %v", err)
	}

	prod2, err := integrationCatalogClient.PostProduct(ctx, "Keyboard", "Mechanical keyboard", 150.00)
	if err != nil {
		t.Fatalf("failed to create product 2: %v", err)
	}

	postRes, err := integrationOrderClient.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: acc.ID,
		Products: []*pb.PostOrderRequest_OrderProduct{
			{ProductId: prod1.ID, Quantity: 3},
			{ProductId: prod2.ID, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("failed to post order: %v", err)
	}

	expectedTotal := 50.00*3 + 150.00
	if postRes.Order.TotalPrice != expectedTotal {
		t.Errorf("expected total %v, got %v", expectedTotal, postRes.Order.TotalPrice)
	}
	if len(postRes.Order.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(postRes.Order.Products))
	}
}

func TestE2E_PostOrderAccountNotFound(t *testing.T) {
	setupIntegrationTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := integrationOrderClient.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: "non-existing-account",
		Products: []*pb.PostOrderRequest_OrderProduct{
			{ProductId: "some-product", Quantity: 1},
		},
	})
	if err == nil {
		t.Fatal("expected error for non-existing account")
	}
}

func TestE2E_PostOrderProductNotFound(t *testing.T) {
	setupIntegrationTest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	acc, err := integrationAccountClient.PostAccount(ctx, "Product Not Found User")
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	_, err = integrationOrderClient.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: acc.ID,
		Products: []*pb.PostOrderRequest_OrderProduct{
			{ProductId: "non-existing-product-id", Quantity: 1},
		},
	})
	if err == nil {
		t.Fatal("expected error for non-existing product")
	}
}


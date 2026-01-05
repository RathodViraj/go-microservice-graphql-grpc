package order

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestUnitService_PostOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	svc := NewOrderService(mockRepo)

	products := []OrderedProduct{
		{ID: "p1", Name: "Product 1", Description: "Desc 1", Price: 10.50, Quantity: 2},
		{ID: "p2", Name: "Product 2", Description: "Desc 2", Price: 5.25, Quantity: 3},
	}

	mockRepo.EXPECT().
		PutOrder(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, order Order) error {
			expectedTotal := (10.50 * 2) + (5.25 * 3)
			if order.TotalPrice != expectedTotal {
				t.Errorf("expected total price %f, got %f", expectedTotal, order.TotalPrice)
			}
			if order.AccountID != "a1" {
				t.Errorf("expected account ID 'a1', got %s", order.AccountID)
			}
			if len(order.Products) != 2 {
				t.Errorf("expected 2 products, got %d", len(order.Products))
			}
			if order.ID == "" {
				t.Error("expected non-empty order ID")
			}
			return nil
		})

	result, err := svc.PostOrder(context.Background(), "a1", products)
	if err != nil {
		t.Fatal(err)
	}

	if result.ID == "" {
		t.Error("expected non-empty order ID")
	}

	expectedTotal := (10.50 * 2) + (5.25 * 3)
	if result.TotalPrice != expectedTotal {
		t.Errorf("expected total price %f, got %f", expectedTotal, result.TotalPrice)
	}

	if result.AccountID != "a1" {
		t.Errorf("expected account ID 'a1', got %s", result.AccountID)
	}

	if len(result.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(result.Products))
	}
}

func TestUnitService_PostOrder_EmptyProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	svc := NewOrderService(mockRepo)

	mockRepo.EXPECT().
		PutOrder(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, order Order) error {
			if order.TotalPrice != 0 {
				t.Errorf("expected total price 0, got %f", order.TotalPrice)
			}
			return nil
		})

	result, err := svc.PostOrder(context.Background(), "a1", []OrderedProduct{})
	if err != nil {
		t.Fatal(err)
	}

	if result.TotalPrice != 0 {
		t.Errorf("expected total price 0 for empty order, got %f", result.TotalPrice)
	}
}

func TestUnitService_PostOrder_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	svc := NewOrderService(mockRepo)

	products := []OrderedProduct{
		{ID: "p1", Name: "Product 1", Price: 10.00, Quantity: 1},
	}

	expectedErr := fmt.Errorf("database error")
	mockRepo.EXPECT().
		PutOrder(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	result, err := svc.PostOrder(context.Background(), "a1", products)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestUnitService_PostOrder_TotalPriceCalculation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	svc := NewOrderService(mockRepo)

	testCases := []struct {
		name          string
		products      []OrderedProduct
		expectedTotal float64
	}{
		{
			name: "Single product",
			products: []OrderedProduct{
				{ID: "p1", Name: "Product 1", Price: 10.50, Quantity: 1},
			},
			expectedTotal: 10.50,
		},
		{
			name: "Multiple quantities",
			products: []OrderedProduct{
				{ID: "p1", Name: "Product 1", Price: 10.00, Quantity: 5},
			},
			expectedTotal: 50.00,
		},
		{
			name: "Multiple products",
			products: []OrderedProduct{
				{ID: "p1", Name: "Product 1", Price: 10.00, Quantity: 2},
				{ID: "p2", Name: "Product 2", Price: 5.50, Quantity: 3},
				{ID: "p3", Name: "Product 3", Price: 7.25, Quantity: 1},
			},
			expectedTotal: (10.00 * 2) + (5.50 * 3) + (7.25 * 1),
		},
		{
			name: "Decimal prices",
			products: []OrderedProduct{
				{ID: "p1", Name: "Product 1", Price: 9.99, Quantity: 3},
				{ID: "p2", Name: "Product 2", Price: 0.99, Quantity: 5},
			},
			expectedTotal: (9.99 * 3) + (0.99 * 5),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.EXPECT().
				PutOrder(gomock.Any(), gomock.Any()).
				Return(nil)

			result, err := svc.PostOrder(context.Background(), "a1", tc.products)
			if err != nil {
				t.Fatal(err)
			}

			if result.TotalPrice != tc.expectedTotal {
				t.Errorf("expected total price %f, got %f", tc.expectedTotal, result.TotalPrice)
			}
		})
	}
}

func TestUnitService_GetOrderForAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	svc := NewOrderService(mockRepo)

	expectedOrders := []Order{
		{
			ID:         "o1",
			AccountID:  "a1",
			TotalPrice: 100.50,
			Products: []OrderedProduct{
				{ID: "p1", Name: "Product 1", Price: 50.25, Quantity: 2},
			},
		},
		{
			ID:         "o2",
			AccountID:  "a1",
			TotalPrice: 75.00,
			Products: []OrderedProduct{
				{ID: "p2", Name: "Product 2", Price: 25.00, Quantity: 3},
			},
		},
	}

	mockRepo.EXPECT().
		GetOrderForAccount(gomock.Any(), "a1").
		Return(expectedOrders, nil)

	result, err := svc.GetOrderForAccount(context.Background(), "a1")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 orders, got %d", len(result))
	}

	if result[0].ID != "o1" {
		t.Errorf("expected order ID 'o1', got %s", result[0].ID)
	}

	if result[1].ID != "o2" {
		t.Errorf("expected order ID 'o2', got %s", result[1].ID)
	}
}

func TestUnitService_GetOrderForAccount_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	svc := NewOrderService(mockRepo)

	expectedErr := fmt.Errorf("database error")
	mockRepo.EXPECT().
		GetOrderForAccount(gomock.Any(), "a1").
		Return(nil, expectedErr)

	result, err := svc.GetOrderForAccount(context.Background(), "a1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestUnitService_GetOrderForAccount_DifferentAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	svc := NewOrderService(mockRepo)

	accountIDs := []string{"a1", "a2", "a3"}

	for _, accountID := range accountIDs {
		mockRepo.EXPECT().
			GetOrderForAccount(gomock.Any(), accountID).
			Return([]Order{{ID: "o-" + accountID, AccountID: accountID}}, nil)

		result, err := svc.GetOrderForAccount(context.Background(), accountID)
		if err != nil {
			t.Fatalf("unexpected error for account %s: %v", accountID, err)
		}

		if len(result) != 1 {
			t.Errorf("expected 1 order for account %s, got %d", accountID, len(result))
		}

		if result[0].AccountID != accountID {
			t.Errorf("expected account ID %s, got %s", accountID, result[0].AccountID)
		}
	}
}

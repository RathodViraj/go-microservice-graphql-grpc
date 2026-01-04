package catalog

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestService_PostProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)

	svc := &catalogService{repository: mockRepo}

	mockRepo.EXPECT().
		PutProduct(gomock.Any(), gomock.AssignableToTypeOf(Product{})).
		Return(nil)

	p, err := svc.PostProduct(
		context.Background(),
		"Pen",
		"Blue ink",
		4.99,
	)
	if err != nil {
		t.Fatal(err)
	}

	if p.ID == "" {
		t.Error("expeced non empty ID")
	}
	if p.Name != "Pen" || p.Description != "Blue ink" || p.Price != 4.99 {
		t.Errorf("unexpected output: %#v", p)
	}
}

func TestService_GetProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)

	svc := &catalogService{repository: mockRepo}

	expected := &Product{ID: "p1", Name: "Book"}

	mockRepo.EXPECT().
		GetProductByID(gomock.Any(), "p1").
		Return(expected, nil)

	res, err := svc.GetProduct(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Name != "Book" {
		t.Errorf("expected Book, got %s", res.Name)
	}
}

func TestService_GetProducts_DefaultTakeApplied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	svc := &catalogService{repository: mockRepo}

	mockRepo.EXPECT().
		ListProducts(gomock.Any(), uint64(0), uint64(100)).
		Return([]Product{}, nil)

	_, err := svc.GetProducts(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_GetProductsById_DelegatesToRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	svc := &catalogService{repository: mockRepo}

	ids := []string{"p1", "p2"}

	mockRepo.EXPECT().
		ListProductsWithIDs(gomock.Any(), ids).
		Return([]Product{}, nil)

	_, err := svc.GetProductsById(context.Background(), ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

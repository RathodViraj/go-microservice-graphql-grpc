package order

import (
	"context"
	"testing"
)

func TestService_PostOrder_Success(t *testing.T) {
	products := []OrderedProduct{
		{ID: "p1", Quantity: 2, Price: 1.82},
		{ID: "p2", Quantity: 1, Price: 2.64},
	}

	svc := &orderService{testRepo}

	o, err := svc.PostOrder(context.Background(), "alice1321", products)
	if err != nil {
		t.Fatal(err)
	}

	if o.TotalPrice != 2*products[0].Price+products[1].Price {
		t.Error("unexpected total price")
	}
}

func TestService_GetOrderByAccount(t *testing.T) {
	products := []OrderedProduct{
		{ID: "p1", Quantity: 2, Price: 1.82},
		{ID: "p2", Quantity: 1, Price: 2.64},
	}

	svc := &orderService{testRepo}

	_, err := svc.PostOrder(context.Background(), "bob4532", products)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.PostOrder(context.Background(), "alice1321", products)
	if err != nil {
		t.Fatal(err)
	}

	ords, err := svc.GetOrderForAccount(context.Background(), "bob4532")
	if err != nil {
		t.Fatal(err)
	}

	if len(ords) != 1 {
		t.Errorf("expected 1 order; got %d", len(ords))
	}
}

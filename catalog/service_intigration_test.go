package catalog

import (
	"context"
	"testing"
)

func TestService_PostGetProduct(t *testing.T) {
	svc := &catalogService{testRepo}
	ctx := context.Background()

	p, err := svc.PostProduct(ctx, "Pen", "black ink", 1.92)
	if err != nil {
		t.Fatal(err)
	}

	if p.ID == "" {
		t.Errorf("expected non-empty ID")
	}

	getPrd, err := svc.GetProduct(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}

	if getPrd.Name != "Pen" || getPrd.Description != "black ink" || getPrd.Price != 1.92 {
		t.Errorf("unexpected output: %#v", p)
	}
}

func TestService_SearchProdct(t *testing.T) {
	svc := &catalogService{testRepo}
	ctx := context.Background()

	_, err := svc.PostProduct(ctx, "Pen", "black ink", 1.92)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.PostProduct(ctx, "Pen", "red ink", 2.64)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.PostProduct(ctx, "Pen", "bue ink", 1)
	if err != nil {
		t.Fatal(err)
	}

	ps, err := svc.SearchProduct(ctx, "Pen", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) == 0 {
		t.Error("0 produts was not exepected")
	}
}

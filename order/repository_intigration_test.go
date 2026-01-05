package order

import (
	"context"
	"os"
	"testing"
	"time"
)

var testRepo Repository

func TestMain(m *testing.M) {
	url := os.Getenv("DATABASE_URL_FOR_TEST")
	if url == "" {
		url = "postgres://viraj:123456@localhost:5434/viraj?sslmode=disable"
	}

	r, err := NewPostgresRepository(url)
	if err != nil {
		panic(err)
	}

	testRepo = r
	code := m.Run()
	r.Close()
	os.Exit(code)
}

func TestRepository_PutOrder(t *testing.T) {
	o := Order{
		ID:         "o1",
		CreatedAt:  time.Now(),
		AccountID:  "a1",
		TotalPrice: 100,
		Products: []OrderedProduct{
			{ID: "p1", Quantity: 2},
			{ID: "p2", Quantity: 1},
		},
	}

	if err := testRepo.PutOrder(context.Background(), o); err != nil {
		t.Fatal(err)
	}
}

func TestRepository_GetByAccount(t *testing.T) {
	o1 := Order{
		ID:         "o2",
		CreatedAt:  time.Now(),
		AccountID:  "a2",
		TotalPrice: 100,
		Products: []OrderedProduct{
			{ID: "p1", Quantity: 2},
			{ID: "p2", Quantity: 1},
		},
	}
	if err := testRepo.PutOrder(context.Background(), o1); err != nil {
		t.Fatal(err)
	}

	o2 := Order{
		ID:         "o3",
		CreatedAt:  time.Now(),
		AccountID:  "a2",
		TotalPrice: 100,
		Products: []OrderedProduct{
			{ID: "p3", Quantity: 1},
		},
	}
	if err := testRepo.PutOrder(context.Background(), o2); err != nil {
		t.Fatal(err)
	}

	ords, err := testRepo.GetOrderForAccount(context.Background(), "a2")
	if err != nil {
		t.Fatal(err)
	}

	if len(ords) != 2 {
		t.Errorf("Expected 2; got %d", len(ords))
	}
	if len(ords[0].Products) != 2 && len(ords[1].Products) != 2 {
		t.Errorf("unexepected output: %#v", ords)
	}
}

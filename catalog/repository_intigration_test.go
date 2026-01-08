package catalog

import (
	"context"
	"os"
	"testing"
)

var testRepo Repository

func TestMain(m *testing.M) {
	url := os.Getenv("ELASTICSEARCH_URL_FOR_TEST")
	if url == "" {
		url = "http://localhost:9200"
	}

	r, err := NewElasticRepository(url)
	if err != nil {
		panic(err)
	}

	testRepo = r
	code := m.Run()
	r.Close()
	os.Exit(code)
}

func TestElasticRepository_PutAndGetProduct(t *testing.T) {
	err := testRepo.PutProduct(
		context.Background(),
		Product{
			ID:          "test-prod-1",
			Name:        "Test Product 1",
			Description: "This is a test product",
			Price:       19.99,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	prod, err := testRepo.GetProductByID(context.Background(), "test-prod-1")
	if err != nil {
		t.Fatal(err)
	}
	if prod.ID != "test-prod-1" || prod.Name != "Test Product 1" {
		t.Errorf("expected product ID 'test-prod-1' and Name 'Test Product 1', got ID '%s' and Name '%s'", prod.ID, prod.Name)
	}
}

func TestElasticRepository_GetProduct_NotFound(t *testing.T) {
	_, err := testRepo.GetProductByID(context.Background(), "non-existent-id")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound for non-existent product ID, got %v", err)
	}
}

func TestElasticRepository_ListProducts(t *testing.T) {
	err := testRepo.PutProduct(
		context.Background(),
		Product{
			ID:          "test-prod-3",
			Name:        "Test Product 1",
			Description: "This is a test product",
			Price:       11,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = testRepo.PutProduct(
		context.Background(),
		Product{
			ID:          "test-prod-1",
			Name:        "Test Product 1",
			Description: "This is a test product",
			Price:       2.99,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = testRepo.PutProduct(
		context.Background(),
		Product{
			ID:          "test-prod-4",
			Name:        "Test Product 1",
			Description: "This is a test product",
			Price:       14.99,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	products, err := testRepo.ListProducts(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(products) != 2 {
		t.Errorf("expected 2 products, got %d", len(products))
	}
}

func TestElasticRepository_SearchProduct(t *testing.T) {
	err := testRepo.PutProduct(
		context.Background(),
		Product{
			ID:          "test-prod-5",
			Name:        "Include Product",
			Description: "This is a test product",
			Price:       11,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = testRepo.PutProduct(
		context.Background(),
		Product{
			ID:          "test-prod-6",
			Name:        "Test Product 1",
			Description: "This is a test product",
			Price:       2.99,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = testRepo.PutProduct(
		context.Background(),
		Product{
			ID:          "test-prod-7",
			Name:        "Include Product",
			Description: "This is a test product",
			Price:       14.99,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	products, err := testRepo.SearchProducts(context.Background(), "Include", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) < 1 {
		t.Errorf("expected at least 1 product, got %d", len(products))
	}
}

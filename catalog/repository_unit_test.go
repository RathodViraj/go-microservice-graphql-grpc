package catalog

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

type mockTransport struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (m mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

func mockResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header: http.Header{
			"Content-Type":      []string{"application/json"},
			"X-Elastic-Product": []string{"Elasticsearch"},
		},
	}
}

func TestPutProduct_Success(t *testing.T) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport{
			fn: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/" {
					return mockResponse(200, `{
						"version": {
							"number": "8.0.0"
						}
					}`), nil
				}
				if req.Method != "PUT" {
					t.Errorf("expected PUT, got %s", req.Method)
				}
				return mockResponse(201, `{"result":"created"}`), nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	mockRepo := &elasticRepository{client}

	err = mockRepo.PutProduct(
		context.Background(),
		Product{
			ID:          "p1",
			Name:        "unit test product",
			Description: "put test",
			Price:       10.02,
		},
	)
	if err != nil {
		t.Error(err)
	}
}

func TestGetProductByID_Success(t *testing.T) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport{
			fn: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/" {
					return mockResponse(200, `{
						"version": {
							"number": "8.0.0"
						}
					}`), nil
				}
				if req.Method != "GET" {
					t.Errorf("expected GET, got %s", req.Method)
				}
				return mockResponse(
					200,
					`{
				  	"_id": "p1",
				 	 "_source": { "name":"Pen","description":"Blue","price":5 }
					}`,
				), nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	mockRepo := &elasticRepository{client}

	p, err := mockRepo.GetProductByID(context.Background(), "p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Name != "Pen" || p.Description != "Blue" || p.Price != 5 {
		t.Errorf("expected Pen, got %s", p.Name)
	}
}

func TestGetProductByID_NotFound(t *testing.T) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport{
			fn: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/" {
					return mockResponse(200, `{
						"version": {
							"number": "8.0.0"
						}
					}`), nil
				}
				if req.Method != "GET" {
					t.Errorf("expected GET, got %s", req.Method)
				}
				return mockResponse(404, `{}`), nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	mockRepo := &elasticRepository{client}

	_, err = mockRepo.GetProductByID(context.Background(), "nonexistent")
	if err.Error() != "Entity not found" {
		t.Fatal(err)
	}
}

func TestListProducts_Success(t *testing.T) {
	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport{
			fn: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/" {
					return mockResponse(200, `{
						"version": {
							"number": "8.0.0"
						}
					}`), nil
				}
				return mockResponse(200, `
				{
				  "hits": {
				    "hits": [
				      {"_id":"p1","_source":{"name":"A","description":"d","price":1}},
				      {"_id":"p2","_source":{"name":"B","description":"e","price":2}}
				    ]
				  }
				}`), nil
			},
		},
	})

	mockRepo := &elasticRepository{client: client}

	res, err := mockRepo.ListProducts(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res) != 2 {
		t.Errorf("expected 2 products, got %d", len(res))
	}
}

func TestSearchProducts_WithQuery(t *testing.T) {
	client, _ := elasticsearch.NewClient(elasticsearch.Config{
		Transport: mockTransport{
			fn: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/" {
					return mockResponse(200, `{
						"version": {
							"number": "8.0.0"
						}
					}`), nil
				}
				return mockResponse(200, `
				{
				  "hits": {
				    "hits": [
				      {"_id":"p1","_source":{"name":"include_product","description":"d","price":1}}
				    ]
				  }
				}`), nil
			},
		},
	})

	mockRepo := &elasticRepository{client: client}

	res, err := mockRepo.SearchProducts(context.Background(), "include product", 1, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 {
		t.Errorf("Expected 1; got %d", len(res))
	}
}

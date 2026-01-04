package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

var (
	ErrNotFound = errors.New("Entity not found")
)

type Repository interface {
	Close()
	PutProduct(ctx context.Context, p Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip, take uint64) ([]Product, error)
}

type elasticRepository struct {
	client *elasticsearch.Client
}

type productDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func NewElasticRepository(url string) (Repository, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{url},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating elasticsearch client: %w", err)
	}

	_, err = client.Info(client.Info.WithContext(context.Background()))
	if err != nil {
		return nil, fmt.Errorf("error connecting to elasticsearch: %w", err)
	}

	return &elasticRepository{client}, nil
}

func (r *elasticRepository) Close() {
	// The official client doesn't require explicit close
}

func (r *elasticRepository) PutProduct(ctx context.Context, p Product) error {
	body := productDocument{
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	res, err := r.client.Index(
		"catalog",
		bytes.NewReader(data),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithDocumentID(p.ID),
		r.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	return nil
}

func (r *elasticRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	res, err := r.client.Get("catalog", id, r.client.Get.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, ErrNotFound
	}

	if res.IsError() {
		return nil, fmt.Errorf("error getting document: %s", res.String())
	}

	var result struct {
		Source productDocument `json:"_source"`
		ID     string          `json:"_id"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &Product{
		ID:          result.ID,
		Name:        result.Source.Name,
		Description: result.Source.Description,
		Price:       result.Source.Price,
	}, nil
}

func (r *elasticRepository) ListProducts(ctx context.Context, skip, take uint64) ([]Product, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"from": skip,
		"size": take,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("catalog"),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching documents: %s", res.String())
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source productDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	products := []Product{}
	for _, hit := range result.Hits.Hits {
		products = append(products, Product{
			ID:          hit.ID,
			Name:        hit.Source.Name,
			Description: hit.Source.Description,
			Price:       hit.Source.Price,
		})
	}

	return products, nil
}

func (r *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": ids,
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("catalog"),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching documents: %s", res.String())
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source productDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	products := []Product{}
	for _, hit := range result.Hits.Hits {
		products = append(products, Product{
			ID:          hit.ID,
			Name:        hit.Source.Name,
			Description: hit.Source.Description,
			Price:       hit.Source.Price,
		})
	}

	return products, nil

}

func (r *elasticRepository) SearchProducts(ctx context.Context, query string, skip, take uint64) ([]Product, error) {
	var buf bytes.Buffer
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name", "description"},
			},
		},
		"from": skip,
		"size": take,
	}

	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("catalog"),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching documents: %s", res.String())
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string          `json:"_id"`
				Source productDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	products := []Product{}
	for _, hit := range result.Hits.Hits {
		products = append(products, Product{
			ID:          hit.ID,
			Name:        hit.Source.Name,
			Description: hit.Source.Description,
			Price:       hit.Source.Price,
		})
	}

	return products, nil
}

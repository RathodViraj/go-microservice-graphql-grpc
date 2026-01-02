package catalog

import (
	"context"
	"fmt"
	"log"

	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.CatalogServiceClient
}

func NewClient(url string) (*Client, error) {
	log.Printf("Connecting to catalog service at %s", url)
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create catalog client: %w", err)
	}

	c := pb.NewCatalogServiceClient(conn)
	log.Printf("Catalog client created successfully")
	return &Client{conn, c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostProduct(ctx context.Context, name, description string, price float64) (*Product, error) {
	log.Printf("PostProduct called with name=%s, description=%s, price=%f", name, description, price)
	res, err := c.service.PostProduct(
		ctx,
		&pb.PostProductRequest{
			Name:        name,
			Description: description,
			Price:       price,
		},
	)

	if err != nil {
		return nil, err
	}

	return &Product{
		ID:          res.Product.Id,
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Price:       res.Product.Price,
	}, nil
}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	res, err := c.service.GetProduct(
		ctx,
		&pb.GetProductRequest{
			Id: id,
		},
	)
	if err != nil {
		return nil, err
	}

	return &Product{
		ID:          res.Product.Id,
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Price:       res.Product.Price,
	}, nil
}

func (c *Client) GetProducts(ctx context.Context, skip, take uint64, ids []string, query string) ([]Product, error) {
	res, err := c.service.GetProducts(
		ctx,
		&pb.GetProductsRequest{
			Skip:  skip,
			Take:  take,
			Ids:   ids,
			Query: query,
		},
	)
	if err != nil {
		return nil, err
	}

	var products []Product
	for _, p := range res.Products {
		products = append(products, Product{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})
	}

	return products, nil
}

package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/order"
)

var (
	ErrInvalidParameter = errors.New("inavlid parameter")
)

type mutationResolver struct {
	server *Server
}

func (r *mutationResolver) CreateAccount(ctx context.Context, in AccountInput) (*Account, error) {
	log.Printf("CreateAccount called with name: %s", in.Name)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	account, err := r.server.accountClient.PostAccount(ctx, in.Name)
	if err != nil {
		log.Printf("ERROR in CreateAccount: %v", err)
		return nil, err
	}

	log.Printf("Account created successfully: ID=%s, Name=%s", account.ID, account.Name)
	return &Account{
		ID:   account.ID,
		Name: account.Name,
	}, nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, in ProductInput) (*Product, error) {
	log.Printf("CreateProduct called with name: %s", in.Name)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Calling catalog service...")
	product, err := r.server.catalogClient.PostProduct(ctx, in.Name, in.Description, in.Price)
	if err != nil {
		log.Printf("ERROR in CreateProduct: %v", err)
		return nil, err
	}

	return &Product{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}, nil
}

func (r *mutationResolver) CreateOrder(ctx context.Context, in OrderInput) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var products []order.OrderedProduct
	for _, p := range in.Products {
		if p.Quantity <= 0 {
			return nil, ErrInvalidParameter
		}
		products = append(products, order.OrderedProduct{
			ID:       p.ID,
			Quantity: uint32(p.Quantity),
		})
	}

	order, err := r.server.orderClient.PostOrder(ctx, in.AccountID, products)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Order{
		ID:         order.ID,
		CreatedAt:  order.CreatedAt,
		TotalPrice: order.TotalPrice,
	}, nil
}

func (r *mutationResolver) UpdateStock(ctx context.Context, requests UpdateStocksRequestInput) (*OutOfStock, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var delatas []int32
	for _, d := range requests.Deltas {
		delatas = append(delatas, int32(d))
	}

	outOfStock, err := r.server.inventoryClient.UpdateStock(ctx, requests.Ids, delatas)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &OutOfStock{Ids: outOfStock}, nil
}

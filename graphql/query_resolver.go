package main

import (
	"context"
	"log"
	"time"
)

type queryResolver struct {
	server *Server
}

func (r *queryResolver) Accounts(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if id != nil {
		res, err := r.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return []*Account{{
			ID:   res.ID,
			Name: res.Name,
		}}, nil
	}

	skip, take := uint64(0), uint64(0)
	if pagination != nil {
		skip, take = pagination.bounds()
	}

	accountsList, err := r.server.accountClient.GetAccounts(ctx, skip, take)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var accounts []*Account
	for _, a := range accountsList {
		accounts = append(accounts, &Account{ID: a.ID, Name: a.Name})
	}

	return accounts, nil
}

func (r *queryResolver) Products(ctx context.Context, pagination *PaginationInput, query, id *string) ([]*ProductInResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if id != nil {
		res, err := r.server.catalogClient.GetProduct(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		resp := []*ProductInResponse{}
		resp = append(
			resp,
			&ProductInResponse{
				Product: &Product{
					ID:          res.Product.ID,
					Name:        res.Product.Name,
					Description: res.Product.Description,
					Price:       res.Product.Price,
				},
				Quantity: int(res.Quantity),
			},
		)

		return resp, nil
	}

	skip, take := uint64(0), uint64(0)
	if pagination != nil {
		skip, take = pagination.bounds()
	}

	q := ""
	if query != nil {
		q = *query
	}
	productsList, err := r.server.catalogClient.GetProducts(ctx, skip, take, nil, q)
	if err != nil {
		return nil, err
	}

	var products []*ProductInResponse

	for _, p := range productsList {
		products = append(
			products,
			&ProductInResponse{
				Product: &Product{
					ID:          p.Product.ID,
					Name:        p.Product.Name,
					Description: p.Product.Description,
					Price:       p.Product.Price,
				},
				Quantity: int(p.Quantity),
			},
		)
	}

	return products, nil
}

func (p PaginationInput) bounds() (uint64, uint64) {
	skipValue := uint64(p.Skip)
	takeValue := uint64(p.Take)

	return skipValue, takeValue
}

func (r *queryResolver) CheckStock(ctx context.Context, in *CheckStockInput) ([]int, error) {
	res_int32, err := r.server.inventoryClient.CheckStock(ctx, in.Ids)
	if err != nil {
		return nil, err
	}

	res := []int{}
	for _, r := range res_int32 {
		res = append(res, int(r))
	}
	return res, nil
}

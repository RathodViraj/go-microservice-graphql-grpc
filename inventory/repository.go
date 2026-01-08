package inventory

import (
	"context"
	"embed"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Repository interface {
	Close()
	UpdateStock(ctx context.Context, requests []Stock) ([]string, error)
	CheckStock(ctx context.Context, pids []string) ([]int32, error)
}

type redisRepository struct {
	client *redis.Client
	script *redis.Script
}

func NewRepository(redisURL string) (Repository, error) {
	script := redis.NewScript(script)
	if script == nil {
		return nil, fmt.Errorf("couldn't load the lua script")
	}

	addr := redisURL
	if strings.HasPrefix(redisURL, "redis://") {
		parsedURL, err := url.Parse(redisURL)
		if err != nil {
			return nil, fmt.Errorf("invalid redis URL: %v", err)
		}
		addr = parsedURL.Host
	}
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		},
	)

	return &redisRepository{
		client: client,
		script: script,
	}, nil
}

func (r *redisRepository) Close() {
	r.client.Close()
}

func (r *redisRepository) UpdateStock(ctx context.Context, requests []Stock) ([]string, error) {
	var keys []string
	var args []int32
	for _, s := range requests {
		key := fmt.Sprintf("inventory:%s", s.Product_id)
		keys = append(keys, key)
		args = append(args, s.Delta)
	}

	interfaceArgs := make([]interface{}, len(args))
	for i, v := range args {
		interfaceArgs[i] = v
	}

	res, err := r.script.Run(
		ctx,
		r.client,
		keys,
		interfaceArgs...,
	).Result()
	if err != nil {
		return nil, err
	}

	switch v := res.(type) {
	case string:
		return nil, nil
	case []any:
		out_of_stock := make([]string, 0, len(v))
		for _, x := range v {
			out_of_stock = append(out_of_stock, strings.TrimPrefix(x.(string), "inventory:"))
		}
		return out_of_stock, nil
	default:
	}

	return nil, nil
}

func (r *redisRepository) CheckStock(ctx context.Context, pids []string) ([]int32, error) {
	inStock := []int32{}
	base := "inventory:"
	for _, id := range pids {
		q, err := r.client.Get(ctx, base+id).Result()
		if err != nil {
			inStock = append(inStock, 0) // return out of stock
		}
		q_int, err := strconv.Atoi(q)
		if err != nil {
			inStock = append(inStock, 0) // return out of stock
		}
		inStock = append(inStock, int32(q_int))
	}

	return inStock, nil
}

//go:embed script.lua
var script string

// reference embed to satisfy linters that don't detect go:embed usage
var _ embed.FS

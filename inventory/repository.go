package inventory

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Repository interface {
	Close()
	UpdateStock(ctx context.Context, requests []Stock) ([]string, error)
}

type redisRepository struct {
	client *redis.Client
	script *redis.Script
}

func NewRepository(url string) (Repository, error) {
	script := redis.NewScript(script)
	if script == nil {
		return nil, fmt.Errorf("couldn't load the lua script")
	}
	if url == "" {
		url = "localhost:6379"
	}
	client := redis.NewClient(
		&redis.Options{
			Addr:     url,
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

//go:embed script.lua
var script string

// reference embed to satisfy linters that don't detect go:embed usage
var _ embed.FS

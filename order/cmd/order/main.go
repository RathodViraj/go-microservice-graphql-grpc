package main

import (
	"log"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/order"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	AccountURL  string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL  string `envconfig:"CATALOG_SERVICE_URL"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Defaults for local development
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "postgres://viraj:123456@localhost:5434/viraj?sslmode=disable"
	}
	if cfg.AccountURL == "" {
		cfg.AccountURL = "localhost:8081"
	}
	if cfg.CatalogURL == "" {
		cfg.CatalogURL = "localhost:8082"
	}

	var r order.Repository
	for {
		r, err = order.NewPostgresRepository(cfg.DatabaseURL)
		if err == nil {
			break
		}
		log.Println(err)
		time.Sleep(2 * time.Second)
	}
	defer r.Close()

	log.Println("Listeneing on port 8083...")
	s := order.NewOrderService(r)
	log.Fatal(order.ListenGRPC(s, cfg.AccountURL, cfg.CatalogURL, 8083))
}

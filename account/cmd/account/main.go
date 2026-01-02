package main

import (
	"log"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/account"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Default for local development
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "postgres://viraj:123456@localhost:5433/viraj?sslmode=disable"
	}

	var r account.Repository
	for {
		r, err = account.NewPostgresRepository(cfg.DatabaseURL)
		if err == nil {
			break
		}
		log.Println(err)
		time.Sleep(2 * time.Second)
	}
	defer r.Close()

	log.Print("Listing on port 8081...")
	s := account.NewService(r)
	log.Fatal(account.ListenGRPC(s, 8081))
}

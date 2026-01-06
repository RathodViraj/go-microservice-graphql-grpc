package main

import (
	"log"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/inventory"
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
		cfg.DatabaseURL = "localhost:6379"
	}

	var r inventory.Repository
	for {
		r, err = inventory.NewRepository(cfg.DatabaseURL)
		if err == nil {
			break
		}
		log.Println(err)
		time.Sleep(2 * time.Second)
	}
	defer r.Close()

	log.Print("Listing on port 8084...")
	s := inventory.NewService(r)
	log.Fatal(inventory.ListenGRPC(s, 8084))
}

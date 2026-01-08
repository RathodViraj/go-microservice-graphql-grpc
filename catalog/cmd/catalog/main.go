package main

import (
	"log"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/catalog"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseURL  string `envconfig:"DATABASE_URL"`
	InventoryURL string `envconfig:"INVENTORY_URL"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Default for local development
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "http://localhost:9200"
	}
	if cfg.InventoryURL == "" {
		cfg.InventoryURL = "http://localhost:8084"
	}

	var r catalog.Repository
	for {
		r, err = catalog.NewElasticRepository(cfg.DatabaseURL)
		if err == nil {
			break
		}
		log.Println(err)
		time.Sleep(2 * time.Second)
	}
	defer r.Close()

	log.Println("Listening on port 8082...")
	s := catalog.NewSerivce(r)
	log.Fatal(catalog.ListenGRPC(s, cfg.InventoryURL, 8082))

}

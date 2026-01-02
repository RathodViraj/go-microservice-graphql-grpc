package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	AccountURL string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL string `envconfig:"CATALOG_SERVICE_URL"`
	OrderURL   string `envconfig:"ORDER_SERVICE_URL"`
}

func main() {
	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Defaults for local development
	if cfg.AccountURL == "" {
		cfg.AccountURL = "localhost:8081"
	}
	if cfg.CatalogURL == "" {
		cfg.CatalogURL = "localhost:8082"
	}
	if cfg.OrderURL == "" {
		cfg.OrderURL = "localhost:8083"
	}

	log.Printf("Starting GraphQL server with:")
	log.Printf("  Account service: %s", cfg.AccountURL)
	log.Printf("  Catalog service: %s", cfg.CatalogURL)
	log.Printf("  Order service: %s", cfg.OrderURL)

	s, err := NewGraphQLServer(cfg.AccountURL, cfg.OrderURL, cfg.CatalogURL)
	if err != nil {
		log.Fatalf("Failed to create GraphQL server: %v", err)
	}

	log.Println("GraphQL server initialized successfully")
	http.Handle("/graphql", handler.NewDefaultServer(s.ToExecutableSchema()))
	http.Handle("/playground", playground.Handler("viraj", "/graphql"))

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

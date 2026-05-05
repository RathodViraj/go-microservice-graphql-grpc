# microservice-graphql-grpc

Small demo microservices project implemented in Go using gRPC, Protobuf, and a GraphQL gateway.

Overview
- Multi-service example: `account`, `catalog`, `inventory`, `order`, and a `graphql` gateway.
- Each service exposes gRPC endpoints and has its own backing store:
  - `account`: PostgreSQL
  - `catalog`: Elasticsearch
  - `inventory`: Redis (simple key-value stock per product)
  - `order`: PostgreSQL and an orders_products join table
- The `graphql` service provides a GraphQL API that calls the other services via gRPC.

Tech stack
- Go (modules)
- gRPC + Protobuf for service-to-service RPC
- GraphQL via `gqlgen` for client-facing API
- Redis, PostgreSQL, Elasticsearch used in examples and tests

Running locally (quick)
1. Start services with Docker Compose:

```bash
docker compose up -d
```

2. Open the GraphQL playground in your browser:

- http://localhost:8080/playground

Notes
- The compose file maps service ports to the host (see `docker-compose.yml`).
- Integration / E2E tests may expect Postgres and Elasticsearch to be available; many unit tests use in-memory or fake gRPC servers so they run without external services.

GraphQL: example operations

Endpoint: `http://localhost:8080/graphql`

1) Create an account

Mutation
```graphql
mutation CreateAccount($a: AccountInput!) {
  createAccount(account: $a) {
    id
    name
  }
}
```

Variables
```json
{ "a": { "name": "Alice" } }
```

2) Create a product (catalog)

Mutation
```graphql
mutation CreateProduct($p: ProductInput!) {
  createProduct(product: $p) {
    id
    name
    price
  }
}
```

Variables
```json
{ "p": { "name": "Laptop", "description": "High-end laptop", "price": 1500.0 } }
```

3) Create an order

Mutation
```graphql
mutation CreateOrder($o: OrderInput!) {
  createOrder(order: $o) {
    id
    totalPrice
    products { id quantity name }
  }
}
```

Variables (replace IDs with values returned from previous mutations)
```json
{
  "o": {
    "accountId": "<ACCOUNT_ID>",
    "products": [{ "id": "<PRODUCT_ID>", "quantity": 2 }]
  }
}
```

4) Update stock (inventory)

Mutation
```graphql
mutation UpdateStock($r: UpdateStocksRequestInput!) {
  updateStock(requests: $r) { ids }
}
```

Variables
```json
{ "r": { "ids": ["<PRODUCT_ID>"], "deltas": [-2] } }
```

5) Check stock

Query
```graphql
query CheckStock($p: CheckStockInput!) {
  checkStock(pids: $p)
}
```

Variables
```json
{ "p": { "ids": ["<PRODUCT_ID>"] } }
```

Troubleshooting
- If tests fail due to missing external services, either start them with `docker compose up -d` or run the unit-test suites which use fakes/bufconn servers.
- CI uses service containers for Postgres, Elasticsearch and Redis; the workflow file is `.github/workflows/ci.yml`.

Contributing
- Open issues or PRs for fixes, improvements, or additional examples.

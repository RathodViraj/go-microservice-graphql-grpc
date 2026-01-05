package order

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockRepo(t *testing.T) (*postgresRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	repo := &postgresRepository{db: db}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestRepoUnit_PutOrder_Success(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	o := Order{
		ID:         "o1",
		CreatedAt:  time.Now(),
		AccountID:  "a1",
		TotalPrice: 100,
		Products: []OrderedProduct{
			{ID: "p1", Quantity: 2},
			{ID: "p2", Quantity: 1},
		},
	}

	mock.ExpectBegin()

	mock.ExpectExec(`INSERT INTO orders`).
		WithArgs(o.ID, o.CreatedAt, o.AccountID, o.TotalPrice).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// pq.CopyIn creates a COPY statement internally
	mock.ExpectPrepare(`COPY "orders_products"`).
		WillBeClosed()

	// one exec per product row
	mock.ExpectExec(`COPY "orders_products"`).
		WithArgs(o.ID, "p1", int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`COPY "orders_products"`).
		WithArgs(o.ID, "p2", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// final flush call: Exec() with no args
	mock.ExpectExec(`COPY "orders_products"`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectCommit()

	err := repo.PutOrder(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRepoUnit_PutProduct_RolsBackOnError(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	o := Order{
		ID: "o1", Products: []OrderedProduct{{ID: "p1", Quantity: 1}},
	}

	mock.ExpectBegin()

	mock.ExpectExec(`INSERT INTO orders`).
		WithArgs(o.ID, o.CreatedAt, o.AccountID, o.TotalPrice).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectPrepare(`COPY orders_products`)

	mock.ExpectExec(`COPY orders_products`).
		WithArgs(o.ID, "p1", int64(1)).
		WillReturnError(fmt.Errorf("copy failed"))

	mock.ExpectRollback()

	err := repo.PutOrder(context.Background(), o)
	if err == nil {
		t.Fatal("Expected error; got nil")
	}
}

func TestRepoUnit_GetOrderForAccounts(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "account_id", "total_price", "product_id", "quantity",
	}).
		AddRow("o1", time.Now(), "a1", 50.0, "p1", int64(2)).
		AddRow("o1", time.Now(), "a1", 50.0, "p2", int64(1)).
		AddRow("o2", time.Now(), "a1", 20.0, "p3", int64(1))

	mock.ExpectQuery(`FROM orders o`).
		WithArgs("a1").
		WillReturnRows(rows)

	orders, err := repo.GetOrderForAccount(context.Background(), "a1")
	if err != nil {
		t.Fatal(err)
	}

	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}

	if len(orders[0].Products) != 2 {
		t.Errorf("expected 2 products in o1, got %d", len(orders[0].Products))
	}
}

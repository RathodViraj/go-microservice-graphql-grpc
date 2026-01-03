package account

import (
	"database/sql"
	"testing"

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

func TestPutAccount(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow("u11", "viraj")

	mock.ExpectQuery(`SELECT id, name FROM accounts WHERE id = \$1`).
		WithArgs("u11").
		WillReturnRows(rows)

	acc, err := repo.GetAccountByID(t.Context(), "u11")
	if err != nil {
		t.Fatal(err)
	}

	if acc.ID != "u11" || acc.Name != "viraj" {
		t.Errorf("unexpected data: %#v", acc)
	}
}

func TestGetAccountById_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name FROM accounts WHERE id = \$1`).
		WithArgs("u21").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetAccountByID(t.Context(), "xyz")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestListAccounts(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("u32", "Alice").
		AddRow("u31", "Bob")

	mock.ExpectQuery(`SELECT id, name FROM accounts ORDER BY id DESC OFFSET \$1 LIMIT \$2`).
		WithArgs(uint64(0), uint64(2)).
		WillReturnRows(rows)

	accs, err := repo.ListAccounts(t.Context(), 0, 2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if len(accs) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accs))
	}
}

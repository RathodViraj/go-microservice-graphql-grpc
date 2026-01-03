package account

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockService(repo Repository) *accountService {
	return &accountService{repo: repo}
}

func Test_service_PutAccount(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()
	service := newMockService(repo)

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow("u11", "viraj")

	mock.ExpectQuery(`SELECT id, name FROM accounts WHERE id = \$1`).
		WithArgs("u11").
		WillReturnRows(rows)

	acc, err := service.GetAccount(t.Context(), "u11")
	if err != nil {
		t.Fatal(err)
	}
	if acc.ID != "u11" || acc.Name != "viraj" {
		t.Errorf("unexpected data: %#v", acc)
	}
}

func Test_service_ListAccount(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()
	service := newMockService(repo)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("u22", "Alice").
		AddRow("u21", "Bob")

	mock.ExpectQuery(`SELECT id, name FROM accounts ORDER BY id DESC OFFSET \$1 LIMIT \$2`).
		WithArgs(uint64(0), uint64(2)).
		WillReturnRows(rows)

	accs, err := service.GetAccounts(t.Context(), 0, 2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if len(accs) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accs))
	}
}

func Test_service_ListAccountWithSkip(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()
	service := newMockService(repo)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("u22", "Alice")

	mock.ExpectQuery(`SELECT id, name FROM accounts ORDER BY id DESC OFFSET \$1 LIMIT \$2`).
		WithArgs(uint64(1), uint64(100)).
		WillReturnRows(rows)

	accs, err := service.GetAccounts(t.Context(), 1, 200)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if len(accs) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accs))
	}
}

func Test_service_ListAccount_InvalidSkipTake(t *testing.T) {
	repo, mock, cleanup := newMockRepo(t)
	defer cleanup()
	service := newMockService(repo)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("u32", "Alice").
		AddRow("u31", "Bob")

	mock.ExpectQuery(`SELECT id, name FROM accounts ORDER BY id DESC OFFSET \$1 LIMIT \$2`).
		WithArgs(uint64(0), uint64(100)).
		WillReturnRows(rows)

	accs, err := service.GetAccounts(t.Context(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if len(accs) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accs))
	}
}

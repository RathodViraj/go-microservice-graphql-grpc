package account

import (
	"context"
	"os"
	"testing"

	"github.com/segmentio/ksuid"
)

var testRepo Repository

func TestMain(m *testing.M) {
	url := os.Getenv("DATABASE_URL_FOR_TEST")

	r, err := NewPostgresRepository(url)
	if err != nil {
		panic(err)
	}

	testRepo = r
	code := m.Run()
	r.Close()
	os.Exit(code)
}

func TestRepo_PutAndGetAccount(t *testing.T) {
	ctx := context.Background()

	a := Account{ID: ksuid.New().String(), Name: "Viraj"}
	if err := testRepo.PutAccount(ctx, a); err != nil {
		t.Fatal(err)
	}

	got, err := testRepo.GetAccountByID(ctx, a.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != a.ID && got.Name != a.Name {
		t.Errorf("unexpected account: %#v", got)
	}
}

func TestRepo_GetAccount_NotFound(t *testing.T) {
	ctx := context.Background()
	_, err := testRepo.GetAccountByID(ctx, "non-existing-id")
	if err != ErrAccountNotFound {
		t.Fatalf("expected ErrAccountNotFound; got %v", err)
	}
}

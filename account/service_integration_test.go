package account

import (
	"context"
	"testing"
)

func TestService_PutAndGetAccount_Success(t *testing.T) {
	ctx := context.Background()

	svc := &accountService{
		repo: testRepo,
	}

	acc, err := svc.PostAccount(ctx, "viraj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dbAcc, err := svc.GetAccount(ctx, acc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dbAcc.Name != "viraj" {
		t.Errorf("unexpected account name: %s", dbAcc.Name)
	}
}

func TestService_GetAccount_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := &accountService{
		repo: testRepo,
	}
	_, err := svc.GetAccount(ctx, "non-existing-id")
	if err != ErrAccountNotFound {
		t.Fatalf("expected ErrAccountNotFound; got %v", err)
	}
}

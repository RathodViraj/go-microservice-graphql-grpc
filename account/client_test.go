package account

import (
	"context"
	"testing"

	"github.com/RathodViraj/go-microservice-graphql-grpc/account/pb"
	"go.uber.org/mock/gomock"
)

func TestClinet_PostAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPB := NewMockAccountServiceClient(ctrl)

	mockPB.EXPECT().
		PostAccount(
			gomock.Any(),
			&pb.PostAccountRequest{Name: "viraj"},
		).
		Return(&pb.PostAccountResponse{
			Account: &pb.Account{
				Id:   "any-id",
				Name: "viraj",
			},
		}, nil)

	c := &Client{
		conn:    nil,
		service: mockPB,
	}

	res, err := c.PostAccount(context.Background(), "viraj")
	if err != nil {
		t.Fatal(err)
	}

	if res.ID == "" {
		t.Error("expected not-nil ID")
	}
	if res.Name != "viraj" {
		t.Errorf("unexpected name in response: %s", res.Name)
	}
}

func TestClient_GetAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPB := NewMockAccountServiceClient(ctrl)

	mockPB.EXPECT().
		GetAccount(gomock.Any(), &pb.GetAccountRequest{Id: "u1"}).
		Return(&pb.GetAccountResponse{
			Account: &pb.Account{Id: "u1", Name: "Viraj"},
		}, nil)

	c := &Client{service: mockPB}

	acc, err := c.GetAccount(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acc.ID != "u1" {
		t.Errorf("wrong id: %s", acc.ID)
	}
}

func TestClient_GetAccounts_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPB := NewMockAccountServiceClient(ctrl)

	mockPB.EXPECT().
		GetAccounts(gomock.Any(), &pb.GetAccountsRequest{Skip: 0, Take: 2}).
		Return(&pb.GetAccountsResponse{
			Accounts: []*pb.Account{
				{Id: "u1", Name: "Viraj"},
				{Id: "u2", Name: "Alice"},
			},
		}, nil)

	c := &Client{service: mockPB}

	res, err := c.GetAccounts(context.Background(), 0, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(res))
	}
}

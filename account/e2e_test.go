package account

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/RathodViraj/go-microservice-graphql-grpc/account/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartE2EServer(t *testing.T) (string, func()) {
	url := os.Getenv("DATABASE_URL_FOR_TEST")
	if url == "" {
		t.Fatal("DATABASE_URL_FOR_TEST not set")
	}

	repo, err := NewPostgresRepository(url)
	if err != nil {
		t.Fatalf("failed to connect repo: %v", err)
	}

	svc := &accountService{repo: repo}

	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAccountServiceServer(s, &grpcServer{service: svc})

	go s.Serve(lis)

	cleanup := func() {
		s.Stop()
		repo.Close()
		lis.Close()
	}

	return "localhost:9091", cleanup
}

func TestE2E_PostAndGetAccount(t *testing.T) {
	addr, cleanup := StartE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewAccountServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	postRes, err := client.PostAccount(ctx, &pb.PostAccountRequest{Name: "Viraj"})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}

	if postRes.Account.Id == "" {
		t.Fatalf("expected generated id")
	}

	getRes, err := client.GetAccount(ctx, &pb.GetAccountRequest{Id: postRes.Account.Id})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if getRes.Account.Name != "Viraj" {
		t.Errorf("unexpected name, got %s", getRes.Account.Name)
	}
}

func TestE2E_ListAccounts(t *testing.T) {
	addr, cleanup := StartE2EServer(t)
	defer cleanup()

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewAccountServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	_, err = client.PostAccount(
		ctx,
		&pb.PostAccountRequest{Name: "alice"},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.PostAccount(
		ctx,
		&pb.PostAccountRequest{Name: "bob"},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.PostAccount(
		ctx,
		&pb.PostAccountRequest{Name: "peter"},
	)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.GetAccounts(ctx, &pb.GetAccountsRequest{Skip: 1, Take: 2})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Accounts) != 2 {
		t.Errorf("expected 2 accounts got %d", len(res.Accounts))
	}
}

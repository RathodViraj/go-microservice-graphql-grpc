package account

import (
	"context"
	"net"
	"testing"

	"github.com/RathodViraj/go-microservice-graphql-grpc/account/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func startIntegrationServer(t *testing.T) (*grpc.ClientConn, func()) {
	svc := &accountService{repo: testRepo}

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	pb.RegisterAccountServiceServer(s, &grpcServer{service: svc})

	go s.Serve(lis)

	conn, _ := grpc.DialContext(
		context.Background(),
		"buf",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)

	cleanup := func() {
		conn.Close()
		s.Stop()
	}

	return conn, cleanup
}

func TestGRPC_PostAccount_Integration(t *testing.T) {
	conn, cleanup := startIntegrationServer(t)
	defer cleanup()

	client := pb.NewAccountServiceClient(conn)

	res, err := client.PostAccount(
		context.Background(),
		&pb.PostAccountRequest{Name: "viraj"},
	)
	if err != nil {
		t.Fatal(err)
	}

	if res.Account.Name != "viraj" {
		t.Errorf("expected Bob, got %s", res.Account.Name)
	}
}

func TestGRPC_ListAccounts_Integration(t *testing.T) {
	conn, cleanup := startIntegrationServer(t)
	defer cleanup()

	client := pb.NewAccountServiceClient(conn)

	_, err := client.PostAccount(
		context.Background(),
		&pb.PostAccountRequest{Name: "alice"},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.PostAccount(
		context.Background(),
		&pb.PostAccountRequest{Name: "bob"},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.PostAccount(
		context.Background(),
		&pb.PostAccountRequest{Name: "peter"},
	)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.GetAccounts(context.Background(), &pb.GetAccountsRequest{Skip: 1, Take: 2})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Accounts) != 2 {
		t.Errorf("expected 2 accounts got %d", len(res.Accounts))
	}
}

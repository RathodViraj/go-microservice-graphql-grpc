package account

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/RathodViraj/go-microservice-graphql-grpc/account/pb"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func startTestServer(t *testing.T, svc Service) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)

	s := grpc.NewServer()
	pb.RegisterAccountServiceServer(s, &grpcServer{service: svc})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("server exited: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		conn.Close()
		s.Stop()
	}

	return conn, cleanup
}

func TestPostAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		PostAccount(gomock.Any(), "viraj").
		Return(&Account{ID: "test-id", Name: "viraj"}, nil)

	conn, cleanup := startTestServer(t, mockSvc)
	defer cleanup()

	client := pb.NewAccountServiceClient(conn)
	res, err := client.PostAccount(
		context.Background(),
		&pb.PostAccountRequest{Name: "viraj"},
	)
	if err != nil {
		t.Fatal("err")
	}

	if res.Account.Id == "" {
		t.Error("expected non-empty ID in response")
	}
	if res.Account.Name != "viraj" {
		t.Errorf("unexpected name in response: %s", res.Account.Name)
	}
}

func TestGetAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		PostAccount(gomock.Any(), "viraj").
		Return(&Account{ID: "any-id", Name: "viraj"}, nil)

	mockSvc.EXPECT().
		GetAccount(gomock.Any(), "any-id").
		Return(&Account{ID: "any-id", Name: "viraj"}, nil)

	conn, cleanup := startTestServer(t, mockSvc)
	defer cleanup()

	client := pb.NewAccountServiceClient(conn)
	acc, err := client.PostAccount(
		context.Background(),
		&pb.PostAccountRequest{Name: "viraj"},
	)
	if err != nil {
		t.Fatal("err")
	}

	res, err := client.GetAccount(context.Background(), &pb.GetAccountRequest{Id: acc.Account.Id})
	if err != nil {
		log.Fatal(err)
	}

	if res.Account.Id != acc.Account.Id || res.Account.Name != acc.Account.Name {
		t.Errorf("unexpected response: %#v", res.Account)
	}
}

func TestGetAccount_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)

	mockSvc.EXPECT().
		GetAccount(gomock.Any(), "fail").
		Return(nil, fmt.Errorf("account not found"))

	conn, cleanup := startTestServer(t, mockSvc)
	defer cleanup()

	client := pb.NewAccountServiceClient(conn)

	_, err := client.GetAccount(context.Background(), &pb.GetAccountRequest{Id: "fail"})
	if err == nil {
		log.Fatal("expected error; got nil.")
	}
}

package inventory

import (
	"context"
	"fmt"

	"github.com/RathodViraj/go-microservice-graphql-grpc/inventory/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Conn    *grpc.ClientConn
	Service pb.InventoryServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect inventory client: %v", err)
	}

	c := pb.NewInventoryServiceClient(conn)
	return &Client{conn, c}, nil
}

func (c *Client) Close() {
	c.Conn.Close()
}

func (c *Client) UpdateStock(ctx context.Context, pids []string, deltas []int32) ([]string, error) {
	res, err := c.Service.UpdateStock(
		ctx,
		&pb.UpdateStockRequest{
			Pids:   pids,
			Deltas: deltas,
		},
	)
	if err != nil {
		return nil, err
	}

	return res.OutOfStock, nil
}

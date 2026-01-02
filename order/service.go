package order

import (
	"context"
	"time"

	"github.com/segmentio/ksuid"
)

type Order struct {
	ID         string
	CreatedAt  time.Time
	TotalPrice float64
	AccountID  string
	Products   []OrderedProduct
}

type OrderedProduct struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Quantity    uint32
}

type Service interface {
	PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error)
	GetOrderForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type orderService struct {
	repository Repository
}

func NewOrderService(r Repository) Service {
	return &orderService{r}
}

func (s *orderService) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	var totalPrice float64 = 0
	for _, p := range products {
		totalPrice += (p.Price * float64(p.Quantity))
	}
	o := &Order{
		ID:         ksuid.New().String(),
		CreatedAt:  time.Now().UTC(),
		TotalPrice: totalPrice,
		AccountID:  accountID,
		Products:   products,
	}
	if err := s.repository.PutOrder(ctx, *o); err != nil {
		return nil, err
	}

	return o, nil
}

func (s *orderService) GetOrderForAccount(ctx context.Context, accountID string) ([]Order, error) {
	return s.repository.GetOrderForAccount(ctx, accountID)
}

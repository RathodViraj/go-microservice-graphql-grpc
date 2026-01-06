package inventory

import (
	"context"
	"fmt"
)

type Stock struct {
	Product_id string
	Delta      int32
}

type Service interface {
	UpdateStock(ctx context.Context, pids []string, deltas []int32) ([]string, error)
}

type inventoryService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &inventoryService{repo}
}

func (s *inventoryService) UpdateStock(ctx context.Context, pids []string, deltas []int32) ([]string, error) {
	if len(pids) == 0 || len(pids) != len(deltas) {
		return nil, fmt.Errorf("invalid input: pids:%d, deltas:%d", len(pids), len(deltas))
	}

	var requests []Stock
	for i := range len(pids) {
		if deltas[i] == 0 {
			continue
		}
		requests = append(
			requests,
			Stock{
				Product_id: pids[i],
				Delta:      deltas[i],
			},
		)
	}
	res, err := s.repo.UpdateStock(ctx, requests)
	if err != nil {
		return nil, err
	}

	if len(res) > len(pids) {
		return nil, fmt.Errorf("something went horribly wrong: pids:%d, oosItems:%d", len(pids), len(res))
	}
	return res, nil
}

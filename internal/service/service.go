package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
	"wb-l0/internal/domain"
	"wb-l0/pkg/e"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go
type DB interface {
	GetByUID(ctx context.Context, uid int) (domain.Order, error)
	CreateOrder(ctx context.Context, order domain.Order) (int, error)
	GetAllOrderIDs(ctx context.Context) ([]domain.Order, error)
}

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, exp time.Duration) error
	Get(ctx context.Context, key string, value *domain.Order) (string, error)
}

type Service struct {
	db     DB
	cache  Cache
	logger *slog.Logger
}

func NewService(logger *slog.Logger, db DB, cache Cache) *Service {
	return &Service{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (s *Service) GetByUID(ctx context.Context, uid int) (domain.Order, error) {
	order, err := s.db.GetByUID(ctx, uid)
	if err != nil {
		s.logger.Error("failed to perform GetByUID", slog.String("error", err.Error()))
		return domain.Order{}, e.Wrap("service.GetOrder", err)
	}

	return order, nil
}

func (s *Service) CreateOrder(ctx context.Context, order domain.Order) (int, error) {
	id, err := s.db.CreateOrder(ctx, order)
	if err != nil {
		s.logger.Error("The CreateOrder failed in service layer", slog.String("error", err.Error()))
		return 0, e.Wrap("service.CreateOrder", err)
	}

	return id, nil
}

func (s *Service) GetOrderJSON(ctx context.Context, order domain.Order) (string, error) {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return "", e.Wrap("service.GetOrderJSON", err)
	}

	return string(orderJSON), nil
}

func (s *Service) Set(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	return s.cache.Set(ctx, key, value, exp)

}
func (s *Service) Get(ctx context.Context, key string, value *domain.Order) (string, error) {
	return s.cache.Get(ctx, key, value)
}

func (s *Service) GetAllOrderIDs(ctx context.Context) ([]domain.Order, error) {
	return s.db.GetAllOrderIDs(ctx)
}

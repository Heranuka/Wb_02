package service

import (
	"context"
	"log/slog"
	"test_18/internal/domain"
	"test_18/pkg/e.go"
	"time"
)

type DB interface {
	CreateEvent(ctx context.Context, event domain.Event) (int, error)
	UpdateEvent(ctx context.Context, id int, req domain.UpdateEventRequest) error
	DeleteEvent(ctx context.Context, id int) error
	GetEventsForTime(ctx context.Context, startDate time.Time, endDate time.Time) ([]domain.Event, error)
}

type Service struct {
	logger *slog.Logger
	db     DB
}

func NewService(logger *slog.Logger, db DB) *Service {
	return &Service{
		logger: logger,
		db:     db,
	}
}

func (s *Service) CreateEvent(ctx context.Context, event domain.Event) (int, error) {
	return s.db.CreateEvent(ctx, event)
}
func (s *Service) UpdateEvent(ctx context.Context, id int, req domain.UpdateEventRequest) error {
	return s.db.UpdateEvent(ctx, id, req)
}
func (s *Service) DeleteEvent(ctx context.Context, id int) error {
	return s.db.DeleteEvent(ctx, id)
}
func (s *Service) GetEventsForTime(ctx context.Context, startDate time.Time, endDate time.Time) ([]domain.Event, error) {
	events, err := s.db.GetEventsForTime(ctx, startDate, endDate)
	if err != nil {
		return nil, e.Wrap("service.GetEventsForTime", err)
	}
	s.logger.Error(" GetEventsForDay", slog.Any("events", events))
	return events, nil
}

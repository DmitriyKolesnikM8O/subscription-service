package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/entity"
	"github.com/DmitriyKolesnikM8O/subscription-service/internal/repo"
	"github.com/DmitriyKolesnikM8O/subscription-service/internal/repo/repoerrs"
	"github.com/google/uuid"
)

type subscriptionService struct {
	repos *repo.Repositories
}

func NewSubscriptionService(repos *repo.Repositories) SubscriptionService {
	return &subscriptionService{repos: repos}
}

func (s *subscriptionService) CreateSubscription(
	ctx context.Context,
	sub entity.Subscription,
) (*entity.Subscription, error) {
	if sub.Service.Name == "" {
		return nil, fmt.Errorf("SubscriptionService.CreateSubscription - empty service name")
	}
	if sub.Service.Price <= 0 {
		return nil, fmt.Errorf("SubscriptionService.CreateSubscription - price must be positive")
	}
	if sub.StartDate.IsZero() {
		sub.StartDate = time.Now().UTC()
	}
	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate) {
		return nil, fmt.Errorf("SubscriptionService.CreateSubscription - end date before start date")
	}

	createdSub, err := s.repos.Subscription.CreateSubscription(ctx, sub)
	if err != nil {
		if errors.Is(err, repoerrs.ErrAlreadyExists) {
			return nil, fmt.Errorf("SubscriptionService.CreateSubscription - %w", err)
		}
		return nil, fmt.Errorf("SubscriptionService.CreateSubscription - repo error: %v", err)
	}

	return createdSub, nil
}

func (s *subscriptionService) GetSubscriptionByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Subscription, error) {
	sub, err := s.repos.Subscription.GetSubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return nil, fmt.Errorf("SubscriptionService.GetSubscriptionByID - %w", err)
		}
		return nil, fmt.Errorf("SubscriptionService.GetSubscriptionByID - repo error: %v", err)
	}
	return &sub, nil
}

func (s *subscriptionService) UpdateSubscription(
	ctx context.Context,
	id uuid.UUID,
	sub entity.Subscription,
) error {
	if sub.Service.Name == "" {
		return fmt.Errorf("SubscriptionService.UpdateSubscription - empty service name")
	}
	if sub.Service.Price <= 0 {
		return fmt.Errorf("SubscriptionService.UpdateSubscription - price must be positive")
	}

	current, err := s.repos.Subscription.GetSubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return fmt.Errorf("SubscriptionService.UpdateSubscription - %w", err)
		}
		return fmt.Errorf("SubscriptionService.UpdateSubscription - get sub error: %v", err)
	}

	if sub.EndDate != nil && sub.EndDate.Before(current.StartDate) {
		return fmt.Errorf("SubscriptionService.UpdateSubscription - end date before start date")
	}

	current.Service.Name = sub.Service.Name
	current.Service.Price = sub.Service.Price
	current.EndDate = sub.EndDate

	if err := s.repos.Subscription.UpdateSubscription(ctx, current); err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return fmt.Errorf("SubscriptionService.UpdateSubscription - %w", err)
		}
		return fmt.Errorf("SubscriptionService.UpdateSubscription - repo error: %v", err)
	}

	return nil
}

func (s *subscriptionService) DeleteSubscription(
	ctx context.Context,
	id uuid.UUID,
) error {
	if err := s.repos.Subscription.DeleteSubscription(ctx, id); err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return fmt.Errorf("SubscriptionService.DeleteSubscription - %w", err)
		}
		return fmt.Errorf("SubscriptionService.DeleteSubscription - repo error: %v", err)
	}
	return nil
}

func (s *subscriptionService) ListSubscriptionsByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]entity.Subscription, error) {
	subs, err := s.repos.Subscription.ListSubscriptions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("SubscriptionService.ListSubscriptionsByUser - repo error: %v", err)
	}
	return subs, nil
}

func (s *subscriptionService) CalculateTotalCost(
	ctx context.Context,
	userID *uuid.UUID, // Меняем на указатель
	serviceName *string, // Теперь это указатель на строку
	startDate, endDate time.Time,
) (int, error) {
	if startDate.After(endDate) {
		return 0, fmt.Errorf("invalid date range")
	}

	total, err := s.repos.Report.GetTotalCost(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("service error: %w", err)
	}

	return total, nil
}

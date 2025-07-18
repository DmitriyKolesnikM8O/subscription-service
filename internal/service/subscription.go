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
	serviceName string,
	price int,
	userID uuid.UUID,
	startDate time.Time,
	endDate *time.Time,
) (*entity.Subscription, error) {

	if serviceName == "" {
		return nil, fmt.Errorf("SubscriptionService.CreateSubscription - empty service name")
	}
	if price <= 0 {
		return nil, fmt.Errorf("SubscriptionService.CreateSubscription - price must be positive")
	}
	if startDate.IsZero() {
		startDate = time.Now().UTC()
	}
	if endDate != nil && endDate.Before(startDate) {
		return nil, fmt.Errorf("SubscriptionService.CreateSubscription - end date before start date")
	}

	sub := entity.Subscription{
		ID:          uuid.New(),
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	id, err := s.repos.Subscription.CreateSubscription(ctx, sub)
	if err != nil {
		if errors.Is(err, repoerrs.ErrAlreadyExists) {
			return nil, fmt.Errorf("SubscriptionService.CreateSubscription - %w", err)
		}
		return nil, fmt.Errorf("SubscriptionService.CreateSubscription - repo error: %v", err)
	}

	sub.ID = id
	return &sub, nil
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
	serviceName string,
	price int,
	endDate *time.Time,
) error {
	if serviceName == "" {
		return fmt.Errorf("SubscriptionService.UpdateSubscription - empty service name")
	}
	if price <= 0 {
		return fmt.Errorf("SubscriptionService.UpdateSubscription - price must be positive")
	}

	current, err := s.repos.Subscription.GetSubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return fmt.Errorf("SubscriptionService.UpdateSubscription - %w", err)
		}
		return fmt.Errorf("SubscriptionService.UpdateSubscription - get sub error: %v", err)
	}

	if endDate != nil && endDate.Before(current.StartDate) {
		return fmt.Errorf("SubscriptionService.UpdateSubscription - end date before start date")
	}

	current.ServiceName = serviceName
	current.Price = price
	current.EndDate = endDate

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
	serviceName string, // Оставляем строку (преобразуем в nil при пустом значении)
	startDate, endDate time.Time,
) (int, error) {
	if startDate.After(endDate) {
		return 0, fmt.Errorf("invalid date range")
	}

	var serviceNamePtr *string
	if serviceName != "" {
		serviceNamePtr = &serviceName
	}

	total, err := s.repos.Report.GetTotalCost(ctx, userID, serviceNamePtr, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("service error: %w", err)
	}

	return total, nil
}

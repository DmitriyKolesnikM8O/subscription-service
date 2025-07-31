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
	page int,
	limit int,
) ([]entity.Subscription, int, error) {

	offset := (page - 1) * limit
	subs, err := s.repos.Subscription.ListSubscriptions(ctx, userID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("SubscriptionService.ListSubscriptionsByUser - repo error: %v", err)
	}
	total, err := s.repos.Subscription.GetTotalByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("SubscriptionService.ListSubscriptionsByUser - failed to get total: %v", err)
	}
	return subs, total, nil
}

func (s *subscriptionService) CalculateTotalCost(
	ctx context.Context,
	userID *uuid.UUID,
	serviceName *string,
	startDate, endDate time.Time,
) (int, error) {

	subscriptions, err := s.repos.Report.GetTotalCost(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("service error: %w", err)
	}

	monthMap := make(map[string]bool)
	totalCost := 0

	for _, data := range subscriptions {

		currentStart := data.StartDate
		if currentStart.Before(startDate) {
			currentStart = startDate
		}
		currentEnd := data.EndDate
		if currentEnd != nil && currentEnd.After(endDate) {
			currentEnd = &endDate
		}
		if currentEnd == nil || currentEnd.After(endDate) {
			currentEnd = &endDate
		}
		if currentEnd != nil && currentEnd.Before(currentStart) {
			continue
		}

		current := currentStart
		for !current.After(*currentEnd) {
			monthKey := current.Format("2006-01")
			if !monthMap[monthKey] {
				monthMap[monthKey] = true
				totalCost += data.Price
			}
			current = current.AddDate(0, 1, 0)
		}
	}

	return totalCost, nil
}

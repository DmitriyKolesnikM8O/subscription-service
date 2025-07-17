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
	// Валидация
	if serviceName == "" {
		return nil, errors.New("название сервиса обязательно")
	}
	if price <= 0 {
		return nil, errors.New("цена должна быть положительной")
	}
	if startDate.IsZero() {
		startDate = time.Now().UTC()
	}
	if endDate != nil && endDate.Before(startDate) {
		return nil, errors.New("дата окончания не может быть раньше даты начала")
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
		return nil, fmt.Errorf("failed to create subscription: %w", err)
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
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
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
	// Получаем текущую подписку
	current, err := s.repos.Subscription.GetSubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Валидация
	if serviceName == "" {
		return errors.New("название сервиса обязательно")
	}
	if price <= 0 {
		return errors.New("цена должна быть положительной")
	}
	if endDate != nil && endDate.Before(current.StartDate) {
		return errors.New("дата окончания не может быть раньше даты начала")
	}

	// Обновляем поля
	current.ServiceName = serviceName
	current.Price = price
	current.EndDate = endDate

	if err := s.repos.Subscription.UpdateSubscription(ctx, current); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (s *subscriptionService) DeleteSubscription(
	ctx context.Context,
	id uuid.UUID,
) error {
	if err := s.repos.Subscription.DeleteSubscription(ctx, id); err != nil {
		if errors.Is(err, repoerrs.ErrNotFound) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

func (s *subscriptionService) ListSubscriptionsByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]entity.Subscription, error) {
	subs, err := s.repos.Subscription.ListSubscriptions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	return subs, nil
}

func (s *subscriptionService) CalculateTotalCost(
	ctx context.Context,
	userID uuid.UUID,
	serviceName string,
	startDate, endDate time.Time,
) (int, error) {
	// Валидация периода
	if startDate.After(endDate) {
		return 0, errors.New("неверный временной период")
	}

	total, err := s.repos.Report.GetTotalCost(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	return total, nil
}

// Дополнительные ошибки сервиса
var (
	ErrSubscriptionNotFound = errors.New("подписка не найдена")
)

package v1

import (
	"net/http"
	"time"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/DmitriyKolesnikM8O/subscription-service/internal/entity"
)

type SubscriptionController struct {
	service service.SubscriptionService
}

func NewSubscriptionController(s service.SubscriptionService) *SubscriptionController {
	return &SubscriptionController{service: s}
}

// Create godoc
// @Summary Создать подписку
// @Description Создает новую подписку пользователя
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Данные подписки"
// @Success 201 {object} entity.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [post]
func (c *SubscriptionController) Create(ctx echo.Context) error {
	var req CreateRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrBadRequest)
	}

	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ValidationError(err))
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrInvalidDateFormat)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrInvalidUserID)
	}

	sub := entity.Subscription{
		Service: entity.Service{
			Name:  req.Service.Name,
			Price: req.Service.Price,
		},
		UserID:    userID,
		StartDate: startDate,
	}

	sub, err = c.service.CreateSubscription(
		ctx.Request().Context(),
		sub,
	)
	if err != nil {
		return HTTPError(err)
	}

	return ctx.JSON(http.StatusCreated, sub)
}

// GetByID godoc
// @Summary Получить подписку по ID
// @Description Возвращает подписку по её идентификатору
// @Tags Subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} entity.Subscription
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (c *SubscriptionController) GetByID(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrInvalidSubscription)
	}

	sub, err := c.service.GetSubscriptionByID(ctx.Request().Context(), id)
	if err != nil {
		return HTTPError(err)
	}

	return ctx.JSON(http.StatusOK, sub)
}

// Update godoc
// @Summary Обновить подписку
// @Description Обновляет данные существующей подписки
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param request body UpdateRequest true "Данные для обновления"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [put]
func (c *SubscriptionController) Update(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrInvalidSubscription)
	}

	var req UpdateRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrBadRequest)
	}

	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ValidationError(err))
	}

	var endDate *time.Time
	if req.EndDate != "" {
		parsedDate, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, ErrInvalidDateFormat)
		}
		endDate = &parsedDate
	}

	sub := entity.Subscription{
		Service: entity.Service{
			Name:  req.Service.Name,
			Price: req.Service.Price,
		},
		EndDate: endDate,
	}

	err = c.service.UpdateSubscription(
		ctx.Request().Context(),
		id,
		sub,
	)
	if err != nil {
		return HTTPError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// Delete godoc
// @Summary Удалить подписку
// @Description Удаляет подписку по её идентификатору
// @Tags Subscriptions
// @Param id path string true "ID подписки"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [delete]
func (c *SubscriptionController) Delete(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrInvalidSubscription)
	}

	err = c.service.DeleteSubscription(ctx.Request().Context(), id)
	if err != nil {
		return HTTPError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ListByUser godoc
// @Summary Список подписок пользователя
// @Description Возвращает все подписки указанного пользователя
// @Tags Subscriptions
// @Produce json
// @Param user_id query string true "ID пользователя"
// @Success 200 {array} entity.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [get]
func (c *SubscriptionController) ListByUser(ctx echo.Context) error {
	userID, err := uuid.Parse(ctx.QueryParam("user_id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrInvalidUserID)
	}

	subscriptions, err := c.service.ListSubscriptionsByUser(ctx.Request().Context(), userID)
	if err != nil {
		return HTTPError(err)
	}

	return ctx.JSON(http.StatusOK, subscriptions)
}

// CalculateTotalCost godoc
// @Summary Расчет стоимости подписок
// @Description Возвращает суммарную стоимость подписок за период
// @Tags Subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param start_date query string true "Начало периода (MM-YYYY)"
// @Param end_date query string true "Конец периода (MM-YYYY)"
// @Success 200 {object} TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/total-cost [get]
func (c *SubscriptionController) CalculateTotalCost(ctx echo.Context) error {
	req := CalculateTotalCostRequest{
		UserID:      ctx.QueryParam("user_id"),
		ServiceName: ctx.QueryParam("service_name"),
		StartDate:   ctx.QueryParam("start_date"),
		EndDate:     ctx.QueryParam("end_date"),
	}

	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ValidationError(err))
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrInvalidDateFormat)
	}
	endDate, err := time.Parse("01-2006", req.EndDate)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrInvalidDateFormat)
	}

	var userID *uuid.UUID
	if req.UserID != "" {
		parsedUUID, err := uuid.Parse(req.UserID)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, ErrInvalidUserID)
		}
		userID = &parsedUUID
	}

	var serviceName *string
	if req.ServiceName != "" {
		serviceName = &req.ServiceName
	}

	total, err := c.service.CalculateTotalCost(
		ctx.Request().Context(),
		userID,
		serviceName,
		startDate,
		endDate,
	)
	if err != nil {
		return HTTPError(err)
	}

	return ctx.JSON(http.StatusOK, TotalCostResponse{Total: total})
}

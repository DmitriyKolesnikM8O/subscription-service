package v1

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/entity"
	"github.com/DmitriyKolesnikM8O/subscription-service/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type SubscriptionController struct {
	service service.SubscriptionService
	logger  *log.Logger
}

func NewSubscriptionController(s service.SubscriptionService, logger *log.Logger) *SubscriptionController {
	if logger == nil {
		logger = log.New()
		logger.SetReportCaller(true)
	}
	return &SubscriptionController{service: s, logger: logger}
}

func (c *SubscriptionController) logRequest(method, path string) {
	c.logger.WithFields(log.Fields{
		"method": method,
		"path":   path,
		"time":   time.Now().UTC().Format(time.RFC3339),
	}).Info("request started")
}

func (c *SubscriptionController) logSuccess(action string, fields log.Fields) {
	if fields == nil {
		fields = log.Fields{}
	}
	fields["time"] = time.Now().UTC().Format(time.RFC3339)
	c.logger.WithFields(fields).Infof("%s completed successfully", action)
}

func (c *SubscriptionController) logError(action string, err error, fields log.Fields) {
	if fields == nil {
		fields = log.Fields{}
	}
	fields["error"] = err.Error()
	fields["time"] = time.Now().UTC().Format(time.RFC3339)
	c.logger.WithFields(fields).Errorf("failed to %s", action)
}

func hashString(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
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
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [post]
func (c *SubscriptionController) Create(ctx echo.Context) error {
	c.logRequest(ctx.Request().Method, ctx.Request().URL.Path)

	var req CreateRequest
	if err := ctx.Bind(&req); err != nil {
		c.logError("bind request", err, log.Fields{
			"request_body": fmt.Sprintf("%+v", req),
		})
		return ctx.JSON(http.StatusBadRequest, ErrBadRequest)
	}

	if err := ctx.Validate(req); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string]string)
			for _, e := range validationErrs {
				errors[e.Field()] = e.Tag()
			}
			c.logError("validate request", err, log.Fields{
				"validation_errors": errors,
			})
		} else {
			c.logError("validate request", err, nil)
		}
		return ctx.JSON(http.StatusBadRequest, ValidationError(err))
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		c.logError("parse start date", err, log.Fields{
			"start_date": req.StartDate,
		})
		return ctx.JSON(http.StatusBadRequest, ErrInvalidDateFormat)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.logError("parse user ID", err, log.Fields{
			"user_id_hash": hashString(req.UserID),
		})
		return ctx.JSON(http.StatusBadRequest, ErrInvalidUserID)
	}

	sub := &entity.Subscription{
		Service: entity.Service{
			Name:  req.Service.Name,
			Price: req.Service.Price,
		},
		UserID:    userID,
		StartDate: startDate,
	}

	sub, err = c.service.CreateSubscription(ctx.Request().Context(), *sub)
	if err != nil {
		c.logError("create subscription", err, log.Fields{
			"subscription": fmt.Sprintf("%+v", sub),
			"user_id_hash": hashString(userID.String()),
		})
		return HTTPError(err)
	}

	c.logSuccess("create subscription", log.Fields{
		"subscription_id": sub.ID,
		"user_id_hash":    hashString(userID.String()),
	})
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
	c.logRequest(ctx.Request().Method, ctx.Request().URL.Path)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		c.logError("parse subscription ID", err, log.Fields{
			"input_id": ctx.Param("id"),
		})
		return ctx.JSON(http.StatusBadRequest, ErrInvalidSubscription)
	}

	sub, err := c.service.GetSubscriptionByID(ctx.Request().Context(), id)
	if err != nil {
		c.logError("get subscription by ID", err, log.Fields{
			"subscription_id": id,
		})
		return HTTPError(err)
	}

	c.logSuccess("get subscription by ID", log.Fields{
		"subscription_id": id,
	})
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
	c.logRequest(ctx.Request().Method, ctx.Request().URL.Path)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		c.logError("parse subscription ID", err, log.Fields{
			"input_id": ctx.Param("id"),
		})
		return ctx.JSON(http.StatusBadRequest, ErrInvalidSubscription)
	}

	var req UpdateRequest
	if err := ctx.Bind(&req); err != nil {
		c.logError("bind request", err, log.Fields{
			"request_body": fmt.Sprintf("%+v", req),
		})
		return ctx.JSON(http.StatusBadRequest, ErrBadRequest)
	}

	if err := ctx.Validate(req); err != nil {
		c.logError("validate request", err, nil)
		return ctx.JSON(http.StatusBadRequest, ValidationError(err))
	}

	var endDate *time.Time
	if req.EndDate != "" {
		parsedDate, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			c.logError("parse end date", err, log.Fields{
				"end_date": req.EndDate,
			})
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

	err = c.service.UpdateSubscription(ctx.Request().Context(), id, sub)
	if err != nil {
		c.logError("update subscription", err, log.Fields{
			"subscription_id": id,
			"update_data":     fmt.Sprintf("%+v", sub),
		})
		return HTTPError(err)
	}

	c.logSuccess("update subscription", log.Fields{
		"subscription_id": id,
	})
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
	c.logRequest(ctx.Request().Method, ctx.Request().URL.Path)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		c.logError("parse subscription ID", err, log.Fields{
			"input_id": ctx.Param("id"),
		})
		return ctx.JSON(http.StatusBadRequest, ErrInvalidSubscription)
	}

	err = c.service.DeleteSubscription(ctx.Request().Context(), id)
	if err != nil {
		c.logError("delete subscription", err, log.Fields{
			"subscription_id": id,
		})
		return HTTPError(err)
	}

	c.logSuccess("delete subscription", log.Fields{
		"subscription_id": id,
	})
	return ctx.NoContent(http.StatusNoContent)
}

// ListByUser godoc
// @Summary Список подписок пользователя
// @Description Возвращает все подписки указанного пользователя с пагинацией
// @Tags Subscriptions
// @Produce json
// @Param user_id query string true "ID пользователя"
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество записей на странице (по умолчанию 10, максимум 100)"
// @Success 200 {array}  PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [get]
func (c *SubscriptionController) ListByUser(ctx echo.Context) error {
	c.logRequest(ctx.Request().Method, ctx.Request().URL.Path)

	userID, err := uuid.Parse(ctx.QueryParam("user_id"))
	if err != nil {
		c.logError("parse user ID", err, log.Fields{
			"user_id_hash": hashString(ctx.QueryParam("user_id")),
		})
		return ctx.JSON(http.StatusBadRequest, ErrInvalidUserID)
	}

	page, err := strconv.Atoi(ctx.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	subscriptions, total, err := c.service.ListSubscriptionsByUser(ctx.Request().Context(), userID, page, limit)
	if err != nil {
		c.logError("list subscriptions by user", err, log.Fields{
			"user_id_hash": hashString(userID.String()),
		})
		return HTTPError(err)
	}

	c.logSuccess("list subscriptions by user", log.Fields{
		"user_id_hash": hashString(userID.String()),
		"count":        len(subscriptions),
	})
	return ctx.JSON(http.StatusOK, PaginatedResponse{
		Items: subscriptions,
		Total: total,
		Page:  page,
		Limit: limit,
	})
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
	c.logRequest(ctx.Request().Method, ctx.Request().URL.Path)

	req := CalculateTotalCostRequest{
		UserID:      ctx.QueryParam("user_id"),
		ServiceName: ctx.QueryParam("service_name"),
		StartDate:   ctx.QueryParam("start_date"),
		EndDate:     ctx.QueryParam("end_date"),
	}

	if err := ctx.Validate(req); err != nil {
		c.logError("validate request", err, nil)
		return ctx.JSON(http.StatusBadRequest, ValidationError(err))
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		c.logError("parse start date", err, log.Fields{
			"start_date": req.StartDate,
		})
		return ctx.JSON(http.StatusBadRequest, ErrInvalidDateFormat)
	}

	endDate, err := time.Parse("01-2006", req.EndDate)
	if err != nil {
		c.logError("parse end date", err, log.Fields{
			"end_date": req.EndDate,
		})
		return ctx.JSON(http.StatusBadRequest, ErrInvalidDateFormat)
	}

	var userID *uuid.UUID
	if req.UserID != "" {
		parsedUUID, err := uuid.Parse(req.UserID)
		if err != nil {
			c.logError("parse user ID", err, log.Fields{
				"user_id_hash": hashString(req.UserID),
			})
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
		c.logError("calculate total cost", err, log.Fields{
			"user_id_hash": hashString(req.UserID),
			"service_name": serviceName,
			"start_date":   startDate.Format("01-2006"),
			"end_date":     endDate.Format("01-2006"),
		})
		return HTTPError(err)
	}

	c.logSuccess("calculate total cost", log.Fields{
		"total":        total,
		"user_id_hash": hashString(req.UserID),
		"service_name": serviceName,
		"start_date":   startDate.Format("01-2006"),
		"end_date":     endDate.Format("01-2006"),
	})
	return ctx.JSON(http.StatusOK, TotalCostResponse{Total: total})
}

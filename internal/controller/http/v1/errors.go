package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/repo/repoerrs"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

var (
	ErrBadRequest           = ErrorResponse{"invalid request"}
	ErrInvalidSubscription  = ErrorResponse{"invalid subscription id"}
	ErrInvalidUserID        = ErrorResponse{"invalid user id"}
	ErrInvalidDateFormat    = ErrorResponse{"invalid date format, use MM-YYYY"}
	ErrInvalidPrice         = ErrorResponse{"price must be positive"}
	ErrInvalidDateRange     = ErrorResponse{"start date must be before end date"}
	ErrEmptyServiceName     = ErrorResponse{"service name cannot be empty"}
	ErrSubscriptionNotFound = ErrorResponse{"subscription not found"}
	ErrSubscriptionExists   = ErrorResponse{"subscription already exists"}
	ErrInternalServer       = ErrorResponse{"internal server error"}
)

func HTTPError(err error) *echo.HTTPError {
	if err == nil {
		return nil
	}

	if errors.Is(err, repoerrs.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, ErrSubscriptionNotFound)
	}
	if errors.Is(err, repoerrs.ErrAlreadyExists) {
		return echo.NewHTTPError(http.StatusConflict, ErrSubscriptionExists)
	}

	errMsg := err.Error()

	switch {

	case strings.Contains(errMsg, "invalid date format"):
		return echo.NewHTTPError(http.StatusBadRequest, ErrInvalidDateFormat)
	case strings.Contains(errMsg, "invalid subscription id"):
		return echo.NewHTTPError(http.StatusBadRequest, ErrInvalidSubscription)
	case strings.Contains(errMsg, "invalid user id"):
		return echo.NewHTTPError(http.StatusBadRequest, ErrInvalidUserID)
	case strings.Contains(errMsg, "invalid request"):
		return echo.NewHTTPError(http.StatusBadRequest, ErrBadRequest)

	case strings.Contains(errMsg, "price must be positive"):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, ErrInvalidPrice)
	case strings.Contains(errMsg, "end date before start date"),
		strings.Contains(errMsg, "invalid date range"):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, ErrInvalidDateRange)
	case strings.Contains(errMsg, "empty service name"):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, ErrEmptyServiceName)

	case strings.Contains(errMsg, "not found"):
		return echo.NewHTTPError(http.StatusNotFound, ErrSubscriptionNotFound)

	default:
		return echo.NewHTTPError(http.StatusInternalServerError, ErrInternalServer)
	}
}

func ValidationError(err error) ErrorResponse {
	if fieldErr, ok := err.(validator.ValidationErrors); ok {
		// Берем только первое поле с ошибкой для простоты
		for _, e := range fieldErr {
			switch e.Tag() {
			case "required":
				return ErrorResponse{
					Message: fmt.Sprintf("field '%s' is required", e.Field()),
				}
			case "gt":
				return ErrorResponse{
					Message: fmt.Sprintf("field '%s' must be greater than %s", e.Field(), e.Param()),
				}
			// Добавьте другие нужные теги валидации
			default:
				return ErrorResponse{
					Message: fmt.Sprintf("field '%s' is invalid", e.Field()),
				}
			}
		}
	}
	return ErrorResponse{Message: err.Error()}
}

func BusinessError(err error) ErrorResponse {
	switch {
	case strings.Contains(err.Error(), "price must be positive"):
		return ErrInvalidPrice
	case strings.Contains(err.Error(), "end date before start date"):
		return ErrInvalidDateRange
	case strings.Contains(err.Error(), "empty service name"):
		return ErrEmptyServiceName
	default:
		return ErrInternalServer
	}
}

package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/repo/repoerrs"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

const (
	CodeBadRequest          = "BAD_REQUEST"
	CodeInvalidSubscription = "INVALID_SUBSCRIPTION_ID"
	CodeInvalidUserID       = "INVALID_USER_ID"
	CodeInvalidDateFormat   = "INVALID_DATE_FORMAT"
	CodeInvalidPrice        = "INVALID_PRICE"
	CodeInvalidDateRange    = "INVALID_DATE_RANGE"
	CodeEmptyServiceName    = "EMPTY_SERVICE_NAME"
	CodeNotFound            = "NOT_FOUND"
	CodeAlreadyExists       = "ALREADY_EXISTS"
	CodeInternalError       = "INTERNAL_ERROR"
)

var (
	ErrBadRequest           = ErrorResponse{Code: CodeBadRequest, Message: "invalid request"}
	ErrInvalidSubscription  = ErrorResponse{Code: CodeInvalidSubscription, Message: "invalid subscription id"}
	ErrInvalidUserID        = ErrorResponse{Code: CodeInvalidUserID, Message: "invalid user id"}
	ErrInvalidDateFormat    = ErrorResponse{Code: CodeInvalidDateFormat, Message: "invalid date format, use MM-YYYY"}
	ErrInvalidPrice         = ErrorResponse{Code: CodeInvalidPrice, Message: "price must be positive"}
	ErrInvalidDateRange     = ErrorResponse{Code: CodeInvalidDateRange, Message: "start date must be before end date"}
	ErrEmptyServiceName     = ErrorResponse{Code: CodeEmptyServiceName, Message: "service name cannot be empty"}
	ErrSubscriptionNotFound = ErrorResponse{Code: CodeNotFound, Message: "subscription not found"}
	ErrSubscriptionExists   = ErrorResponse{Code: CodeAlreadyExists, Message: "subscription already exists"}
	ErrInternalServer       = ErrorResponse{Code: CodeInternalError, Message: "internal server error"}
)

type ValidationError struct {
	Field   string
	Tag     string
	Param   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("validation failed for field '%s' with tag '%s'", e.Field, e.Tag)
}

type BusinessError struct {
	Code    string
	Message string
}

func (e *BusinessError) Error() string {
	return e.Message
}

func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

func HTTPError(err error) *echo.HTTPError {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, repoerrs.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, ErrSubscriptionNotFound)
	case errors.Is(err, repoerrs.ErrAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, ErrSubscriptionExists)

	case errors.As(err, new(*validator.ValidationErrors)):
		return handleValidationError(err)
	case errors.As(err, new(*ValidationError)):
		return handleCustomValidationError(err)
	case errors.As(err, new(*BusinessError)):
		return handleBusinessError(err)

	default:
		return echo.NewHTTPError(http.StatusInternalServerError, ErrInternalServer)
	}
}

func handleValidationError(err error) *echo.HTTPError {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return echo.NewHTTPError(http.StatusBadRequest, ErrBadRequest)
	}

	fieldErr := ve[0]
	resp := ErrorResponse{
		Message: fmt.Sprintf("field '%s' is invalid", fieldErr.Field()),
	}

	switch fieldErr.Tag() {
	case "required":
		resp.Message = fmt.Sprintf("field '%s' is required", fieldErr.Field())
	case "gt":
		resp.Message = fmt.Sprintf("field '%s' must be greater than %s", fieldErr.Field(), fieldErr.Param())
	}

	return echo.NewHTTPError(http.StatusUnprocessableEntity, resp)
}

func handleCustomValidationError(err error) *echo.HTTPError {
	var ve *ValidationError
	if !errors.As(err, &ve) {
		return echo.NewHTTPError(http.StatusBadRequest, ErrBadRequest)
	}

	resp := ErrorResponse{
		Message: ve.Message,
	}

	switch {
	case ve.Field == "date" && ve.Tag == "format":
		resp.Code = CodeInvalidDateFormat
		return echo.NewHTTPError(http.StatusBadRequest, resp)
	case ve.Field == "price" && ve.Tag == "gt":
		resp.Code = CodeInvalidPrice
		return echo.NewHTTPError(http.StatusUnprocessableEntity, resp)
	default:
		resp.Code = CodeBadRequest
		return echo.NewHTTPError(http.StatusBadRequest, resp)
	}
}

func handleBusinessError(err error) *echo.HTTPError {
	var be *BusinessError
	if !errors.As(err, &be) {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrInternalServer)
	}

	resp := ErrorResponse{
		Code:    be.Code,
		Message: be.Message,
	}

	switch be.Code {
	case CodeInvalidPrice:
		return echo.NewHTTPError(http.StatusUnprocessableEntity, resp)
	case CodeInvalidDateRange:
		return echo.NewHTTPError(http.StatusUnprocessableEntity, resp)
	case CodeEmptyServiceName:
		return echo.NewHTTPError(http.StatusUnprocessableEntity, resp)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, resp)
	}
}

func ValidationErrorResponse(err error) ErrorResponse {
	if fieldErr, ok := err.(validator.ValidationErrors); ok {
		for _, e := range fieldErr {
			switch e.Tag() {
			case "required":
				return ErrorResponse{
					Code:    CodeBadRequest,
					Message: fmt.Sprintf("field '%s' is required", e.Field()),
				}
			case "gt":
				return ErrorResponse{
					Code:    CodeInvalidPrice,
					Message: fmt.Sprintf("field '%s' must be greater than %s", e.Field(), e.Param()),
				}
			default:
				return ErrorResponse{
					Code:    CodeBadRequest,
					Message: fmt.Sprintf("field '%s' is invalid", e.Field()),
				}
			}
		}
	}
	return ErrorResponse{Code: CodeInternalError, Message: err.Error()}
}

func BusinessErrorResponse(err error) ErrorResponse {
	var be *BusinessError
	if errors.As(err, &be) {
		return ErrorResponse{
			Code:    be.Code,
			Message: be.Message,
		}
	}
	return ErrInternalServer
}

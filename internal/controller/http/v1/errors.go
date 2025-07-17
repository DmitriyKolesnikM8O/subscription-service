package v1

import (
	"net/http"

	"github.com/labstack/echo"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

var (
	ErrBadRequest          = ErrorResponse{"invalid request"}
	ErrInvalidSubscription = ErrorResponse{"invalid subscription id"}
	ErrInvalidUserID       = ErrorResponse{"invalid user id"}
	ErrInvalidDateFormat   = ErrorResponse{"invalid date format, use MM-YYYY"}
	ErrInvalidPrice        = ErrorResponse{"price must be positive"}
	ErrInvalidDateRange    = ErrorResponse{"start date must be before end date"}

	ErrSubscriptionNotFound = ErrorResponse{"subscription not found"}

	ErrInternalServer = ErrorResponse{"internal server error"}
)

func HTTPError(err error) *echo.HTTPError {
	switch err.Error() {
	case "subscription not found":
		return echo.NewHTTPError(http.StatusNotFound, ErrSubscriptionNotFound)
	case "invalid request body":
		return echo.NewHTTPError(http.StatusBadRequest, ErrBadRequest)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, ErrInternalServer)
	}
}

func ValidationError(err error) ErrorResponse {
	return ErrorResponse{Message: err.Error()}
}

func BusinessError(err error) ErrorResponse {
	switch err.Error() {
	case "price must be positive":
		return ErrInvalidPrice
	case "end date before start date":
		return ErrInvalidDateRange
	default:
		return ErrInternalServer
	}
}

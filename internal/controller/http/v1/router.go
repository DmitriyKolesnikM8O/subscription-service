package v1

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func NewRouter(handler *echo.Echo, services service.SubscriptionService) {

	handler.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}", "method":"${method}","uri":"${uri}", "status":${status},"error":"${error}"}` + "\n",
		Output: setLogsFile(),
	}))
	handler.Use(middleware.Recover())

	handler.GET("/health", func(c echo.Context) error {
		return c.NoContent(200)
	})
	handler.GET("/swagger/*", echoSwagger.WrapHandler)

	api := handler.Group("/api/v1")
	{
		SetupSubscriptionRoutes(api, services)
	}
}

func SetupSubscriptionRoutes(group *echo.Group, subService service.SubscriptionService) {
	ctrl := NewSubscriptionController(subService)

	group.POST("/subscriptions", ctrl.Create)
	group.GET("/subscriptions/:id", ctrl.GetByID)
	group.PUT("/subscriptions/:id", ctrl.Update)
	group.DELETE("/subscriptions/:id", ctrl.Delete)
	group.GET("/subscriptions", ctrl.ListByUser)
	group.GET("/subscriptions/total-cost", ctrl.CalculateTotalCost)
}

func setLogsFile() *os.File {
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile("logs/requests.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

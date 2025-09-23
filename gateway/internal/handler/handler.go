package handler

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vctrl/currency-service/gateway/internal/service"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type controller struct {
	authService     service.AuthService
	currencyService service.CurrencyService
	router          *gin.Engine
	logger          *zap.Logger
}

func RegisterRoutes(authSvc service.AuthService,
	currencySvc service.CurrencyService,
	router *gin.Engine,
	logger *zap.Logger) controller {

	cntrl := controller{
		authService:     authSvc,
		currencyService: currencySvc,
		router:          router,
		logger:          logger,
	}

	cntrl.router.GET(
		"/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		},
	)

	cntrl.router.GET("/metrics", func(c *gin.Context) {
		promhttp.Handler()
	},
	)

	cntrl.router.GET("/api/v1/rate", cntrl.GetCurrencyRates)
	cntrl.router.PATCH("/api/v1/rate", cntrl.UpdateCurrencyRate)
	cntrl.router.POST("/api/v1/login", cntrl.Login)
	cntrl.router.POST("/api/v1/register", cntrl.Register)
	cntrl.router.POST("/api/v1/logout", cntrl.Logout)

	return cntrl
}

package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vctrl/currency-service/auth/internal/service"
	"go.uber.org/zap"
)

type Controller struct {
	jwtService *service.JWTService
	router     *gin.Engine
	logger     *zap.Logger
}

func RegisterRoutes(jwtService *service.JWTService, router *gin.Engine, logger *zap.Logger) *Controller {
	controller := &Controller{jwtService: jwtService, router: router, logger: logger}
	controller.initRoutes()
	return controller
}

func (c *Controller) initRoutes() {
	c.router.GET("/ping", c.Ping)
	c.router.GET("/generate", c.GenerateToken)
	c.router.GET("/validate", c.ValidateToken)
}

func (c *Controller) Ping(ctx *gin.Context) {
	ctx.String(http.StatusOK, "pong")
}

func (c *Controller) GenerateToken(ctx *gin.Context) {
	userID := ctx.Query("login")
	token, err := c.jwtService.GenerateToken(userID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.String(http.StatusOK, token)
}

func (c *Controller) ValidateToken(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")
	if header == "" {
		ctx.String(http.StatusBadRequest, "jwt token missing")
		return
	}

	token := strings.Split(header, " ")[1]
	err := c.jwtService.ValidateToken(token)
	if err != nil {
		ctx.String(http.StatusUnauthorized, err.Error())
		return
	}

	ctx.Status(http.StatusOK)
}

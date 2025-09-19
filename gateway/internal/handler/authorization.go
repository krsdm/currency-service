package handler

import (
	"net/http"

	"github.com/vctrl/currency-service/gateway/internal/dto"

	"github.com/gin-gonic/gin"
)

type registerRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func (s *controller) Register(c *gin.Context) {
	var req registerRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = s.authService.Register(c.Request.Context(), dto.RegisterRequest(req))
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

type loginRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func (s *controller) Login(c *gin.Context) {
	var req loginRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		s.failedLogin.WithLabelValues(req.Username).Inc()
		return
	}

	token, err := s.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		s.failedLogin.WithLabelValues(req.Username).Inc()
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *controller) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization token is required"})
		return
	}

	err := s.authService.Logout(token)
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

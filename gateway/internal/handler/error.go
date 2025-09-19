package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vctrl/currency-service/gateway/internal/clients/auth"
	innnerErrors "github.com/vctrl/currency-service/gateway/internal/errors"
	"github.com/vctrl/currency-service/gateway/internal/password"
	"github.com/vctrl/currency-service/gateway/internal/repository"
)

// todo move errors to separate package
func (s *controller) handleError(c *gin.Context, err error) {
	var nferr innnerErrors.NotFoundError
	if errors.As(err, &nferr) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": nferr.Error(),
		})
	}

	log.Printf("internal error: %v", err)
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	case errors.Is(err, repository.ErrUserAlreadyExist):
		c.JSON(http.StatusConflict, gin.H{"error": "User already exist"})
	case errors.Is(err, auth.ErrUnexpectedStatusCode):
		log.Printf("unexpected status code error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected server error"}) // Обычный ответ клиенту
	case errors.Is(err, auth.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	case errors.Is(err, auth.ErrTokenGeneration):
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to generate token"},
		)
	case errors.Is(err, auth.ErrTokenNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token not found"})
	case errors.Is(err, auth.ErrInvalidOrExpiredToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is invalid or expired"})
	case errors.Is(err, password.ErrShortPassword),
		errors.Is(err, password.ErrLongPassword),
		errors.Is(err, password.ErrInvalidPassword):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.As(err, &password.ErrRequiredSymbols{}):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

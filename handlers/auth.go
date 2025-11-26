package handlers

import (
	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"go-multi-tenant/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	APIKey string       `json:"api_key"`
	User   *models.User `json:"user"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	user, apiKey, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Login failed", err)
		return
	}

	response := LoginResponse{
		APIKey: apiKey,
		User:   user,
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	apiKey := c.GetHeader("Authorization")
	if apiKey == "" {
		apiKey = c.GetHeader("X-API-Key")
	}

	if apiKey == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "API key required", nil)
		return
	}

	err := h.authService.Logout(apiKey)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Logout failed", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}

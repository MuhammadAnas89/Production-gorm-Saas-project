package handlers

import (
	"net/http"
	"strconv"

	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"go-multi-tenant/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService  *services.UserService
	queueService *services.QueueService
}

func NewUserHandler(userService *services.UserService, queueService *services.QueueService) *UserHandler {
	return &UserHandler{
		userService:  userService,
		queueService: queueService,
	}
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	RoleID   uint   `json:"role_id" binding:"required"`
}

type CreateUserResponse struct {
	JobID   string       `json:"job_id"`
	Message string       `json:"message"`
	User    *models.User `json:"user,omitempty"`
	Error   string       `json:"error,omitempty"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	// ❌ REMOVED: tenantDB ki yahan zaroorat nahi hai, Worker khud connect karega
	// tenantDB := c.MustGet("tenantDB").(*gorm.DB)

	currentUser := c.MustGet("user").(*models.User)

	// ✅ FIX: Removed 'tenantDB' from arguments
	jobID, err := h.queueService.EnqueueUserCreation(
		currentUser.TenantID,
		&services.CreateUserRequest{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
			RoleID:   req.RoleID,
		},
		currentUser,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	response := CreateUserResponse{
		JobID:   jobID,
		Message: "User creation job enqueued successfully",
	}

	utils.SuccessResponse(c, http.StatusAccepted, "User creation initiated", response)
}

func (h *UserHandler) GetQueueStats(c *gin.Context) {
	stats := h.queueService.GetQueueStats()
	utils.SuccessResponse(c, http.StatusOK, "Queue statistics", stats)
}

func (h *UserHandler) GetAvailableRoles(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	currentUser := c.MustGet("user").(*models.User)

	roles, err := h.userService.GetAvailableRoles(tenantDB, currentUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get available roles", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Available roles retrieved successfully", roles)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	currentUser := c.MustGet("user").(*models.User)

	user, err := h.userService.GetUser(tenantDB, uint(userID), currentUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	currentUser := c.MustGet("user").(*models.User)

	user, err := h.userService.UpdateUser(tenantDB, uint(userID), updateData, currentUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update user", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	currentUser := c.MustGet("user").(*models.User)

	err = h.userService.DeleteUser(tenantDB, uint(userID), currentUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	currentUser := c.MustGet("user").(*models.User)

	users, err := h.userService.ListUsers(tenantDB, currentUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list users", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Users retrieved successfully", users)
}

func (h *UserHandler) GetSelf(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	currentUser := c.MustGet("user").(*models.User)

	user, err := h.userService.GetSelf(tenantDB, currentUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user info", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User info retrieved successfully", user)
}

// Assign role to user
type AssignRoleRequest struct {
	RoleID uint `json:"role_id" binding:"required"`
}

func (h *UserHandler) AssignRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	currentUser := c.MustGet("user").(*models.User)

	err = h.userService.AssignRole(tenantDB, uint(userID), req.RoleID, currentUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign role", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Role assigned successfully", nil)
}

// Remove role from user
func (h *UserHandler) RemoveRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	roleIDStr := c.Param("rid")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err)
		return
	}

	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	currentUser := c.MustGet("user").(*models.User)

	err = h.userService.RemoveRole(tenantDB, uint(userID), uint(roleID), currentUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove role", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Role removed successfully", nil)
}

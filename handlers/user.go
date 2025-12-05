package handlers

import (
	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// 1. Create User
func (h *UserHandler) CreateUser(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	tenantID := c.MustGet("tenantID").(uint)
	userID := c.MustGet("userID").(uint)

	// Load Current User for permissions
	var currentUser models.User
	tenantDB.Preload("Roles.Permissions").First(&currentUser, userID)

	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(tenantDB, tenantID, &req, &currentUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created", "user": user})
}

// 2. Get User
func (h *UserHandler) GetUser(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.MustGet("userID").(uint)

	var currentUser models.User
	tenantDB.Preload("Roles").First(&currentUser, userID)

	user, err := h.userService.GetUser(tenantDB, uint(id), &currentUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// 3. List Users
func (h *UserHandler) ListUsers(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	var currentUser models.User
	tenantDB.Preload("Roles.Permissions").First(&currentUser, userID)

	users, err := h.userService.ListUsers(tenantDB, &currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

// 4. Update User
func (h *UserHandler) UpdateUser(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.MustGet("userID").(uint)

	var currentUser models.User
	tenantDB.Preload("Roles.Permissions").First(&currentUser, userID)

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(tenantDB, uint(id), req, &currentUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated", "user": user})
}

// 5. Delete User
func (h *UserHandler) DeleteUser(c *gin.Context) {
	tenantDB := c.MustGet("tenantDB").(*gorm.DB)
	id, _ := strconv.Atoi(c.Param("id"))
	userID := c.MustGet("userID").(uint)

	var currentUser models.User
	tenantDB.Preload("Roles.Permissions").First(&currentUser, userID)

	if err := h.userService.DeleteUser(tenantDB, uint(id), &currentUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

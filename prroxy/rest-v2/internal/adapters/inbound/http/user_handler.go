package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/domain/user"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *user.Service
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *user.Service) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// UserResponse represents the HTTP response for a user
type UserResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
}

// ErrorResponse represents an HTTP error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// GetUser handles GET /api/user/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	// Get user ID from URL parameter
	idParam := c.Param("id")

	// Parse user ID
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid user ID",
		})
		return
	}

	// Call service layer
	usr, err := h.userService.GetUser(userID)
	if err != nil {
		// Handle specific errors
		switch err {
		case user.ErrUserNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "User not found",
			})
		case user.ErrInvalidUserID:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid user ID",
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to fetch user",
			})
		}
		return
	}

	// Convert domain model to response
	response := UserResponse{
		ID:       usr.ID,
		Name:     usr.Name,
		Username: usr.Username,
		Email:    usr.Email,
		Phone:    usr.Phone,
		Website:  usr.Website,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserSummary handles GET /api/user/:id/summary
func (h *UserHandler) GetUserSummary(c *gin.Context) {
	// Get user ID from URL parameter
	idParam := c.Param("id")

	// Parse user ID
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid user ID",
		})
		return
	}

	// Call service layer
	summary, err := h.userService.GetUserSummary(userID)
	if err != nil {
		// Handle specific errors
		switch err {
		case user.ErrUserNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "User not found",
			})
		case user.ErrInvalidUserID:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid user ID",
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to fetch user summary",
			})
		}
		return
	}

	c.JSON(http.StatusOK, summary)
}

// PostUserReport handles POST /api/user/:id/report
func (h *UserHandler) PostUserReport(c *gin.Context) {
	// Get user ID from URL parameter
	idParam := c.Param("id")

	// Parse user ID
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid user ID",
		})
		return
	}

	// Parse optional request body
	var options user.ReportOptions
	if c.Request.Body != nil {
		if err := c.ShouldBindJSON(&options); err != nil {
			// If body is empty or invalid JSON, just use defaults
			// Don't return error, as body is optional
		}
	}

	// Call service layer
	report, err := h.userService.GetUserReport(userID, &options)
	if err != nil {
		// Handle specific errors
		switch err {
		case user.ErrUserNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "User not found",
			})
		case user.ErrInvalidUserID:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid user ID",
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to generate user report",
			})
		}
		return
	}

	c.JSON(http.StatusOK, report)
}

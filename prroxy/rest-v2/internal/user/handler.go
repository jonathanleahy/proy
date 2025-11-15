package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	apperrors "github.com/jonathanleahy/prroxy/rest-v2/internal/common/errors"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/common/response"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for user endpoints.
// It parses requests, delegates to the service layer, and formats responses.
type Handler struct {
	service UserService
	logger  *zap.Logger
}

// NewHandler creates a new user handler.
func NewHandler(service UserService) *Handler {
	logger, _ := zap.NewProduction()
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetUser handles GET /api/user/:id
// Returns basic user information.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	userID, err := h.extractUserID(r.URL.Path, "/api/user/")
	if err != nil {
		h.logger.Error("invalid user ID", zap.Error(err))
		response.Error(w, apperrors.New("BAD_REQUEST", "Invalid user ID", http.StatusBadRequest))
		return
	}

	// Call service
	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user", zap.Int("userID", userID), zap.Error(err))
		response.Error(w, err)
		return
	}

	response.Success(w, user)
}

// GetUserSummary handles GET /api/user/:id/summary
// Returns user information with post statistics.
func (h *Handler) GetUserSummary(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	userID, err := h.extractUserID(r.URL.Path, "/api/user/")
	if err != nil {
		h.logger.Error("invalid user ID", zap.Error(err))
		response.Error(w, apperrors.New("BAD_REQUEST", "Invalid user ID", http.StatusBadRequest))
		return
	}

	// Call service
	summary, err := h.service.GetUserSummary(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user summary", zap.Int("userID", userID), zap.Error(err))
		response.Error(w, err)
		return
	}

	response.Success(w, summary)
}

// GetUserReport handles POST /api/user/:id/report
// Returns comprehensive user report with filtering options.
func (h *Handler) GetUserReport(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	userID, err := h.extractUserID(r.URL.Path, "/api/user/")
	if err != nil {
		h.logger.Error("invalid user ID", zap.Error(err))
		response.Error(w, apperrors.New("BAD_REQUEST", "Invalid user ID", http.StatusBadRequest))
		return
	}

	// Parse request body
	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		response.Error(w, apperrors.New("BAD_REQUEST", "Invalid request body", http.StatusBadRequest))
		return
	}

	// Call service
	report, err := h.service.GetUserReport(r.Context(), userID, req)
	if err != nil {
		h.logger.Error("failed to get user report", zap.Int("userID", userID), zap.Error(err))
		response.Error(w, err)
		return
	}

	response.Success(w, report)
}

// extractUserID extracts the user ID from the URL path.
// Handles both /api/user/:id and /api/user/:id/summary paths.
func (h *Handler) extractUserID(path, prefix string) (int, error) {
	// Remove prefix
	trimmed := strings.TrimPrefix(path, prefix)

	// Split by / to get just the ID part
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, fmt.Errorf("missing user ID")
	}

	// Parse ID
	userID, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %w", err)
	}

	return userID, nil
}

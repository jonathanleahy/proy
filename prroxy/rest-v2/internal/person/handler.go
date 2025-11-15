package person

import (
	"net/http"
	"regexp"

	apperrors "github.com/jonathanleahy/prroxy/rest-v2/internal/common/errors"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/common/response"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for person endpoints.
// It parses requests, delegates to the service layer, and formats responses.
type Handler struct {
	service PersonService
	logger  *zap.Logger
}

// NewHandler creates a new person handler.
func NewHandler(service PersonService) *Handler {
	logger, _ := zap.NewProduction()
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// FindPerson handles GET /api/person?surname=X&dob=YYYY-MM-DD
// Returns a single person by exact match.
func (h *Handler) FindPerson(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	surname := r.URL.Query().Get("surname")
	dob := r.URL.Query().Get("dob")

	// Validate required parameters
	if surname == "" || dob == "" {
		h.logger.Error("missing required parameters", zap.String("surname", surname), zap.String("dob", dob))
		response.Error(w, apperrors.New("BAD_REQUEST", "Both surname and dob are required", http.StatusBadRequest))
		return
	}

	// Validate date format
	if !isValidDateFormat(dob) {
		h.logger.Error("invalid date format", zap.String("dob", dob))
		response.Error(w, apperrors.New("BAD_REQUEST", "dob must be in format YYYY-MM-DD", http.StatusBadRequest))
		return
	}

	// Call service
	person, err := h.service.FindPerson(r.Context(), surname, dob)
	if err != nil {
		h.logger.Error("failed to find person", zap.String("surname", surname), zap.String("dob", dob), zap.Error(err))
		response.Error(w, err)
		return
	}

	response.Success(w, person)
}

// FindPeople handles GET /api/people?surname=X or GET /api/people?dob=YYYY-MM-DD
// Returns array of people matching the search criteria.
func (h *Handler) FindPeople(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	surname := r.URL.Query().Get("surname")
	dob := r.URL.Query().Get("dob")

	// Validate at least one parameter provided
	if surname == "" && dob == "" {
		h.logger.Error("missing required parameters")
		response.Error(w, apperrors.New("BAD_REQUEST", "At least one of surname or dob is required", http.StatusBadRequest))
		return
	}

	// Validate date format if provided
	if dob != "" && !isValidDateFormat(dob) {
		h.logger.Error("invalid date format", zap.String("dob", dob))
		response.Error(w, apperrors.New("BAD_REQUEST", "dob must be in format YYYY-MM-DD", http.StatusBadRequest))
		return
	}

	// Call service
	people, err := h.service.FindPeople(r.Context(), surname, dob)
	if err != nil {
		h.logger.Error("failed to find people", zap.String("surname", surname), zap.String("dob", dob), zap.Error(err))
		response.Error(w, err)
		return
	}

	response.Success(w, people)
}

// isValidDateFormat checks if date string is in YYYY-MM-DD format
func isValidDateFormat(date string) bool {
	// Match YYYY-MM-DD format
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, date)
	return matched
}

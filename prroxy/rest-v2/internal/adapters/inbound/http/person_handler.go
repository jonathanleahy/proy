package http

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/domain/person"
)

// PersonHandler handles HTTP requests for person endpoints
type PersonHandler struct {
	personService *person.Service
}

// NewPersonHandler creates a new person handler
func NewPersonHandler(personService *person.Service) *PersonHandler {
	return &PersonHandler{
		personService: personService,
	}
}

// GetPerson handles GET /api/person?surname=X&dob=Y
// Finds a person by both surname and date of birth
func (h *PersonHandler) GetPerson(c *gin.Context) {
	surname := c.Query("surname")
	dob := c.Query("dob")

	// Validate required parameters
	if surname == "" || dob == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required parameters",
			"message": "Both surname and dob are required",
		})
		return
	}

	// Validate dob format (YYYY-MM-DD)
	dobRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dobRegex.MatchString(dob) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid date format",
			"message": "dob must be in format YYYY-MM-DD",
		})
		return
	}

	// Find person
	foundPerson, err := h.personService.FindPerson(surname, dob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Handle not found
	if foundPerson == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Person not found",
			"message": "No person found with surname \"" + surname + "\" and dob \"" + dob + "\"",
		})
		return
	}

	// Return found person as a single object to match v1 response format
	c.JSON(http.StatusOK, foundPerson)
}

// GetPeople handles GET /api/people?surname=X or GET /api/people?dob=Y
// Searches for people by surname OR dob (partial search)
func (h *PersonHandler) GetPeople(c *gin.Context) {
	surnameParam := c.Query("surname")
	dobParam := c.Query("dob")

	// Validate at least one parameter is provided
	if surnameParam == "" && dobParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing required parameters",
			"message": "At least one of surname or dob is required",
		})
		return
	}

	// Validate dob format if provided
	if dobParam != "" {
		dobRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
		if !dobRegex.MatchString(dobParam) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid date format",
				"message": "dob must be in format YYYY-MM-DD",
			})
			return
		}
	}

	// Convert empty strings to nil pointers for optional parameters
	var surname, dob *string
	if surnameParam != "" {
		surname = &surnameParam
	}
	if dobParam != "" {
		dob = &dobParam
	}

	// Search for people
	people, err := h.personService.FindPeople(surname, dob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Return results (empty array if no matches)
	c.JSON(http.StatusOK, people)
}

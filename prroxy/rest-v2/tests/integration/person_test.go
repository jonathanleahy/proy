package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	httpAdapter "github.com/jonathanleahy/prroxy/rest-v2/internal/adapters/inbound/http"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/domain/person"
	"github.com/stretchr/testify/assert"
)

func TestGetPerson_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	personService := person.NewService()
	personHandler := httpAdapter.NewPersonHandler(personService)

	router.GET("/api/person", personHandler.GetPerson)

	// Test with valid surname and dob
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/person?surname=Schmidt&dob=1991-09-03", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

	var result person.Person
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err, "Should unmarshal JSON response")
	assert.Equal(t, "Schmidt", result.Surname, "Surname should match")
	assert.Equal(t, "1991-09-03", result.DOB, "DOB should match")
	assert.NotEmpty(t, result.Firstname, "Should have firstname")
	assert.NotEmpty(t, result.Country, "Should have country")
}

func TestGetPerson_ResponseMatchesV1FieldNames(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	personService := person.NewService()
	personHandler := httpAdapter.NewPersonHandler(personService)

	router.GET("/api/person", personHandler.GetPerson)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/person?surname=Schmidt&dob=1991-09-03", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err, "Should parse JSON body")
	_, hasSurname := body["surname"]
	_, hasSurnamea := body["surnamea"]
	assert.True(t, hasSurname, "Response must include surname field to match v1")
	assert.False(t, hasSurnamea, "Response must not include legacy surnamea field")

	// CRITICAL: V1 uses "country" (singular), not "countries" (plural)
	_, hasCountry := body["country"]
	_, hasCountries := body["countries"]
	assert.True(t, hasCountry, "Response must include 'country' field (singular) to match v1")
	assert.False(t, hasCountries, "Response must NOT include 'countries' field (plural) - v1 uses singular")
}

func TestGetPerson_MissingParameters(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	personService := person.NewService()
	personHandler := httpAdapter.NewPersonHandler(personService)

	router.GET("/api/person", personHandler.GetPerson)

	// Test missing surname
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/person?dob=1991-09-03", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 when surname is missing")

	// Test missing dob
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/person?surname=Schmidt", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 when dob is missing")
}

func TestGetPerson_InvalidDateFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	personService := person.NewService()
	personHandler := httpAdapter.NewPersonHandler(personService)

	router.GET("/api/person", personHandler.GetPerson)

	// Test invalid date format
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/person?surname=Schmidt&dob=invalid-date", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 for invalid date format")
}

func TestGetPeople_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	personService := person.NewService()
	personHandler := httpAdapter.NewPersonHandler(personService)

	router.GET("/api/people", personHandler.GetPeople)

	// Test with surname only
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/people?surname=Schmidt", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

	var results []person.Person
	err := json.Unmarshal(w.Body.Bytes(), &results)
	assert.NoError(t, err, "Should unmarshal JSON response")
	assert.NotEmpty(t, results, "Should return at least one person")
}

func TestGetPeople_ResponseMatchesV1FieldNames(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	personService := person.NewService()
	personHandler := httpAdapter.NewPersonHandler(personService)

	router.GET("/api/people", personHandler.GetPeople)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/people?surname=Schmidt", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

	var body []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err, "Should parse JSON array response")
	assert.NotEmpty(t, body, "Expected at least one person in response")

	for _, person := range body {
		_, hasSurname := person["surname"]
		_, hasSurnamea := person["surnamea"]
		assert.True(t, hasSurname, "Each person must include surname field to match v1")
		assert.False(t, hasSurnamea, "Each person must not include legacy surnamea field")
	}
}

func TestGetPeople_MissingParameters(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	personService := person.NewService()
	personHandler := httpAdapter.NewPersonHandler(personService)

	router.GET("/api/people", personHandler.GetPeople)

	// Test with no parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/people", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 when no parameters provided")
}

// TestGetPerson_SpecialCharacters tests URL encoding with special characters like ü in Müller
func TestGetPerson_SpecialCharacters(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	personService := person.NewService()
	personHandler := httpAdapter.NewPersonHandler(personService)

	router.GET("/api/person", personHandler.GetPerson)

	// Test with surname containing special characters (ü)
	// This tests proper URL encoding of the target parameter when calling the proxy
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/person?surname=Müller&dob=1985-02-28", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK for Müller")

	var result person.Person
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err, "Should unmarshal JSON response")
	assert.Equal(t, "Müller", result.Surname, "Surname should be Müller")
	assert.Equal(t, "1985-02-28", result.DOB, "DOB should match")
	assert.Equal(t, "Sebastian", result.Firstname, "Firstname should be Sebastian")
	assert.Equal(t, "Austria", result.Country, "Country should be Austria")
}

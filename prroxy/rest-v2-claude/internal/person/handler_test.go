package person_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apperrors "github.com/jonathanleahy/prroxy/rest-v2/internal/common/errors"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/person"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/person/mocks"
)

func TestHandler_FindPerson(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mockReturn     *person.Person
		mockError      error
		wantStatus     int
		wantBody       map[string]interface{}
		wantErrMessage string
	}{
		{
			name: "success",
			path: "/api/person?surname=Thompson&dob=1985-03-15",
			mockReturn: &person.Person{
				Firstname: "Emma",
				Surname:   "Thompson",
				DOB:       "1985-03-15",
				Country:   "United Kingdom",
			},
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"firstname": "Emma",
				"surname":   "Thompson",
				"dob":       "1985-03-15",
				"country":   "United Kingdom",
			},
		},
		{
			name:           "missing surname parameter",
			path:           "/api/person?dob=1985-03-15",
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "Both surname and dob are required",
		},
		{
			name:           "missing dob parameter",
			path:           "/api/person?surname=Thompson",
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "Both surname and dob are required",
		},
		{
			name:           "invalid date format",
			path:           "/api/person?surname=Thompson&dob=invalid",
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "dob must be in format YYYY-MM-DD",
		},
		{
			name:           "person not found",
			path:           "/api/person?surname=NotFound&dob=2000-01-01",
			mockError:      apperrors.ErrNotFound,
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "Resource not found",
		},
		{
			name:           "service error",
			path:           "/api/person?surname=Error&dob=1990-01-01",
			mockError:      errors.New("service error"),
			wantStatus:     http.StatusInternalServerError,
			wantErrMessage: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.PersonService{}
			if tt.mockError != nil || tt.mockReturn != nil {
				mockService.On("FindPerson", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(tt.mockReturn, tt.mockError)
			}

			handler := person.NewHandler(mockService)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.FindPerson(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErrMessage != "" {
				var body map[string]string
				json.NewDecoder(w.Body).Decode(&body)
				assert.Equal(t, tt.wantErrMessage, body["error"])
			} else if tt.wantBody != nil {
				var body map[string]interface{}
				json.NewDecoder(w.Body).Decode(&body)
				assert.Equal(t, tt.wantBody, body)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_FindPeople(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mockReturn     []person.Person
		mockError      error
		wantStatus     int
		wantCount      int
		wantErrMessage string
	}{
		{
			name: "success by surname",
			path: "/api/people?surname=Thompson",
			mockReturn: []person.Person{
				{
					Firstname: "Emma",
					Surname:   "Thompson",
					DOB:       "1985-03-15",
					Country:   "United Kingdom",
				},
				{
					Firstname: "James",
					Surname:   "Thompson",
					DOB:       "1990-05-20",
					Country:   "United States",
				},
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name: "success by dob",
			path: "/api/people?dob=1985-03-15",
			mockReturn: []person.Person{
				{
					Firstname: "Emma",
					Surname:   "Thompson",
					DOB:       "1985-03-15",
					Country:   "United Kingdom",
				},
			},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			name:           "missing both parameters",
			path:           "/api/people",
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "At least one of surname or dob is required",
		},
		{
			name:           "invalid date format",
			path:           "/api/people?dob=invalid",
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "dob must be in format YYYY-MM-DD",
		},
		{
			name:           "service error",
			path:           "/api/people?surname=Error",
			mockError:      errors.New("service error"),
			wantStatus:     http.StatusInternalServerError,
			wantErrMessage: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.PersonService{}
			if tt.mockError != nil || tt.mockReturn != nil {
				mockService.On("FindPeople", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(tt.mockReturn, tt.mockError)
			}

			handler := person.NewHandler(mockService)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.FindPeople(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErrMessage != "" {
				var body map[string]string
				json.NewDecoder(w.Body).Decode(&body)
				assert.Equal(t, tt.wantErrMessage, body["error"])
			} else if tt.wantCount > 0 {
				var body []person.Person
				json.NewDecoder(w.Body).Decode(&body)
				assert.Len(t, body, tt.wantCount)
			}

			mockService.AssertExpectations(t)
		})
	}
}

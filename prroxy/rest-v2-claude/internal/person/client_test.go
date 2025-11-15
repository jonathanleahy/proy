package person_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonathanleahy/prroxy/rest-v2/internal/person"
)

func TestNewClient(t *testing.T) {
	client := person.NewClient("http://0.0.0.0:3006")
	assert.NotNil(t, client)
}

func TestClient_FindPerson(t *testing.T) {
	tests := []struct {
		name         string
		surname      string
		dob          string
		mockResponse interface{}
		mockStatus   int
		wantErr      bool
		wantPerson   *person.Person
	}{
		{
			name:    "successful find person",
			surname: "Thompson",
			dob:     "1985-03-15",
			mockResponse: map[string]interface{}{
				"firstname": "Emma",
				"surname":   "Thompson",
				"dob":       "1985-03-15",
				"country":   "United Kingdom",
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
			wantPerson: &person.Person{
				Firstname: "Emma",
				Surname:   "Thompson",
				DOB:       "1985-03-15",
				Country:   "United Kingdom",
			},
		},
		{
			name:         "person not found",
			surname:      "NotFound",
			dob:          "2000-01-01",
			mockResponse: map[string]interface{}{},
			mockStatus:   http.StatusNotFound,
			wantErr:      true,
		},
		{
			name:         "server error",
			surname:      "Error",
			dob:          "1990-01-01",
			mockResponse: map[string]string{"error": "internal error"},
			mockStatus:   http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server to simulate proxy
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify query parameters are in target
				target := r.URL.Query().Get("target")
				assert.Contains(t, target, "surname="+tt.surname)
				assert.Contains(t, target, "dob="+tt.dob)

				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client := person.NewClient(server.URL)
			ctx := context.Background()

			gotPerson, err := client.FindPerson(ctx, tt.surname, tt.dob)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantPerson, gotPerson)
		})
	}
}

func TestClient_FindPeople(t *testing.T) {
	tests := []struct {
		name         string
		surname      string
		dob          string
		mockResponse interface{}
		mockStatus   int
		wantErr      bool
		wantCount    int
	}{
		{
			name:    "successful find by surname",
			surname: "Thompson",
			dob:     "",
			mockResponse: []map[string]interface{}{
				{
					"firstname": "Emma",
					"surname":   "Thompson",
					"dob":       "1985-03-15",
					"country":   "United Kingdom",
				},
				{
					"firstname": "James",
					"surname":   "Thompson",
					"dob":       "1990-05-20",
					"country":   "United States",
				},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:    "successful find by dob",
			surname: "",
			dob:     "1985-03-15",
			mockResponse: []map[string]interface{}{
				{
					"firstname": "Emma",
					"surname":   "Thompson",
					"dob":       "1985-03-15",
					"country":   "United Kingdom",
				},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:         "server error",
			surname:      "Error",
			dob:          "",
			mockResponse: map[string]string{"error": "internal error"},
			mockStatus:   http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				target := r.URL.Query().Get("target")
				if tt.surname != "" {
					assert.Contains(t, target, "surname="+tt.surname)
				}
				if tt.dob != "" {
					assert.Contains(t, target, "dob="+tt.dob)
				}

				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client := person.NewClient(server.URL)
			ctx := context.Background()

			people, err := client.FindPeople(ctx, tt.surname, tt.dob)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, people, tt.wantCount)
		})
	}
}

package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/jonathanleahy/prroxy/rest-v2/internal/common/errors"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/common/response"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		wantStatus     int
		wantBody       string
		wantHeader     string
	}{
		{
			name:       "success with map",
			status:     200,
			data:       map[string]string{"message": "success"},
			wantStatus: 200,
			wantBody:   `{"message":"success"}`,
			wantHeader: "application/json",
		},
		{
			name:       "success with struct",
			status:     201,
			data:       struct{ ID int }{ID: 123},
			wantStatus: 201,
			wantBody:   `{"ID":123}`,
			wantHeader: "application/json",
		},
		{
			name:       "error status",
			status:     400,
			data:       map[string]string{"error": "bad request"},
			wantStatus: 400,
			wantBody:   `{"error":"bad request"}`,
			wantHeader: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			response.JSON(w, tt.status, tt.data)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantHeader, w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{
			name:       "app error not found",
			err:        apperrors.ErrNotFound,
			wantStatus: 404,
			wantError:  "Resource not found",
		},
		{
			name:       "app error bad request",
			err:        apperrors.ErrBadRequest,
			wantStatus: 400,
			wantError:  "Invalid request",
		},
		{
			name:       "app error internal",
			err:        apperrors.ErrInternal,
			wantStatus: 500,
			wantError:  "Internal server error",
		},
		{
			name: "wrapped app error",
			err: apperrors.Wrap(apperrors.ErrNotFound, errors.New("user not found")),
			wantStatus: 404,
			wantError:  "Resource not found",
		},
		{
			name:       "generic error",
			err:        errors.New("something went wrong"),
			wantStatus: 500,
			wantError:  "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			response.Error(w, tt.err)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var body map[string]string
			err := json.NewDecoder(w.Body).Decode(&body)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantError, body["error"])
		})
	}
}

func TestSuccess(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success with data",
			data:       map[string]string{"name": "John"},
			wantStatus: 200,
			wantBody:   `{"name":"John"}`,
		},
		{
			name:       "success with nil",
			data:       nil,
			wantStatus: 200,
			wantBody:   `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			response.Success(w, tt.data)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestCreated(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
		wantStatus int
		wantBody   string
	}{
		{
			name:       "created with data",
			data:       map[string]int{"id": 123},
			wantStatus: 201,
			wantBody:   `{"id":123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			response.Created(w, tt.data)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	response.NoContent(w)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

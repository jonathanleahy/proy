package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apperrors "github.com/jonathanleahy/prroxy/rest-v2/internal/common/errors"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/user"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/user/mocks"
)

func TestHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mockReturn     *user.User
		mockError      error
		wantStatus     int
		wantBody       map[string]interface{}
		wantErrMessage string
	}{
		{
			name: "success",
			path: "/api/user/1",
			mockReturn: &user.User{
				ID:       1,
				Name:     "John Doe",
				Username: "johndoe",
				Email:    "john@example.com",
				Phone:    "123-456-7890",
				Website:  "example.com",
			},
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"id":       float64(1),
				"name":     "John Doe",
				"username": "johndoe",
				"email":    "john@example.com",
				"phone":    "123-456-7890",
				"website":  "example.com",
			},
		},
		{
			name:           "invalid user ID",
			path:           "/api/user/abc",
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "Invalid user ID",
		},
		{
			name:           "user not found",
			path:           "/api/user/999",
			mockError:      apperrors.ErrNotFound,
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "Resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.UserService{}
			if tt.mockError != nil || tt.mockReturn != nil {
				mockService.On("GetUser", mock.Anything, mock.AnythingOfType("int")).
					Return(tt.mockReturn, tt.mockError)
			}

			handler := user.NewHandler(mockService)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.GetUser(w, req)

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

func TestHandler_GetUserSummary(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mockReturn     *user.UserSummary
		mockError      error
		wantStatus     int
		wantErrMessage string
	}{
		{
			name: "success",
			path: "/api/user/1/summary",
			mockReturn: &user.UserSummary{
				UserID:      1,
				UserName:    "John Doe",
				Email:       "john@example.com",
				PostCount:   10,
				RecentPosts: []string{"Post 1", "Post 2"},
				Summary:     "User John Doe has written 10 posts",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			path:           "/api/user/invalid/summary",
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "Invalid user ID",
		},
		{
			name:           "service error",
			path:           "/api/user/1/summary",
			mockError:      errors.New("service error"),
			wantStatus:     http.StatusInternalServerError,
			wantErrMessage: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.UserService{}
			if tt.mockError != nil || tt.mockReturn != nil {
				mockService.On("GetUserSummary", mock.Anything, mock.AnythingOfType("int")).
					Return(tt.mockReturn, tt.mockError)
			}

			handler := user.NewHandler(mockService)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.GetUserSummary(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErrMessage != "" {
				var body map[string]string
				json.NewDecoder(w.Body).Decode(&body)
				assert.Equal(t, tt.wantErrMessage, body["error"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_GetUserReport(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		body           interface{}
		mockReturn     *user.UserReport
		mockError      error
		wantStatus     int
		wantErrMessage string
	}{
		{
			name: "success",
			path: "/api/user/1/report",
			body: map[string]interface{}{
				"includeCompleted": true,
				"maxPosts":         5,
			},
			mockReturn: &user.UserReport{
				UserID:   1,
				UserName: "John Doe",
				Email:    "john@example.com",
				Stats: user.ReportStats{
					TotalPosts:     10,
					TotalTodos:     20,
					CompletedTodos: 10,
					PendingTodos:   10,
					CompletionRate: "50.0%",
				},
				Posts:       []user.PostPreview{},
				Todos:       user.TodoGroups{},
				GeneratedAt: "2024-01-01T00:00:00Z",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			path:           "/api/user/abc/report",
			body:           map[string]interface{}{},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "Invalid user ID",
		},
		{
			name:           "invalid JSON body",
			path:           "/api/user/1/report",
			body:           "invalid json",
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "Invalid request body",
		},
		{
			name: "service error",
			path: "/api/user/1/report",
			body: map[string]interface{}{
				"includeCompleted": true,
			},
			mockError:      apperrors.ErrInternal,
			wantStatus:     http.StatusInternalServerError,
			wantErrMessage: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mocks.UserService{}
			if tt.mockError != nil || tt.mockReturn != nil {
				mockService.On("GetUserReport", mock.Anything, mock.AnythingOfType("int"), mock.AnythingOfType("user.ReportRequest")).
					Return(tt.mockReturn, tt.mockError)
			}

			handler := user.NewHandler(mockService)

			var reqBody []byte
			if str, ok := tt.body.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.GetUserReport(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErrMessage != "" {
				var body map[string]string
				json.NewDecoder(w.Body).Decode(&body)
				assert.Equal(t, tt.wantErrMessage, body["error"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

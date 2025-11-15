package user_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonathanleahy/prroxy/rest-v2/internal/user"
)

func TestNewClient(t *testing.T) {
	client := user.NewClient("https://jsonplaceholder.typicode.com")
	assert.NotNil(t, client)
}

func TestClient_GetUser(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		mockResponse interface{}
		mockStatus   int
		wantErr      bool
		wantUser     *user.User
	}{
		{
			name:   "successful get user",
			userID: 1,
			mockResponse: map[string]interface{}{
				"id":       1,
				"name":     "Leanne Graham",
				"username": "Bret",
				"email":    "Sincere@april.biz",
				"phone":    "1-770-736-8031 x56442",
				"website":  "hildegard.org",
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
			wantUser: &user.User{
				ID:       1,
				Name:     "Leanne Graham",
				Username: "Bret",
				Email:    "Sincere@april.biz",
				Phone:    "1-770-736-8031 x56442",
				Website:  "hildegard.org",
			},
		},
		{
			name:         "user not found",
			userID:       999,
			mockResponse: map[string]interface{}{},
			mockStatus:   http.StatusNotFound,
			wantErr:      true,
		},
		{
			name:         "server error",
			userID:       1,
			mockResponse: map[string]string{"error": "internal error"},
			mockStatus:   http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server to simulate proxy
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client := user.NewClient(server.URL)
			ctx := context.Background()

			gotUser, err := client.GetUser(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUser, gotUser)
		})
	}
}

func TestClient_GetPosts(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		mockResponse interface{}
		mockStatus   int
		wantErr      bool
		wantCount    int
	}{
		{
			name:   "successful get posts",
			userID: 1,
			mockResponse: []map[string]interface{}{
				{
					"userId": 1,
					"id":     1,
					"title":  "Post 1",
					"body":   "Body 1",
				},
				{
					"userId": 1,
					"id":     2,
					"title":  "Post 2",
					"body":   "Body 2",
				},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:         "server error",
			userID:       1,
			mockResponse: map[string]string{"error": "internal error"},
			mockStatus:   http.StatusInternalServerError,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.Query().Get("target"), "posts")
				assert.Contains(t, r.URL.Query().Get("target"), "userId=1")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client := user.NewClient(server.URL)
			ctx := context.Background()

			posts, err := client.GetPosts(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, posts, tt.wantCount)
		})
	}
}

func TestClient_GetTodos(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		mockResponse interface{}
		mockStatus   int
		wantErr      bool
		wantCount    int
	}{
		{
			name:   "successful get todos",
			userID: 1,
			mockResponse: []map[string]interface{}{
				{
					"userId":    1,
					"id":        1,
					"title":     "Todo 1",
					"completed": true,
				},
				{
					"userId":    1,
					"id":        2,
					"title":     "Todo 2",
					"completed": false,
				},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.Query().Get("target"), "todos")
				assert.Contains(t, r.URL.Query().Get("target"), "userId=1")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client := user.NewClient(server.URL)
			ctx := context.Background()

			todos, err := client.GetTodos(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, todos, tt.wantCount)
		})
	}
}

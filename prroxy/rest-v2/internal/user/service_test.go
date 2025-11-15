package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jonathanleahy/prroxy/rest-v2/internal/user"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/user/mocks"
)

func TestService_GetUser(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		mockReturn *user.User
		mockError  error
		wantErr    bool
	}{
		{
			name:   "success",
			userID: 1,
			mockReturn: &user.User{
				ID:    1,
				Name:  "John Doe",
				Email: "john@example.com",
			},
			wantErr: false,
		},
		{
			name:      "client error",
			userID:    999,
			mockError: errors.New("not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewUserClient(t)
			mockClient.On("GetUser", mock.Anything, tt.userID).
				Return(tt.mockReturn, tt.mockError)

			service := user.NewService(mockClient)
			ctx := context.Background()

			result, err := service.GetUser(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.mockReturn, result)
		})
	}
}

func TestService_GetUserSummary(t *testing.T) {
	mockUser := &user.User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	mockPosts := []user.Post{
		{ID: 1, UserID: 1, Title: "Post 1", Body: "Body 1"},
		{ID: 2, UserID: 1, Title: "Post 2", Body: "Body 2"},
	}

	tests := []struct {
		name         string
		userID       int
		mockUser     *user.User
		mockUserErr  error
		mockPosts    []user.Post
		mockPostsErr error
		wantErr      bool
		wantCount    int
	}{
		{
			name:      "success",
			userID:    1,
			mockUser:  mockUser,
			mockPosts: mockPosts,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:        "user fetch error",
			userID:      1,
			mockUserErr: errors.New("user not found"),
			wantErr:     true,
		},
		{
			name:         "posts fetch error",
			userID:       1,
			mockUser:     mockUser,
			mockPostsErr: errors.New("posts fetch failed"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewUserClient(t)
			mockClient.On("GetUser", mock.Anything, tt.userID).
				Return(tt.mockUser, tt.mockUserErr)

			if tt.mockUserErr == nil {
				mockClient.On("GetPosts", mock.Anything, tt.userID).
					Return(tt.mockPosts, tt.mockPostsErr)
			}

			service := user.NewService(mockClient)
			ctx := context.Background()

			result, err := service.GetUserSummary(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.mockUser.ID, result.UserID)
			assert.Equal(t, tt.mockUser.Name, result.UserName)
			assert.Equal(t, tt.wantCount, result.PostCount)
			assert.Len(t, result.RecentPosts, tt.wantCount)
		})
	}
}

func TestService_GetUserReport(t *testing.T) {
	mockUser := &user.User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	mockPosts := []user.Post{
		{ID: 1, UserID: 1, Title: "Post 1", Body: "Body 1"},
		{ID: 2, UserID: 1, Title: "Post 2", Body: "Body 2"},
		{ID: 3, UserID: 1, Title: "Post 3", Body: "Body 3"},
	}

	mockTodos := []user.Todo{
		{ID: 1, UserID: 1, Title: "Todo 1", Completed: true},
		{ID: 2, UserID: 1, Title: "Todo 2", Completed: false},
		{ID: 3, UserID: 1, Title: "Todo 3", Completed: true},
	}

	tests := []struct {
		name           string
		userID         int
		request        user.ReportRequest
		wantPostCount  int
		wantTodosPend  int
		wantTodosComp  int
	}{
		{
			name:   "success with all completed todos",
			userID: 1,
			request: user.ReportRequest{
				IncludeCompleted: true,
			},
			wantPostCount: 3,
			wantTodosPend: 1,
			wantTodosComp: 2,
		},
		{
			name:   "success without completed todos",
			userID: 1,
			request: user.ReportRequest{
				IncludeCompleted: false,
			},
			wantPostCount: 3,
			wantTodosPend: 1,
			wantTodosComp: 0,
		},
		{
			name:   "success with max posts limit",
			userID: 1,
			request: user.ReportRequest{
				IncludeCompleted: true,
				MaxPosts:         func() *int { i := 2; return &i }(),
			},
			wantPostCount: 2,
			wantTodosPend: 1,
			wantTodosComp: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewUserClient(t)
			mockClient.On("GetUser", mock.Anything, tt.userID).Return(mockUser, nil)
			mockClient.On("GetPosts", mock.Anything, tt.userID).Return(mockPosts, nil)
			mockClient.On("GetTodos", mock.Anything, tt.userID).Return(mockTodos, nil)

			service := user.NewService(mockClient)
			ctx := context.Background()

			result, err := service.GetUserReport(ctx, tt.userID, tt.request)

			assert.NoError(t, err)
			assert.Equal(t, mockUser.ID, result.UserID)
			assert.Equal(t, mockUser.Name, result.UserName)
			assert.Len(t, result.Posts, tt.wantPostCount)
			assert.Len(t, result.Todos.Pending, tt.wantTodosPend)
			assert.Len(t, result.Todos.Completed, tt.wantTodosComp)
			assert.NotEmpty(t, result.GeneratedAt)
		})
	}
}

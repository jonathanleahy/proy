package user

import (
	"context"
	"fmt"
)

// UserService is the interface for user business logic.
// This interface allows for easy mocking in handler tests.
type UserService interface {
	GetUser(ctx context.Context, userID int) (*User, error)
	GetUserSummary(ctx context.Context, userID int) (*UserSummary, error)
	GetUserReport(ctx context.Context, userID int, req ReportRequest) (*UserReport, error)
}

// Service handles user-related business logic.
// It coordinates between the client and presents a unified interface
// to the HTTP layer.
type Service struct {
	client UserClient
}

// NewService creates a new user service.
func NewService(client UserClient) *Service {
	return &Service{
		client: client,
	}
}

// GetUser retrieves basic user information by ID.
func (s *Service) GetUser(ctx context.Context, userID int) (*User, error) {
	return s.client.GetUser(ctx, userID)
}

// GetUserSummary retrieves user information with post statistics.
// Makes two API calls: one for user data, one for posts.
func (s *Service) GetUserSummary(ctx context.Context, userID int) (*UserSummary, error) {
	// Fetch user data
	userData, err := s.client.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Fetch user's posts
	posts, err := s.client.GetPosts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}

	// Extract post titles
	postTitles := make([]string, len(posts))
	for i, post := range posts {
		postTitles[i] = post.Title
	}

	// Build summary
	summary := &UserSummary{
		UserID:      userData.ID,
		UserName:    userData.Name,
		Email:       userData.Email,
		PostCount:   len(posts),
		RecentPosts: postTitles,
		Summary:     fmt.Sprintf("User %s has written %d posts", userData.Name, len(posts)),
	}

	return summary, nil
}

// GetUserReport generates a comprehensive user report with posts and todos.
// Makes three parallel API calls for efficiency.
// Applies filtering based on request parameters.
func (s *Service) GetUserReport(ctx context.Context, userID int, req ReportRequest) (*UserReport, error) {
	// Fetch all data in parallel for efficiency
	// In a real implementation, we might use goroutines and channels
	// For simplicity, we'll do sequential calls here
	userData, err := s.client.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	posts, err := s.client.GetPosts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}

	todos, err := s.client.GetTodos(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch todos: %w", err)
	}

	// Apply business logic
	limitedPosts := limitPosts(posts, req.MaxPosts)
	stats := generateReportStats(posts, todos)
	todoGroups := groupTodos(todos, req.IncludeCompleted)
	postPreviews := postsToPreview(limitedPosts)

	report := &UserReport{
		UserID:      userData.ID,
		UserName:    userData.Name,
		Email:       userData.Email,
		Stats:       stats,
		Posts:       postPreviews,
		Todos:       todoGroups,
		GeneratedAt: formatTimestamp(),
	}

	return report, nil
}

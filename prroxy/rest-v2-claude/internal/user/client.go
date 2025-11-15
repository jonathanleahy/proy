package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	apperrors "github.com/jonathanleahy/prroxy/rest-v2/internal/common/errors"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/common/httpclient"
)

// UserClient is the interface for fetching user-related data from external API.
// This interface allows for easy mocking in tests.
type UserClient interface {
	GetUser(ctx context.Context, userID int) (*User, error)
	GetPosts(ctx context.Context, userID int) ([]Post, error)
	GetTodos(ctx context.Context, userID int) ([]Todo, error)
}

// Client handles HTTP requests to the jsonplaceholder API via proxy.
// It provides methods for fetching user data, posts, and todos.
type Client struct {
	httpClient *httpclient.Client
	baseTarget string
}

// NewClient creates a new user API client.
// baseTarget is the target API URL (e.g., "https://jsonplaceholder.typicode.com")
// Proxy configuration is handled externally via source code modification.
func NewClient(baseTarget string) *Client {
	return &Client{
		httpClient: httpclient.New(0), // Use default timeout
		baseTarget: baseTarget,
	}
}

// GetUser retrieves basic user information by ID from jsonplaceholder.
// Returns ErrNotFound if user doesn't exist.
func (c *Client) GetUser(ctx context.Context, userID int) (*User, error) {
	target := fmt.Sprintf("%s/users/%d", c.baseTarget, userID)

	resp, err := c.httpClient.Get(ctx, target)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to fetch user: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperrors.ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to read response body: %w", err))
	}

	var fullUser fullUser
	if err := json.Unmarshal(body, &fullUser); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to decode user: %w", err))
	}

	return fullUser.toUser(), nil
}

// GetPosts retrieves all posts for a given user from jsonplaceholder.
func (c *Client) GetPosts(ctx context.Context, userID int) ([]Post, error) {
	target := fmt.Sprintf("%s/posts?userId=%d", c.baseTarget, userID)

	resp, err := c.httpClient.Get(ctx, target)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to fetch posts: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to read response body: %w", err))
	}

	var posts []Post
	if err := json.Unmarshal(body, &posts); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to decode posts: %w", err))
	}

	return posts, nil
}

// GetTodos retrieves all todos for a given user from jsonplaceholder.
func (c *Client) GetTodos(ctx context.Context, userID int) ([]Todo, error) {
	target := fmt.Sprintf("%s/todos?userId=%d", c.baseTarget, userID)

	resp, err := c.httpClient.Get(ctx, target)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to fetch todos: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to read response body: %w", err))
	}

	var todos []Todo
	if err := json.Unmarshal(body, &todos); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to decode todos: %w", err))
	}

	return todos, nil
}

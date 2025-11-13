package user

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
)

// Service implements the user business logic
type Service struct {
	baseURL string
	client  *http.Client
}

// NewService creates a new user service
func NewService() *Service {
	// Base URL - will be modified by add-target.sh to include proxy prefix
	baseURL := "http://localhost:8099/proxy?target=https://jsonplaceholder.typicode.com"

	return &Service{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetUser retrieves a user by ID from the external API
func (s *Service) GetUser(userID int) (*User, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}

	// Call external API through proxy
	url := fmt.Sprintf("%s/users/%d", s.baseURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Don't set Accept-Encoding - let Go's http client handle compression automatically
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, ErrExternalServiceUnavailable
	}
	defer resp.Body.Close()

	// Handle HTTP status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Continue processing
	case http.StatusNotFound:
		return nil, ErrUserNotFound
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Handle content encoding
	var reader io.Reader = resp.Body
	switch strings.ToLower(resp.Header.Get("Content-Encoding")) {
	case "br", "brotli":
		reader = brotli.NewReader(resp.Body)
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Read response body
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse external user format
	var externalUser ExternalUser
	if err := json.Unmarshal(body, &externalUser); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	// Convert to internal user format
	user := externalUser.ToUser()

	// Validate
	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserSummary retrieves user with their posts and returns a summary
func (s *Service) GetUserSummary(userID int) (*UserSummary, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}

	// Fetch user
	userURL := fmt.Sprintf("%s/users/%d", s.baseURL, userID)
	userReq, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}

	userResp, err := s.client.Do(userReq)
	if err != nil {
		return nil, ErrExternalServiceUnavailable
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		if userResp.StatusCode == http.StatusNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("unexpected status code: %d", userResp.StatusCode)
	}

	// Handle content encoding for user
	var userReader io.Reader = userResp.Body
	switch strings.ToLower(userResp.Header.Get("Content-Encoding")) {
	case "br", "brotli":
		userReader = brotli.NewReader(userResp.Body)
	case "gzip":
		gzipReader, err := gzip.NewReader(userResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		userReader = gzipReader
	}

	userBody, err := io.ReadAll(userReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read user response: %w", err)
	}

	var externalUser ExternalUser
	if err := json.Unmarshal(userBody, &externalUser); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	user := externalUser.ToUser()

	// Fetch posts
	postsURL := fmt.Sprintf("%s/posts?userId=%d", s.baseURL, userID)
	postsReq, err := http.NewRequest("GET", postsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create posts request: %w", err)
	}

	postsResp, err := s.client.Do(postsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer postsResp.Body.Close()

	if postsResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected posts status code: %d", postsResp.StatusCode)
	}

	// Handle content encoding for posts
	var postsReader io.Reader = postsResp.Body
	switch strings.ToLower(postsResp.Header.Get("Content-Encoding")) {
	case "br", "brotli":
		postsReader = brotli.NewReader(postsResp.Body)
	case "gzip":
		gzipReader, err := gzip.NewReader(postsResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		postsReader = gzipReader
	}

	postsBody, err := io.ReadAll(postsReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read posts response: %w", err)
	}

	var posts []Post
	if err := json.Unmarshal(postsBody, &posts); err != nil {
		return nil, fmt.Errorf("failed to parse posts data: %w", err)
	}

	// Build summary
	postTitles := make([]string, len(posts))
	for i, post := range posts {
		postTitles[i] = post.Title
	}

	summary := &UserSummary{
		UserID:      user.ID,
		UserName:    user.Name,
		Email:       user.Email,
		PostCount:   len(posts),
		RecentPosts: postTitles,
		Summary:     fmt.Sprintf("User %s has written %d posts", user.Name, len(posts)),
	}

	return summary, nil
}

// GetUserReport generates a comprehensive user report with parallel API calls
func (s *Service) GetUserReport(userID int, options *ReportOptions) (*UserReport, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}

	// Set default options
	includeCompleted := true
	if options != nil && options.IncludeCompleted != nil {
		includeCompleted = *options.IncludeCompleted
	}

	// Use channels for parallel API calls
	type userResult struct {
		user *User
		err  error
	}
	type postsResult struct {
		posts []Post
		err   error
	}
	type todosResult struct {
		todos []Todo
		err   error
	}

	userChan := make(chan userResult, 1)
	postsChan := make(chan postsResult, 1)
	todosChan := make(chan todosResult, 1)

	// Fetch user in parallel
	go func() {
		user, err := s.fetchUser(userID)
		userChan <- userResult{user: user, err: err}
	}()

	// Fetch posts in parallel
	go func() {
		posts, err := s.fetchPosts(userID)
		postsChan <- postsResult{posts: posts, err: err}
	}()

	// Fetch todos in parallel
	go func() {
		todos, err := s.fetchTodos(userID)
		todosChan <- todosResult{todos: todos, err: err}
	}()

	// Wait for all results
	userRes := <-userChan
	postsRes := <-postsChan
	todosRes := <-todosChan

	// Check for errors
	if userRes.err != nil {
		return nil, userRes.err
	}
	if postsRes.err != nil {
		return nil, postsRes.err
	}
	if todosRes.err != nil {
		return nil, todosRes.err
	}

	// Apply maxPosts filter
	posts := postsRes.posts
	if options != nil && options.MaxPosts != nil && *options.MaxPosts > 0 {
		maxPosts := *options.MaxPosts
		if len(posts) > maxPosts {
			posts = posts[:maxPosts]
		}
	}

	// Calculate stats
	totalTodos := len(todosRes.todos)
	completedTodos := 0
	pendingTodos := 0
	for _, todo := range todosRes.todos {
		if todo.Completed {
			completedTodos++
		} else {
			pendingTodos++
		}
	}

	var completionRate string
	if totalTodos > 0 {
		rate := float64(completedTodos) / float64(totalTodos) * 100
		completionRate = fmt.Sprintf("%.1f%%", rate)
	} else {
		completionRate = "0.0%"
	}

	stats := ReportStats{
		TotalPosts:     len(postsRes.posts),
		TotalTodos:     totalTodos,
		CompletedTodos: completedTodos,
		PendingTodos:   pendingTodos,
		CompletionRate: completionRate,
	}

	// Format posts
	reportPosts := make([]ReportPost, len(posts))
	for i, post := range posts {
		reportPosts[i] = ReportPost{
			ID:      post.ID,
			Title:   post.Title,
			Preview: post.Body,
		}
	}

	// Format todos
	pendingTitles := []string{}
	completedTitles := []string{}
	for _, todo := range todosRes.todos {
		if todo.Completed {
			if includeCompleted {
				completedTitles = append(completedTitles, todo.Title)
			}
		} else {
			pendingTitles = append(pendingTitles, todo.Title)
		}
	}

	// Create report
	report := &UserReport{
		UserID:   userRes.user.ID,
		UserName: userRes.user.Name,
		Email:    userRes.user.Email,
		Stats:    stats,
		Posts:    reportPosts,
		Todos: ReportTodos{
			Pending:   pendingTitles,
			Completed: completedTitles,
		},
		GeneratedAt: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}

	return report, nil
}

// fetchUser is a helper to fetch user data
func (s *Service) fetchUser(userID int) (*User, error) {
	url := fmt.Sprintf("%s/users/%d", s.baseURL, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, ErrExternalServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrUserNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	switch strings.ToLower(resp.Header.Get("Content-Encoding")) {
	case "br", "brotli":
		reader = brotli.NewReader(resp.Body)
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var externalUser ExternalUser
	if err := json.Unmarshal(body, &externalUser); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return externalUser.ToUser(), nil
}

// fetchPosts is a helper to fetch posts data
func (s *Service) fetchPosts(userID int) ([]Post, error) {
	url := fmt.Sprintf("%s/posts?userId=%d", s.baseURL, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	switch strings.ToLower(resp.Header.Get("Content-Encoding")) {
	case "br", "brotli":
		reader = brotli.NewReader(resp.Body)
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var posts []Post
	if err := json.Unmarshal(body, &posts); err != nil {
		return nil, fmt.Errorf("failed to parse posts data: %w", err)
	}

	return posts, nil
}

// fetchTodos is a helper to fetch todos data
func (s *Service) fetchTodos(userID int) ([]Todo, error) {
	url := fmt.Sprintf("%s/todos?userId=%d", s.baseURL, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch todos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	switch strings.ToLower(resp.Header.Get("Content-Encoding")) {
	case "br", "brotli":
		reader = brotli.NewReader(resp.Body)
	case "gzip":
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var todos []Todo
	if err := json.Unmarshal(body, &todos); err != nil {
		return nil, fmt.Errorf("failed to parse todos data: %w", err)
	}

	return todos, nil
}

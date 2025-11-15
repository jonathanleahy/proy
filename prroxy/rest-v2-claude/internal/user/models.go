// Package user provides user management functionality.
// It handles user data retrieval from external APIs and aggregates
// related information like posts and todos.
package user

import (
	"fmt"
	"time"
)

// User represents basic user information from jsonplaceholder
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
}

// Post represents a blog post from jsonplaceholder
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// Todo represents a todo item from jsonplaceholder
type Todo struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// UserSummary represents aggregated user summary response
type UserSummary struct {
	UserID      int      `json:"userId"`
	UserName    string   `json:"userName"`
	Email       string   `json:"email"`
	PostCount   int      `json:"postCount"`
	RecentPosts []string `json:"recentPosts"`
	Summary     string   `json:"summary"`
}

// ReportRequest represents user report request parameters
type ReportRequest struct {
	IncludeCompleted bool `json:"includeCompleted"`
	MaxPosts         *int `json:"maxPosts,omitempty"`
}

// UserReport represents comprehensive user report response
type UserReport struct {
	UserID      int         `json:"userId"`
	UserName    string      `json:"userName"`
	Email       string      `json:"email"`
	Stats       ReportStats `json:"stats"`
	Posts       []PostPreview `json:"posts"`
	Todos       TodoGroups  `json:"todos"`
	GeneratedAt string      `json:"generatedAt"`
}

// ReportStats contains statistics for a user report
type ReportStats struct {
	TotalPosts     int    `json:"totalPosts"`
	TotalTodos     int    `json:"totalTodos"`
	CompletedTodos int    `json:"completedTodos"`
	PendingTodos   int    `json:"pendingTodos"`
	CompletionRate string `json:"completionRate"`
}

// PostPreview represents a post with title and preview
type PostPreview struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Preview string `json:"preview"`
}

// TodoGroups groups todos by status
type TodoGroups struct {
	Pending   []string `json:"pending"`
	Completed []string `json:"completed"`
}

// fullUser is the complete user structure from jsonplaceholder
// Used internally for API responses that include nested objects
type fullUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Address  struct {
		Street  string `json:"street"`
		Suite   string `json:"suite"`
		City    string `json:"city"`
		Zipcode string `json:"zipcode"`
		Geo     struct {
			Lat string `json:"lat"`
			Lng string `json:"lng"`
		} `json:"geo"`
	} `json:"address"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
	Company struct {
		Name        string `json:"name"`
		CatchPhrase string `json:"catchPhrase"`
		BS          string `json:"bs"`
	} `json:"company"`
}

// toUser converts a fullUser to a User by extracting only the required fields
func (fu *fullUser) toUser() *User {
	return &User{
		ID:       fu.ID,
		Name:     fu.Name,
		Username: fu.Username,
		Email:    fu.Email,
		Phone:    fu.Phone,
		Website:  fu.Website,
	}
}

// generateReportStats generates statistics from posts and todos
func generateReportStats(posts []Post, todos []Todo) ReportStats {
	completedCount := 0
	for _, todo := range todos {
		if todo.Completed {
			completedCount++
		}
	}

	completionRate := "0.0"
	if len(todos) > 0 {
		rate := (float64(completedCount) / float64(len(todos))) * 100
		completionRate = fmt.Sprintf("%.1f%%", rate)
	}

	return ReportStats{
		TotalPosts:     len(posts),
		TotalTodos:     len(todos),
		CompletedTodos: completedCount,
		PendingTodos:   len(todos) - completedCount,
		CompletionRate: completionRate,
	}
}

// limitPosts limits the number of posts if maxPosts is specified
func limitPosts(posts []Post, maxPosts *int) []Post {
	if maxPosts != nil && *maxPosts > 0 && *maxPosts < len(posts) {
		return posts[:*maxPosts]
	}
	return posts
}

// groupTodos separates todos into pending and completed groups
func groupTodos(todos []Todo, includeCompleted bool) TodoGroups {
	groups := TodoGroups{
		Pending:   make([]string, 0),
		Completed: make([]string, 0),
	}

	for _, todo := range todos {
		if todo.Completed {
			if includeCompleted {
				groups.Completed = append(groups.Completed, todo.Title)
			}
		} else {
			groups.Pending = append(groups.Pending, todo.Title)
		}
	}

	return groups
}

// postsToPreview converts posts to post previews
func postsToPreview(posts []Post) []PostPreview {
	previews := make([]PostPreview, len(posts))
	for i, post := range posts {
		previews[i] = PostPreview{
			ID:      post.ID,
			Title:   post.Title,
			Preview: post.Body,
		}
	}
	return previews
}

// formatTimestamp returns current time in ISO 8601 format
func formatTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

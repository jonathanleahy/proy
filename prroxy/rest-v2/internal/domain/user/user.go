package user

import "time"

// User represents a simplified user from JSONPlaceholder API
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
}

// ExternalUser represents the full user structure from JSONPlaceholder
type ExternalUser struct {
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
		Bs          string `json:"bs"`
	} `json:"company"`
}

// ToUser converts ExternalUser to simplified User
func (eu *ExternalUser) ToUser() *User {
	return &User{
		ID:       eu.ID,
		Name:     eu.Name,
		Username: eu.Username,
		Email:    eu.Email,
		Phone:    eu.Phone,
		Website:  eu.Website,
	}
}

// NewUser creates a new User instance
func NewUser(id int, name, username, email, phone, website string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Username: username,
		Email:    email,
		Phone:    phone,
		Website:  website,
	}
}

// Validate checks if the user data is valid
func (u *User) Validate() error {
	if u.ID <= 0 {
		return ErrInvalidUserID
	}
	if u.Name == "" {
		return ErrInvalidUserData
	}
	if u.Email == "" {
		return ErrInvalidUserData
	}
	return nil
}

// CreatedAt returns a mock creation timestamp
// In a real system, this would come from the database
func (u *User) CreatedAt() time.Time {
	return time.Now()
}

// Post represents a post from JSONPlaceholder API
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// UserSummary represents aggregated user data with posts
type UserSummary struct {
	UserID      int      `json:"userId"`
	UserName    string   `json:"userName"`
	Email       string   `json:"email"`
	PostCount   int      `json:"postCount"`
	RecentPosts []string `json:"recentPosts"`
	Summary     string   `json:"summary"`
}

// Todo represents a todo item from JSONPlaceholder API
type Todo struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// ReportOptions represents optional parameters for user report generation
type ReportOptions struct {
	IncludeCompleted *bool `json:"includeCompleted"`
	MaxPosts         *int  `json:"maxPosts"`
}

// ReportStats represents statistical data in the user report
type ReportStats struct {
	TotalPosts     int    `json:"totalPosts"`
	TotalTodos     int    `json:"totalTodos"`
	CompletedTodos int    `json:"completedTodos"`
	PendingTodos   int    `json:"pendingTodos"`
	CompletionRate string `json:"completionRate"`
}

// ReportPost represents a post in the user report
type ReportPost struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Preview string `json:"preview"`
}

// ReportTodos represents todos categorized by status
type ReportTodos struct {
	Pending   []string `json:"pending"`
	Completed []string `json:"completed"`
}

// UserReport represents the comprehensive user report
type UserReport struct {
	UserID      int          `json:"userId"`
	UserName    string       `json:"userName"`
	Email       string       `json:"email"`
	Stats       ReportStats  `json:"stats"`
	Posts       []ReportPost `json:"posts"`
	Todos       ReportTodos  `json:"todos"`
	GeneratedAt string       `json:"generatedAt"`
}

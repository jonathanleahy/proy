package person

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	apperrors "github.com/jonathanleahy/prroxy/rest-v2/internal/common/errors"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/common/httpclient"
)

// PersonClient is the interface for fetching person-related data from external API.
// This interface allows for easy mocking in tests.
type PersonClient interface {
	FindPerson(ctx context.Context, surname, dob string) (*Person, error)
	FindPeople(ctx context.Context, surname, dob string) ([]Person, error)
}

// Client handles HTTP requests to the rest-external-user API via proxy.
// It provides methods for finding people by surname and/or date of birth.
type Client struct {
	httpClient *httpclient.Client
	baseTarget string
}

// NewClient creates a new person API client.
// baseTarget is the target API URL (e.g., "http://0.0.0.0:3006")
// Proxy configuration is handled externally via source code modification.
func NewClient(baseTarget string) *Client {
	return &Client{
		httpClient: httpclient.New(0), // Use default timeout
		baseTarget: baseTarget,
	}
}

// FindPerson finds a single person by exact match of surname and date of birth.
// Returns ErrNotFound if person doesn't exist.
func (c *Client) FindPerson(ctx context.Context, surname, dob string) (*Person, error) {
	// Build target URL with query parameters
	targetURL, err := c.buildPersonURL(surname, dob)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to build URL: %w", err))
	}

	resp, err := c.httpClient.Get(ctx, targetURL)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to fetch person: %w", err))
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

	var p Person
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to decode person: %w", err))
	}

	return &p, nil
}

// FindPeople finds people by surname OR date of birth (partial search).
// At least one parameter must be provided.
// Returns empty array if no matches found.
func (c *Client) FindPeople(ctx context.Context, surname, dob string) ([]Person, error) {
	// Build target URL with query parameters
	targetURL, err := c.buildPersonURL(surname, dob)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to build URL: %w", err))
	}

	resp, err := c.httpClient.Get(ctx, targetURL)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to fetch people: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to read response body: %w", err))
	}

	var people []Person
	if err := json.Unmarshal(body, &people); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, fmt.Errorf("failed to decode people: %w", err))
	}

	return people, nil
}

// buildPersonURL constructs the target URL with query parameters for person lookup.
// Handles proxy wrapping by checking if baseTarget already contains proxy URL.
func (c *Client) buildPersonURL(surname, dob string) (string, error) {
	// Check if baseTarget is already wrapped by proxy (contains "proxy?target=")
	if containsProxyWrapper(c.baseTarget) {
		// Extract the actual target from the proxy URL
		actualTarget, err := extractTargetFromProxy(c.baseTarget)
		if err != nil {
			return "", err
		}

		// Build the full URL with the actual target
		targetURL, err := url.Parse(actualTarget + "/person")
		if err != nil {
			return "", err
		}

		query := targetURL.Query()
		if surname != "" {
			query.Set("surname", surname)
		}
		if dob != "" {
			query.Set("dob", dob)
		}
		targetURL.RawQuery = query.Encode()

		// Re-wrap with proxy, properly URL-encoding the target
		return wrapWithProxy(targetURL.String())
	}

	// Direct URL (no proxy wrapping)
	targetURL, err := url.Parse(c.baseTarget + "/person")
	if err != nil {
		return "", err
	}

	query := targetURL.Query()
	if surname != "" {
		query.Set("surname", surname)
	}
	if dob != "" {
		query.Set("dob", dob)
	}
	targetURL.RawQuery = query.Encode()

	return targetURL.String(), nil
}

// containsProxyWrapper checks if a URL contains proxy wrapping
func containsProxyWrapper(urlStr string) bool {
	return url.QueryEscape("proxy?target=") != "" &&
		(containsString(urlStr, "localhost:8099/proxy?target=") ||
		 containsString(urlStr, "0.0.0.0:8099/proxy?target="))
}

// containsString is a simple string contains check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring checks if substr is in s
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// extractTargetFromProxy extracts the target URL from a proxy-wrapped URL
func extractTargetFromProxy(proxyURL string) (string, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return "", err
	}
	target := u.Query().Get("target")
	if target == "" {
		return "", fmt.Errorf("no target parameter in proxy URL")
	}
	// Decode the URL-encoded target
	decoded, err := url.QueryUnescape(target)
	if err != nil {
		return "", fmt.Errorf("failed to decode target: %w", err)
	}
	return decoded, nil
}

// wrapWithProxy wraps a target URL with proxy URL and proper encoding
func wrapWithProxy(targetURL string) (string, error) {
	// Build proxy URL with properly encoded target parameter
	proxyBase := "http://localhost:8099/proxy"
	encoded := url.QueryEscape(targetURL)
	return proxyBase + "?target=" + encoded, nil
}

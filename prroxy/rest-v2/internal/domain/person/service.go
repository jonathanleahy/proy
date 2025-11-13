package person

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// Service handles person-related business logic
type Service struct {
	httpClient         *http.Client
	proxyURL           string
	externalServiceURL string
}

// NewService creates a new person service
func NewService() *Service {
	proxyURL := os.Getenv("PROXY_URL")
	externalServiceTarget := "http://0.0.0.0:3006"

	// Store the base proxy URL without the target parameter
	// We'll add the properly encoded target for each request
	var baseProxyURL string
	if proxyURL != "" {
		baseProxyURL = proxyURL
	} else {
		baseProxyURL = "http://localhost:8099/proxy"
	}

	return &Service{
		httpClient:         &http.Client{},
		proxyURL:           baseProxyURL,
		externalServiceURL: externalServiceTarget,
	}
}

// FindPerson finds a person by surname and date of birth
// Returns the person if found, nil if not found
func (s *Service) FindPerson(surname, dob string) (*Person, error) {
	// Build query parameters for the external service
	queryParams := url.Values{}
	queryParams.Add("surname", surname)
	queryParams.Add("dob", dob)

	// Build the complete target URL (external service endpoint + query params)
	targetEndpoint := fmt.Sprintf("%s/person?%s", s.externalServiceURL, queryParams.Encode())

	// URL-encode the entire target URL to pass it as a parameter to the proxy
	encodedTarget := url.QueryEscape(targetEndpoint)

	// Build the final proxy URL with the properly encoded target parameter
	proxyURL := fmt.Sprintf("%s?target=%s", s.proxyURL, encodedTarget)

	// Make HTTP request
	resp, err := s.httpClient.Get(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch person: %w", err)
	}
	defer resp.Body.Close()

	// Handle 404 as person not found
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// When both surname and dob are provided, external service returns a single object
	// Try to unmarshal as a single Person first
	var person Person
	if err := json.Unmarshal(body, &person); err == nil {
		return &person, nil
	}

	// If that fails, try as an array (for backwards compatibility)
	var people []Person
	if err := json.Unmarshal(body, &people); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// If we got results, return the first one
	if len(people) > 0 {
		return &people[0], nil
	}

	// No results found
	return nil, nil
}

// FindPeople searches for people by surname or dob (partial search)
// At least one parameter must be provided
func (s *Service) FindPeople(surname, dob *string) ([]Person, error) {
	// Build query parameters for the external service
	queryParams := url.Values{}
	if surname != nil && *surname != "" {
		queryParams.Add("surname", *surname)
	}
	if dob != nil && *dob != "" {
		queryParams.Add("dob", *dob)
	}

	// Build the complete target URL (external service endpoint + query params)
	targetEndpoint := fmt.Sprintf("%s/person?%s", s.externalServiceURL, queryParams.Encode())

	// URL-encode the entire target URL to pass it as a parameter to the proxy
	encodedTarget := url.QueryEscape(targetEndpoint)

	// Build the final proxy URL with the properly encoded target parameter
	proxyURL := fmt.Sprintf("%s?target=%s", s.proxyURL, encodedTarget)

	// Make HTTP request
	resp, err := s.httpClient.Get(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch people: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse array of people
	var people []Person
	if err := json.Unmarshal(body, &people); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return people, nil
}

package health

// Service implements the health check business logic
type Service struct {
	version string
}

// NewService creates a new health service
func NewService(version string) *Service {
	return &Service{
		version: version,
	}
}

// GetHealth returns the current health status
func (s *Service) GetHealth() (*Health, error) {
	return NewHealth(s.version), nil
}

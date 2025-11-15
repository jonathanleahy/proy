package person

import (
	"context"
	"fmt"
)

// PersonService is the interface for person business logic.
// This interface allows for easy mocking in handler tests.
type PersonService interface {
	FindPerson(ctx context.Context, surname, dob string) (*Person, error)
	FindPeople(ctx context.Context, surname, dob string) ([]Person, error)
}

// Service handles person-related business logic.
// It coordinates between the client and presents a unified interface
// to the HTTP layer.
type Service struct {
	client PersonClient
}

// NewService creates a new person service.
func NewService(client PersonClient) *Service {
	return &Service{
		client: client,
	}
}

// FindPerson finds a single person by exact match of surname and date of birth.
func (s *Service) FindPerson(ctx context.Context, surname, dob string) (*Person, error) {
	person, err := s.client.FindPerson(ctx, surname, dob)
	if err != nil {
		return nil, fmt.Errorf("failed to find person: %w", err)
	}
	return person, nil
}

// FindPeople finds people by surname OR date of birth (partial search).
func (s *Service) FindPeople(ctx context.Context, surname, dob string) ([]Person, error) {
	people, err := s.client.FindPeople(ctx, surname, dob)
	if err != nil {
		return nil, fmt.Errorf("failed to find people: %w", err)
	}
	return people, nil
}

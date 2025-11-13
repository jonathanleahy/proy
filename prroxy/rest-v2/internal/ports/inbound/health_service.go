package inbound

import "github.com/jonathanleahy/prroxy/rest-v2/internal/domain/health"

// HealthService defines the contract for health check operations
// This is an inbound port - what the application offers to the outside world
type HealthService interface {
	GetHealth() (*health.Health, error)
}

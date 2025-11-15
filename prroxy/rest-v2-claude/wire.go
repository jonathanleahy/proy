//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/person"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/user"
	"go.uber.org/zap"
)

// InitializeServer wires up all dependencies and returns configured handlers.
func InitializeServer(proxyURL, jsonPlaceholderTarget, externalUserTarget string) (*user.Handler, *person.Handler, *zap.Logger, error) {
	wire.Build(
		// User domain
		user.NewClient,
		user.NewService,
		user.NewHandler,

		// Person domain
		person.NewClient,
		person.NewService,
		person.NewHandler,

		// Logger
		zap.NewProduction,
	)
	return nil, nil, nil, nil
}

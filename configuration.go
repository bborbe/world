package world

import "context"

//go:generate counterfeiter -o mocks/configuration.go --fake-name Configuration . Configuration
type Configuration interface {
	Children() []Configuration
	Applier() (Applier, error)
	Validate(ctx context.Context) error
}

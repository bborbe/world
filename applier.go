package world

import "context"

//go:generate counterfeiter -o mocks/applier.go --fake-name Applier . Applier
type Applier interface {
	Satisfied(ctx context.Context) (bool, error)
	Apply(ctx context.Context) error
	Validate(ctx context.Context) error
}

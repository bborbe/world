package k8s

import (
	"context"

	"github.com/pkg/errors"
)

type Kind string

func (k Kind) Validate(ctx context.Context) error {
	if k == "" {
		return errors.New("Kind missing")
	}
	return nil
}

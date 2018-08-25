package k8s

import (
	"context"

	"github.com/pkg/errors"
)

type ApiVersion string

func (a ApiVersion) Validate(ctx context.Context) error {
	if a == "" {
		return errors.New("ApiVersion missing")
	}
	return nil
}

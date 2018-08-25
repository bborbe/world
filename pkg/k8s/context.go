package k8s

import (
	"context"

	"github.com/pkg/errors"
)

type Context string

func (c Context) String() string {
	return string(c)
}

func (w Context) Validate(ctx context.Context) error {
	if w == "" {
		return errors.New("Context missing")
	}
	return nil
}

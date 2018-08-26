package k8s

import (
	"context"

	"github.com/pkg/errors"
)

type Context string

func (c Context) String() string {
	return string(c)
}

func (c Context) Validate(ctx context.Context) error {
	if c == "" {
		return errors.New("Context missing")
	}
	return nil
}

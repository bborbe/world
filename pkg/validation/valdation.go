package validation

import (
	"context"

	"github.com/pkg/errors"
)

type Validator interface {
	Validate(ctx context.Context) error
}

func Validate(ctx context.Context, validators ...Validator) error {
	for _, v := range validators {
		if v == nil {
			return errors.New("validatior nil")
		}
		if err := v.Validate(ctx); err != nil {
			return errors.Wrap(err, "validate failed")
		}
	}
	return nil
}
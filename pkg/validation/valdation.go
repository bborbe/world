package validation

import (
	"context"
)

type Validator interface {
	Validate(ctx context.Context) error
}

func Validate(ctx context.Context, validators ...Validator) error {
	for _, v := range validators {
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

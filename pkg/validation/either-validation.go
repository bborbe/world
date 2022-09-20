// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/pkg/errors"
)

func EitherValidation(a Validator, b Validator) Validator {
	return ValidatorFunc(func(ctx context.Context) error {
		if a != nil && b == nil {
			if err := a.Validate(ctx); err != nil {
				return err
			}
			return nil
		}
		if a == nil && b != nil {
			if err := b.Validate(ctx); err != nil {
				return err
			}
			return nil
		}
		return errors.Errorf("only one validator (%#v %#v) should be filled", a, b)
	})
}

// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/pkg/errors"
)

type Validator interface {
	Validate(ctx context.Context) error
}

type ValidatorFunc func(ctx context.Context) error

func (v ValidatorFunc) Validate(ctx context.Context) error {
	return v(ctx)
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

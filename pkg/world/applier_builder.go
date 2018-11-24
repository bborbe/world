// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package world

import (
	"context"

	"github.com/pkg/errors"
)

type ApplierBuildFunc func(ctx context.Context) (Applier, error)

type ApplierBuilder struct {
	Build ApplierBuildFunc
}

func (a *ApplierBuilder) Satisfied(ctx context.Context) (bool, error) {
	applier, err := a.Build(ctx)
	if err != nil {
		return false, errors.Wrap(err, "build applier failed")
	}
	return applier.Satisfied(ctx)
}

func (a *ApplierBuilder) Apply(ctx context.Context) error {
	applier, err := a.Build(ctx)
	if err != nil {
		return errors.Wrap(err, "build applier failed")
	}
	return applier.Apply(ctx)
}

func (a *ApplierBuilder) Validate(ctx context.Context) error {
	if a.Build == nil {
		return errors.New("Build missing")
	}
	return nil
}

// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package world

import "context"

//go:generate counterfeiter -o mocks/applier.go --fake-name Applier . Applier
type Applier interface {
	Satisfied(ctx context.Context) (bool, error)
	Apply(ctx context.Context) error
	Validate(ctx context.Context) error
}

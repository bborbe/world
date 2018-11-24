// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package world

import "context"

//go:generate counterfeiter -o mocks/configuration.go --fake-name Configuration . Configuration
type Configuration interface {
	Children() []Configuration
	Applier() (Applier, error)
	Validate(ctx context.Context) error
}

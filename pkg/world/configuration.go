// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package world

import "context"

type Configurations []Configuration

//go:generate go run -mod=vendor github.com/maxbrunsfeld/counterfeiter/v6 -o ../../mocks/configuration.go --fake-name Configuration . Configuration
type Configuration interface {
	Children(ctx context.Context) (Configurations, error)
	Applier() (Applier, error)
	Validate(ctx context.Context) error
}

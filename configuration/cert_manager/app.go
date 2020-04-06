// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cert_manager

import (
	"context"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type App struct {
	Context k8s.Context
}

func (a *App) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		a.Context,
	)
}
func (a *App) Children() []world.Configuration {
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: a.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "cert-manager",
					Name:      "cert-manager",
				},
			},
		},
	}
}
func (a *App) Applier() (world.Applier, error) {
	return nil, nil
}

// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"context"

	"github.com/bborbe/world/pkg/build"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type AuthHttpProxy struct {
	Image docker.Image
}

func (a *AuthHttpProxy) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		a.Image,
	)
}

func (a *AuthHttpProxy) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		build.Configuration(
			&docker.GolangBuilder{
				Name:            "auth-http-proxy",
				GitRepo:         "https://github.com/bborbe/auth-http-proxy.git",
				SourceDirectory: "github.com/bborbe/auth-http-proxy",
				Package:         "github.com/bborbe/auth-http-proxy",
				Image:           a.Image,
			},
		),
	}, nil
}

func (a *AuthHttpProxy) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: a.Image,
	}, nil
}

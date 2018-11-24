// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type AuthHttpProxy struct {
	Image docker.Image
}

func (t *AuthHttpProxy) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Image,
	)
}

func (i *AuthHttpProxy) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.GolangBuilder{
				Name:            "auth-http-proxy",
				GitRepo:         "https://github.com/bborbe/auth-http-proxy.git",
				SourceDirectory: "github.com/bborbe/auth-http-proxy",
				Package:         "github.com/bborbe/auth-http-proxy",
				Image:           i.Image,
			},
		},
	}
}

func (i *AuthHttpProxy) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: i.Image,
	}, nil
}

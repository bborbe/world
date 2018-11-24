// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package world_test

import (
	"context"
	"testing"

	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/world/configuration"
	"github.com/bborbe/world/pkg/secret"
	"github.com/bborbe/world/pkg/world"
)

func TestValidate(t *testing.T) {
	ctx := context.Background()
	builder := world.Builder{
		Configuration: &configuration.World{
			TeamvaultSecrets: &secret.Teamvault{
				TeamvaultConnector: connector.NewDummy(),
			},
		},
	}
	r, err := builder.Build(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Validate(ctx); err != nil {
		t.Fatal(err)
	}

}

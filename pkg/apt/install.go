// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apt

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
)

type Package string

type Install struct {
	SSH     *ssh.SSH
	Package string
}

func (i *Install) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (i *Install) Apply(ctx context.Context) error {
	return i.SSH.RunCommand(ctx, fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install --quiet --yes --no-install-recommends %s", i.Package))
}

func (i *Install) Validate(ctx context.Context) error {
	if i.Package == "" {
		return errors.New("Package missing")
	}
	return validation.Validate(
		ctx,
		i.SSH,
	)
}

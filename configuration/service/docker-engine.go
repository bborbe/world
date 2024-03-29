// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type DockerEngine struct {
	SSH *ssh.SSH
}

func (d *DockerEngine) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     d.SSH,
			Command: "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "software-properties-common",
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     d.SSH,
			Command: `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     d.SSH,
			Command: `apt-get --quiet --yes update`,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     d.SSH,
			Command: `apt-get --quiet --yes --no-install-recommends install docker-ce docker-ce-cli containerd.io`,
		}),
	}, nil
}

func (d *DockerEngine) Applier() (world.Applier, error) {
	return &remote.ServiceStart{
		SSH:  d.SSH,
		Name: "docker",
	}, nil
}

func (d *DockerEngine) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
	)
}

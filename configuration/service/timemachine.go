// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type TimeMachineConfigs []TimeMachineConfig

func (t TimeMachineConfigs) Validate(ctx context.Context) error {
	for _, tt := range t {
		if err := tt.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type TimeMachineConfig struct {
	Username string
	Path     string
	Password deployer.SecretValue
	Size     int
}

type TimeMachine struct {
	SSH     *ssh.SSH
	Configs TimeMachineConfigs
}

func (t TimeMachineConfig) Validate(ctx context.Context) error {
	if t.Username == "" {
		return errors.Errorf("username missing")
	}
	if t.Path == "" {
		return errors.Errorf("path missing")
	}
	if t.Size <= 0 {
		return errors.Errorf("size missing")
	}
	if err := t.Password.Validate(ctx); err != nil {
		return err
	}
	return nil
}

func (d *TimeMachine) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "netatalk",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "avahi-daemon",
		}),
		&remote.File{
			SSH:  d.SSH,
			Path: file.Path("/etc/netatalk/AppleVolumes.default"),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				buf := &bytes.Buffer{}
				fmt.Fprintf(buf, ":DEFAULT: options:upriv,usedots\n")
				for _, config := range d.Configs {
					fmt.Fprintf(buf, "%s \"%s\" options:tm volsizelimit:%d allow:%s\n", config.Path, config.Username, config.Size, config.Username)
				}
				return buf.Bytes(), nil
			}),
			User:  "root",
			Group: "root",
			Perm:  0644,
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     d.SSH,
			Command: "systemctl daemon-reload",
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.IptablesAllowInput{
			SSH:      d.SSH,
			Port:     network.PortStatic(548),
			Protocol: network.TCP,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.ServiceStart{
			SSH:  d.SSH,
			Name: "netatalk",
		}),
	}, nil
}

func (d *TimeMachine) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *TimeMachine) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
		d.Configs,
	)
}

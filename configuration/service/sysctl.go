// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type SysctlOption struct {
	Option string
	Value  string
}

func (s SysctlOption) Validate(ctx context.Context) error {
	if s.Option == "" {
		return errors.New("option missing")
	}
	return nil
}

type SysctlOptions []SysctlOption

func (s SysctlOptions) Validate(ctx context.Context) error {
	for _, option := range s {
		if err := option.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type Sysctl struct {
	SSH     *ssh.SSH
	Options SysctlOptions
}

func (s *Sysctl) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
		s.Options,
	)
}

func (s *Sysctl) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		&remote.File{
			SSH:  s.SSH,
			Path: file.Path("/etc/sysctl.d/60-custom.conf"),
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				buf := &bytes.Buffer{}
				for _, option := range s.Options {
					fmt.Fprintf(buf, "%s = %s\n", option.Option, option.Value)
				}
				return buf.Bytes(), nil
			}),
			User:  "root",
			Group: "root",
			Perm:  0664,
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     s.SSH,
			Command: "systemctl restart systemd-sysctl",
		}),
	}, nil
}

func (s *Sysctl) Applier() (world.Applier, error) {
	return nil, nil
}

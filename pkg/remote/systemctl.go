// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/ssh"
)

type SystemCtl struct {
	SSH  *ssh.SSH
	Name ServiceName
}

func (s *SystemCtl) ServiceRunning(ctx context.Context) (bool, error) {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl status -- %s", s.Name)) == nil, nil
}

func (s *SystemCtl) ServiceNotRunning(ctx context.Context) (bool, error) {
	running, err := s.ServiceRunning(ctx)
	return !running, err
}

func (s *SystemCtl) StartService(ctx context.Context) error {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl start -- %s", s.Name))
}

func (s *SystemCtl) StopService(ctx context.Context) error {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl stop -- %s", s.Name))
}

func (s *SystemCtl) ServiceEnabled(ctx context.Context) (bool, error) {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl is-enabled -- %s", s.Name)) == nil, nil
}

func (s *SystemCtl) ServiceEnable(ctx context.Context) error {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl enable -- %s", s.Name))
}

func (s *SystemCtl) ServiceDisable(ctx context.Context) error {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl disable -- %s", s.Name))
}

func (s *SystemCtl) ServiceDisabled(ctx context.Context) (bool, error) {
	enabled, err := s.ServiceEnabled(ctx)
	return !enabled, err
}

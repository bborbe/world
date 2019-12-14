// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type ServiceStop struct {
	SSH  *ssh.SSH
	Name ServiceName
}

func (s *ServiceStop) Satisfied(ctx context.Context) (bool, error) {
	systemCtl := &SystemCtl{
		SSH:  s.SSH,
		Name: s.Name,
	}

	running, err := systemCtl.ServiceDisabled(ctx)
	if err != nil {
		return false, errors.Wrap(err, "check service running failed")
	}
	enabled, err := systemCtl.ServiceNotRunning(ctx)
	if err != nil {
		return false, errors.Wrap(err, "check service enabled failed")
	}
	return running && enabled, nil
}

func (s *ServiceStop) Apply(ctx context.Context) error {
	systemCtl := &SystemCtl{
		SSH:  s.SSH,
		Name: s.Name,
	}

	if err := systemCtl.StopService(ctx); err != nil {
		return errors.Wrap(err, "stop service failed")
	}

	if err := systemCtl.ServiceDisable(ctx); err != nil {
		return errors.Wrap(err, "enable service failed")
	}
	return nil
}

func (s *ServiceStop) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
		s.Name,
	)
}

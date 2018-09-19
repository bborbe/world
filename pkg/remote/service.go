package remote

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type ServiceName string

func (h ServiceName) Validate(ctx context.Context) error {
	if h == "" {
		return errors.New("ServiceName missing")
	}
	return nil
}

type Service struct {
	SSH ssh.SSH

	Name ServiceName
}

func (f *Service) Satisfied(ctx context.Context) (bool, error) {
	running, err := f.ServiceRunning(ctx)
	if err != nil {
		return false, errors.Wrap(err, "check service running failed")
	}
	enabled, err := f.ServiceRunning(ctx)
	if err != nil {
		return false, errors.Wrap(err, "check service enabled failed")
	}
	return running && enabled, nil
}

func (f *Service) Apply(ctx context.Context) error {

	if err := f.StartService(ctx); err != nil {
		return errors.Wrap(err, "start service failed")
	}

	if err := f.ServiceEnable(ctx); err != nil {
		return errors.Wrap(err, "enable service failed")
	}
	return nil
}

func (f *Service) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.SSH,
		f.Name,
	)
}

func (s *Service) ServiceRunning(ctx context.Context) (bool, error) {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl status -- %s", s.Name)) == nil, nil
}

func (s *Service) StartService(ctx context.Context) error {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl start -- %s", s.Name))
}

func (s *Service) StopService(ctx context.Context) error {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl stop -- %s", s.Name))
}

func (s *Service) ServiceEnabled(ctx context.Context) (bool, error) {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl is-enabled -- %s", s.Name)) == nil, nil
}

func (s *Service) ServiceEnable(ctx context.Context) error {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl enable -- %s", s.Name))
}

func (s *Service) ServiceDisable(ctx context.Context) error {
	return s.SSH.RunCommand(ctx, fmt.Sprintf("systemctl enable -- %s", s.Name))
}

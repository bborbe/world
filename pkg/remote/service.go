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
	SSH  ssh.SSH
	Name ServiceName
}

func (s *Service) Satisfied(ctx context.Context) (bool, error) {
	running, err := s.ServiceRunning(ctx)
	if err != nil {
		return false, errors.Wrap(err, "check service running failed")
	}
	enabled, err := s.ServiceRunning(ctx)
	if err != nil {
		return false, errors.Wrap(err, "check service enabled failed")
	}
	return running && enabled, nil
}

func (s *Service) Apply(ctx context.Context) error {
	if err := s.StartService(ctx); err != nil {
		return errors.Wrap(err, "start service failed")
	}

	if err := s.ServiceEnable(ctx); err != nil {
		return errors.Wrap(err, "enable service failed")
	}
	return nil
}

func (s *Service) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
		s.Name,
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

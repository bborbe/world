package service

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/configuration"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Service struct {
	SSH     ssh.SSH
	Name    remote.ServiceName
	Content remote.ServiceContent
}

func (s *Service) Children() []world.Configuration {
	return []world.Configuration{
		configuration.New().WithApplier(&remote.File{
			SSH:     s.SSH,
			Path:    fmt.Sprintf("/etc/systemd/system/%s.service", s.Name),
			Content: s.Content,
			User:    "root",
			Group:   "root",
			Perm:    0664,
		}),
		configuration.New().WithApplier(&remote.Command{
			SSH:     s.SSH,
			Command: "systemctl daemon-reload",
		}),
		configuration.New().WithApplier(&remote.Service{
			SSH:  s.SSH,
			Name: s.Name,
		}),
	}
}

func (s *Service) Applier() (world.Applier, error) {
	return nil, nil
}

func (s *Service) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.SSH,
		s.Name,
		s.Content,
	)
}
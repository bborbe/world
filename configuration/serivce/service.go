package service

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Service struct {
	SSH     ssh.SSH
	Name    remote.ServiceName
	Content remote.HasContent
}

func (s *Service) Children() []world.Configuration {
	return []world.Configuration{
		&File{
			SSH:     s.SSH,
			Path:    remote.Path(fmt.Sprintf("/etc/systemd/system/%s.service", s.Name)),
			Content: s.Content,
			User:    "root",
			Group:   "root",
			Perm:    0664,
		},
		world.NewConfiguraionBuilder().WithApplier(&remote.Command{
			SSH:     s.SSH,
			Command: "systemctl daemon-reload",
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Service{
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
	)
}

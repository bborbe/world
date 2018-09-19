package server

import (
	"context"

	"github.com/bborbe/world/configuration/serivce"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Sun struct {
}

func (r *Sun) Children() []world.Configuration {
	ssh := ssh.SSH{
		Host:           "pn.benjamin-borbe.de:22",
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return []world.Configuration{
		&service.Kubernetes{
			SSH: ssh,
		},
	}
}

func (r *Sun) Applier() (world.Applier, error) {
	return nil, nil
}

func (r *Sun) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
	)
}

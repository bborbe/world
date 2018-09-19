package service

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Kubernetes struct {
	SSH ssh.SSH
}

func (k *Kubernetes) Children() []world.Configuration {
	version := docker.Tag("v1.11.2")
	return []world.Configuration{
		&Etcd{
			SSH: k.SSH,
		},
		&Kubelet{
			SSH:     k.SSH,
			Version: version,
		},
	}
}

func (k *Kubernetes) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *Kubernetes) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.SSH,
	)
}

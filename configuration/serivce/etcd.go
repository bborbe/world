package service

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Etcd struct {
	SSH ssh.SSH
}

func (e *Etcd) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/etcd",
		Tag:        "v3.3.1",
	}
	return []world.Configuration{
		&build.Etcd{
			Image: image,
		},
		&Directory{
			SSH:   e.SSH,
			Path:  "/var/lib/etcd",
			User:  "root",
			Group: "root",
			Perm:  0755,
		},
		&Docker{
			SSH:     e.SSH,
			Name:    "etcd",
			Memory:  512,
			Ports:   []int{2379, 2380},
			Volumes: []string{"/var/lib/etcd:/var/lib/etcd"},
			Image:   image,
			Command: "/usr/local/bin/etcd",
			Args: []string{
				"--advertise-client-urls http://172.16.22.10:2379",
				"--data-dir=/var/lib/etcd",
				"--initial-advertise-peer-urls http://172.16.22.10:2380",
				"--initial-cluster kubernetes=http://172.16.22.10:2380",
				"--initial-cluster-state new",
				"--initial-cluster-token cluster-fire",
				"--listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001",
				"--listen-peer-urls http://0.0.0.0:2380",
				"--name kubernetes",
			},
		},
	}
}

func (e *Etcd) Applier() (world.Applier, error) {
	return nil, nil
}

func (e *Etcd) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		e.SSH,
	)
}

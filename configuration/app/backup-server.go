// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type BackupServer struct {
	Context k8s.Context
}

func (b *BackupServer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Context,
	)
}

func (b *BackupServer) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/backup-rsync-server",
		Tag:        "1.1.0",
	}
	port := deployer.Port{
		Port:     22,
		HostPort: 2222,
		Name:     "ssh",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: b.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "backup",
					Name:      "backup",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   b.Context,
			Namespace: "backup",
			Name:      "rsync",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "backup",
					Image: image,
					Requirement: &build.BackupRsyncServer{
						Image: image,
					},
					Ports: []deployer.Port{port},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "1000m",
							Memory: "200Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "100Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     "backup",
							Path:     "/data",
							ReadOnly: true,
						},
						{
							Name:     "ssh",
							Path:     "/etc/ssh",
							ReadOnly: true,
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "backup",
					Host: k8s.PodVolumeHost{
						Path: "/data",
					},
				},
				{
					Name: "ssh",
					Host: k8s.PodVolumeHost{
						Path: "/data/backup-ssh",
					},
				},
			},
		},
	}
}

func (b *BackupServer) Applier() (world.Applier, error) {
	return nil, nil
}

// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"time"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Webdav struct {
	Context  k8s.Context
	Domains  k8s.IngressHosts
	Password deployer.SecretValue
}

func (w *Webdav) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Context,
		w.Domains,
		w.Password,
	)
}

func (w *Webdav) Children() []world.Configuration {
	version := "1.0.2"
	image := docker.Image{
		Repository: "bborbe/webdav",
		Tag:        docker.TagWithTime(version, time.Now()),
	}
	port := deployer.Port{
		Port:     80,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: w.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "webdav",
					Name:      "webdav",
				},
			},
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.SecretApplier{
				Context:   w.Context,
				Namespace: "webdav",
				Name:      "webdav",
				Secrets: deployer.Secrets{
					"password": w.Password,
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   w.Context,
			Namespace: "webdav",
			Name:      "webdav",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "webdav",
					Image: image,
					Requirement: &build.Webdav{
						GitBranch: docker.GitBranch(version),
						Image:     image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "50m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Ports: []deployer.Port{port},
					Env: []k8s.Env{
						{
							Name:  "WEBDAV_USERNAME",
							Value: "bborbe",
						},
						{
							Name: "WEBDAV_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "password",
									Name: "webdav",
								},
							},
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "webdav",
							Path: "/data",
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						FailureThreshold:    3,
						InitialDelaySeconds: 30,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: port.Port,
						},
						FailureThreshold:    1,
						InitialDelaySeconds: 10,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						TimeoutSeconds:      5,
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "webdav",
					Host: k8s.PodVolumeHost{
						Path: "/data/webdav",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   w.Context,
			Namespace: "webdav",
			Name:      "webdav",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   w.Context,
			Namespace: "webdav",
			Name:      "webdav",
			Port:      "http",
			Domains:   w.Domains,
		},
	}
}

func (w *Webdav) Applier() (world.Applier, error) {
	return nil, nil
}

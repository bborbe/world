// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Portfolio struct {
	Context              k8s.Context
	Domains              k8s.IngressHosts
	OverlayServerVersion docker.Tag
	GitSyncPassword      deployer.SecretValue
}

func (t *Portfolio) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Context,
		t.Domains,
		t.OverlayServerVersion,
		t.GitSyncPassword,
	)
}

func (p *Portfolio) Children() []world.Configuration {
	overlayServerImage := docker.Image{
		Repository: "bborbe/portfolio",
		Tag:        p.OverlayServerVersion,
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: p.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "portfolio",
					Name:      "portfolio",
				},
			},
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.SecretApplier{
				Context:   p.Context,
				Namespace: "portfolio",
				Name:      "portfolio",
				Secrets: deployer.Secrets{
					"git-sync-password": p.GitSyncPassword,
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   p.Context,
			Namespace: "portfolio",
			Name:      "portfolio",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "portfolio",
					Image: overlayServerImage,
					Requirement: &build.OverlayWebserver{
						Image: overlayServerImage,
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
					Args: []k8s.Arg{"-logtostderr", "-v=1"},
					Ports: []deployer.Port{
						port,
					},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: port.Port.String(),
						},
						{
							Name:  "ROOT",
							Value: "/portfolio/files",
						},
						{
							Name:  "OVERLAYS",
							Value: "/overlay",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     "portfolio",
							Path:     "/portfolio",
							ReadOnly: true,
						},
						{
							Name:     "overlay",
							Path:     "/overlay",
							ReadOnly: true,
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
					},
				},
				&container.GitSync{
					MountName:  "portfolio",
					GitRepoUrl: "https://github.com/bborbe/portfolio.git",
				},
				&container.GitSync{
					MountName:                 "overlay",
					GitRepoUrl:                "https://bborbereadonly@bitbucket.org/bborbe/benjaminborbe_portfolio.git",
					GitSyncUsername:           "bborbereadonly",
					GitSyncPasswordSecretPath: "git-sync-password",
					GitSyncPasswordSecretName: "portfolio",
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "portfolio",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
				{
					Name:     "overlay",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Context,
			Namespace: "portfolio",
			Name:      "portfolio",
			Ports: []deployer.Port{
				port,
			},
		},
		&deployer.IngressDeployer{
			Context:   p.Context,
			Namespace: "portfolio",
			Name:      "portfolio",
			Port:      "http",
			Domains:   p.Domains,
		},
	}
}

func (p *Portfolio) Applier() (world.Applier, error) {
	return nil, nil
}

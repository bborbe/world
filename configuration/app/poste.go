// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Poste struct {
	Context      k8s.Context
	PosteVersion docker.Tag
	Domains      k8s.IngressHosts
	Requirements []world.Configuration
}

func (p *Poste) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.Context,
		p.PosteVersion,
		p.Domains,
	)
}

func (p *Poste) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, p.Requirements...)
	result = append(result, p.poste()...)
	return result
}

func (p *Poste) poste() []world.Configuration {
	var buildVersion docker.GitBranch = "2.0.1"
	image := docker.Image{
		Repository: "bborbe/poste.io",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", p.PosteVersion, buildVersion)),
	}
	ports := []deployer.Port{
		{
			Name:     "smtp",
			Port:     25,
			HostPort: 25,
			Protocol: "TCP",
		},
		{
			Name:     "http",
			Port:     80,
			Protocol: "TCP",
		},
		{
			Name:     "pop3",
			Port:     110,
			Protocol: "TCP",
		},
		{
			Name:     "imap",
			Port:     143,
			Protocol: "TCP",
		},
		{
			Name:     "https",
			Port:     443,
			Protocol: "TCP",
		},
		{
			Name:     "smtptls",
			Port:     465,
			HostPort: 465,
			Protocol: "TCP",
		},
		{
			Name:     "smtps",
			Port:     587,
			HostPort: 587,
			Protocol: "TCP",
		},
		{
			Name:     "imaps",
			Port:     993,
			HostPort: 993,
			Protocol: "TCP",
		},
		{
			Name:     "pop3s",
			Port:     995,
			Protocol: "TCP",
		},
		{
			Name:     "sieve",
			Port:     4190,
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: p.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "poste",
					Name:      "poste",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   p.Context,
			Namespace: "poste",
			Name:      "poste",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name: "poste",
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "1500m",
							Memory: "1000Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "500Mi",
						},
					},
					Image: image,
					Requirement: &build.Poste{
						Image:         image,
						GitBranch:     buildVersion,
						VendorVersion: p.PosteVersion,
					},
					Env: []k8s.Env{
						{
							Name:  "HTTPS",
							Value: "OFF",
						},
					},
					Ports: ports,
					Mounts: []k8s.ContainerMount{
						{
							Name: "poste",
							Path: "/data",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "poste",
					Host: k8s.PodVolumeHost{
						Path: "/data/poste",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Context,
			Namespace: "poste",
			Name:      "poste",
			Ports:     ports,
		},
		k8s.BuildIngressConfigurationWithCertManager(
			p.Context,
			"poste",
			"poste",
			"poste",
			"http",
			"/",
			p.Domains...,
		),
	}
}

func (p *Poste) Applier() (world.Applier, error) {
	return nil, nil
}

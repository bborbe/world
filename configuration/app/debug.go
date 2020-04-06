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

type Debug struct {
	Context      k8s.Context
	Domain       k8s.IngressHost
	Requirements []world.Configuration
}

func (k *Debug) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Context,
		k.Domain,
	)
}

func (k *Debug) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *Debug) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, k.Requirements...)
	result = append(result, k.debugApp()...)
	return result
}

func (k *Debug) debugApp() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/debug-server",
		Tag:        "1.0.0",
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: k.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "debug",
					Name:      "debug",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   k.Context,
			Namespace: "debug",
			Name:      "debug",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name: "server",
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: port.Port.String(),
						},
					},
					Image: image,
					Requirement: &build.DebugServer{
						Image: image,
					},
					Ports: []deployer.Port{port},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "25Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
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
			},
		},
		&deployer.ServiceDeployer{
			Context:   k.Context,
			Namespace: "debug",
			Name:      "debug",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   k.Context,
			Namespace: "debug",
			Name:      "debug",
			Port:      port.Name,
			Domains:   k8s.IngressHosts{k.Domain},
		},
		&k8s.IngresseConfiguration{
			Context: k.Context,
			Ingress: k8s.Ingress{
				ApiVersion: "extensions/v1beta1",
				Kind:       "Ingress",
				Metadata: k8s.Metadata{
					Namespace: "debug",
					Name:      "debug",
					Annotations: k8s.Annotations{
						"kubernetes.io/ingress.class":                    "nginx",
						"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
						"cert-manager.io/cluster-issuer":                 "letsencrypt-http-live",
					},
				},
				Spec: k8s.IngressSpec{
					TLS: []k8s.IngressTLS{
						{
							Hosts: []string{
								k.Domain.String(),
							},
							SecretName: k.Domain.String(),
						},
					},
					Rules: []k8s.IngressRule{
						{
							Host: k8s.IngressHost(k.Domain.String()),
							Http: k8s.IngressHttp{
								Paths: []k8s.IngressPath{
									{
										Backends: k8s.IngressBackend{
											ServiceName: "debug",
											ServicePort: "http",
										},
										Path: "/",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

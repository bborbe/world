// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"strings"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaInstalledVersionCollector struct {
	Context k8s.Context
	Apps    []struct {
		Name  string
		Regex string
		Url   string
	}
}

func (k *KafkaInstalledVersionCollector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Context,
	)
}

func (k *KafkaInstalledVersionCollector) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *KafkaInstalledVersionCollector) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/kafka-installed-version-collector",
		Tag:        "1.1.0",
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	result := []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: k.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "kafka-installed-version-collector",
					Name:      "kafka-installed-version-collector",
				},
			},
		},
	}
	for _, app := range k.Apps {
		result = append(result, &deployer.DeploymentDeployer{
			Context:   k.Context,
			Namespace: "kafka-installed-version-collector",
			Name:      k8s.MetadataName(strings.ToLower(app.Name)),
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name: "collector",
					Args: []k8s.Arg{"-v=2"},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: port.Port.String(),
						},
						{
							Name:  "KAFKA_BROKERS",
							Value: "kafka-cp-kafka-headless.kafka.svc.cluster.local:9092",
						},
						{
							Name:  "KAFKA_TOPIC",
							Value: "application-version-installed",
						},
						{
							Name:  "KAFKA_SCHEMA_REGISTRY_URL",
							Value: "http://kafka-cp-schema-registry.kafka.svc.cluster.local:8081",
						},
						{
							Name:  "APP_NAME",
							Value: app.Name,
						},
						{
							Name:  "APP_REGEX",
							Value: app.Regex,
						},
						{
							Name:  "APP_URL",
							Value: app.Url,
						},
					},
					Image: image,
					Requirement: &build.KafkaInstalledVersionCollector{
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
		})
	}
	return result
}

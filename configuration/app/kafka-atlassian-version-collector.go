package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaAtlassianVersionCollector struct {
	Context k8s.Context
}

func (k *KafkaAtlassianVersionCollector) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Context,
	)
}

func (k *KafkaAtlassianVersionCollector) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *KafkaAtlassianVersionCollector) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/kafka-atlassian-version-collector",
		Tag:        "2.0.0",
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
					Namespace: "kafka-atlassian-version-collector",
					Name:      "kafka-atlassian-version-collector",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   k.Context,
			Namespace: "kafka-atlassian-version-collector",
			Name:      "kafka-atlassian-version-collector",
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
							Value: "application-version-available",
						},
						{
							Name:  "KAFKA_SCHEMA_REGISTRY_URL",
							Value: "http://kafka-cp-schema-registry.kafka.svc.cluster.local:8081",
						},
					},
					Image: image,
					Requirement: &build.KafkaAtlassianVersionCollector{
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
	}
}
package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type KafkaLatestVersions struct {
	Cluster      cluster.Cluster
	Domain       k8s.IngressHost
	Requirements []world.Configuration
}

func (k *KafkaLatestVersions) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Cluster,
		k.Domain,
	)
}

func (k *KafkaLatestVersions) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *KafkaLatestVersions) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, k.Requirements...)
	result = append(result, k.app()...)
	return result
}

func (k *KafkaLatestVersions) app() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/kafka-latest-versions",
		Tag:        "1.1.0",
	}
	port := deployer.Port{
		Port:     80,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kafka-latest-versions",
		},
		&deployer.DeploymentDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kafka-latest-versions",
			Name:      "kafka-latest-versions",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "server",
					Image: image,
					Requirement: &build.KafkaLatestVersions{
						Image: image,
					},
					Ports: []deployer.Port{port},
					Args:  []k8s.Arg{"-v=2"},
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
							Value: "versions",
						},
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "25m",
							Memory: "25Mi",
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
			Context:   k.Cluster.Context,
			Namespace: "kafka-latest-versions",
			Name:      "kafka-latest-versions",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kafka-latest-versions",
			Name:      "kafka-latest-versions",
			Port:      "http",
			Domains:   k8s.IngressHosts{k.Domain},
		},
	}
}

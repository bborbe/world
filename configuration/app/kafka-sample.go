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

type KafkaSample struct {
	Cluster      cluster.Cluster
	Domain       k8s.IngressHost
	Requirements []world.Configuration
}

func (k *KafkaSample) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.Cluster,
		k.Domain,
	)
}

func (k *KafkaSample) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *KafkaSample) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, k.Requirements...)
	result = append(result, k.sampleApp()...)
	return result
}

func (k *KafkaSample) sampleApp() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/kafka-sample",
		Tag:        "1.1.1",
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kafka-sample",
		},
		&deployer.DeploymentDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kafka-sample",
			Name:      "kafka-sample",
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
							Name:  "BROKERS",
							Value: "kafka-0.kafka.kafka.svc.cluster.local:9093",
						},
					},
					Image: image,
					Requirement: &build.KafkaSample{
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
			Context:   k.Cluster.Context,
			Namespace: "kafka-sample",
			Name:      "kafka-sample",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kafka-sample",
			Name:      "kafka-sample",
			Port:      port.Name,
			Domains:   k8s.IngressHosts{k.Domain},
		},
	}
}

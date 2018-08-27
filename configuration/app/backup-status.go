package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type BackupStatus struct {
	Cluster cluster.Cluster
	Domains k8s.IngressHosts
}

func (t *BackupStatus) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
	)
}

func (b *BackupStatus) Children() []world.Configuration {
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	image := docker.Image{
		Repository: "bborbe/backup-status-client",
		Tag:        "2.0.0",
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
		},
		&deployer.DeploymentDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "status",
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
					Requirement: &build.BackupStatusClient{
						Image: image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args:  []k8s.Arg{"-logtostderr", "-v=1"},
					Ports: []deployer.Port{port},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: "8080",
						},
						{
							Name:  "SERVER",
							Value: "http://backup.pn.benjamin-borbe.de:1080",
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
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "status",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "status",
			Port:      "http",
			Domains:   b.Domains,
		},
	}
}

func (b *BackupStatus) Applier() (world.Applier, error) {
	return nil, nil
}

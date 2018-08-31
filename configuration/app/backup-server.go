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

type BackupServer struct {
	Cluster cluster.Cluster
}

func (t *BackupServer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
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
		&deployer.NamespaceDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
		},
		&deployer.DeploymentDeployer{
			Context:   b.Cluster.Context,
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
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data",
						Server: b.Cluster.NfsServer,
					},
				},
				{
					Name: "ssh",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/backup-ssh",
						Server: b.Cluster.NfsServer,
					},
				},
			},
		},
	}
}

func (b *BackupServer) Applier() (world.Applier, error) {
	return nil, nil
}

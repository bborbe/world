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

type Backup struct {
	Cluster cluster.Cluster
	Domains k8s.IngressHosts
}

func (t *Backup) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
	)
}

func (b *Backup) Children() []world.Configuration {
	result := []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
		},
	}
	result = append(result, b.rsync()...)
	result = append(result, b.status()...)
	return result
}

func (b *Backup) rsync() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/backup-rsync-server",
		Tag:        "1.1.0",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "rsync",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "backup",
					Image: image,
					Requirement: &build.BackupRsyncServer{
						Image: image,
					},
					Ports: []deployer.Port{
						{
							Port:     22,
							HostPort: 2222,
							Name:     "ssh",
							Protocol: "TCP",
						},
					},
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "1000m",
							Memory: "200Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "250m",
							Memory: "100Mi",
						},
					},
					Mounts: []k8s.VolumeMount{
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

func (b *Backup) status() []world.Configuration {
	ports := []deployer.Port{
		{
			Port:     8080,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	image := docker.Image{
		Repository: "bborbe/backup-status-client",
		Tag:        "2.0.0",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "status",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "backup",
					Image: image,
					Requirement: &build.BackupStatusClient{
						Image: image,
					},
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "100m",
							Memory: "50Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args:  []k8s.Arg{"-logtostderr", "-v=1"},
					Ports: ports,
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
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "status",
			Ports:     ports,
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

func (b *Backup) Applier() (world.Applier, error) {
	return nil, nil
}

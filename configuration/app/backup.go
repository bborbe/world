package app

import (
	"context"

	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Backup struct {
	Cluster cluster.Cluster
	Domains []k8s.IngressHost
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
		Registry:   "docker.io",
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
					CpuLimit:      "1000m",
					MemoryLimit:   "200Mi",
					CpuRequest:    "250m",
					MemoryRequest: "100Mi",
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
					Nfs: k8s.PodNfs{
						Path:   "/data",
						Server: b.Cluster.NfsServer,
					},
				},
				{
					Name: "ssh",
					Nfs: k8s.PodNfs{
						Path:   "/data/backup-ssh",
						Server: b.Cluster.NfsServer,
					},
				},
			},
		},
	}
}

func (b *Backup) status() []world.Configuration {
	var vendorVersion docker.Tag = "2.0.0"
	var buildVersion docker.GitBranch = "1.0.1"
	ports := []deployer.Port{
		{
			Port:     8080,
			Name:     "http",
			Protocol: "TCP",
		},
	}
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/backup-status-client",
		Tag:        docker.Tag(fmt.Sprintf("%s-%s", vendorVersion, buildVersion)),
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
						VendorVersion: vendorVersion,
						GitBranch:     buildVersion,
						Image:         image,
					},
					CpuLimit:      "100m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Args:          []k8s.Arg{"-logtostderr", "-v=1"},
					Ports:         ports,
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

func (b *Backup) Applier() world.Applier {
	return nil
}

func (b *Backup) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate backup app ...")
	if err := b.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate cluster failed")
	}
	if len(b.Domains) == 0 {
		return errors.New("domains empty")
	}
	return nil
}

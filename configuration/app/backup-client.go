package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type BackupClient struct {
	Cluster         cluster.Cluster
	Domains         k8s.IngressHosts
	GitSyncPassword deployer.SecretValue
	BackupSshKey    deployer.SecretValue
	GitRepoUrl      container.GitRepoUrl
}

func (t *BackupClient) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
		t.GitSyncPassword,
		t.BackupSshKey,
		t.GitRepoUrl,
	)
}

func (b *BackupClient) Children() []world.Configuration {
	result := []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
		},
	}
	result = append(result, b.rsync()...)
	result = append(result, b.cleanup()...)
	result = append(result, b.statusServer()...)
	result = append(result, b.statusClient()...)
	return result
}

func (b *BackupClient) rsync() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/backup-rsync-client",
		Tag:        "1.2.1",
	}
	return []world.Configuration{
		&deployer.SecretDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "backup",
			Secrets: deployer.Secrets{
				"git-sync-password": b.GitSyncPassword,
				"backup-ssh-key":    b.BackupSshKey,
			},
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
					Name:  "rsync",
					Image: image,
					Requirement: &build.BackupStatusServer{
						Image: image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "2000m",
							Memory: "1000Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "500Mi",
						},
					},
					Args: []k8s.Arg{"-logtostderr", "-v=4"},
					Env: []k8s.Env{
						{
							Name:  "CONFIG",
							Value: "/config/backup-config.json",
						},
						{
							Name:  "TARGET",
							Value: "/backup",
						},
						{
							Name:  "DELAY",
							Value: "5m",
						},
						{
							Name:  "ONE_TIME",
							Value: "false",
						},
						{
							Name:  "LOCK",
							Value: "/backup-rsync-client.lock",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     "config",
							Path:     "/config",
							ReadOnly: true,
						},
						{
							Name:     "ssh",
							Path:     "/root/.ssh",
							ReadOnly: true,
						},
						{
							Name: "backup",
							Path: "/backup",
						},
					},
				},
				&container.GitSync{
					MountName:                 "config",
					GitRepoUrl:                b.GitRepoUrl,
					GitSyncUsername:           "bborbereadonly",
					GitSyncPasswordSecretName: "backup",
					GitSyncPasswordSecretPath: "git-sync-password",
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "config",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
				{
					Name: "backup",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/backup",
						Server: b.Cluster.NfsServer,
					},
				},
				{
					Name: "ssh",
					Secret: k8s.PodVolumeSecret{
						Name: "backup",
						Items: []k8s.PodSecretItem{
							{
								Key:  "backup-ssh-key",
								Mode: 384,
								Path: "id_rsa",
							},
						},
					},
				},
			},
		},
	}
}

func (b *BackupClient) cleanup() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/backup-rsync-cleanup",
		Tag:        "1.3.1",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "cleanup",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "cleanup",
					Image: image,
					Requirement: &build.BackupStatusServer{
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
					Args: []k8s.Arg{"-logtostderr", "-v=1"},
					Env: []k8s.Env{
						{
							Name:  "TARGET",
							Value: "/backup",
						},
						{
							Name:  "DELAY",
							Value: "5m",
						},
						{
							Name:  "ONE_TIME",
							Value: "false",
						},
						{
							Name:  "LOCK",
							Value: "/backup-rsync-cleanup.lock",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "backup",
							Path: "/backup",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "backup",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/backup",
						Server: b.Cluster.NfsServer,
					},
				},
			},
		},
	}
}

func (b *BackupClient) statusServer() []world.Configuration {
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	image := docker.Image{
		Repository: "bborbe/backup-status-server",
		Tag:        "1.2.3",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "status-server",
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
					Requirement: &build.BackupStatusServer{
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
							Value: port.Port.String(),
						},
						{
							Name:  "ROOTDIR",
							Value: "/backup",
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
					Mounts: []k8s.ContainerMount{
						{
							Name:     "backup",
							Path:     "/backup",
							ReadOnly: true,
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "backup",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/backup",
						Server: b.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "status-server",
			Ports:     []deployer.Port{port},
		},
	}
}

func (b *BackupClient) statusClient() []world.Configuration {
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
		&deployer.DeploymentDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "status-client",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "client",
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
							Value: port.Port.String(),
						},
						{
							Name:  "SERVER",
							Value: "http://status-server:8080",
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
			Name:      "status-client",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
			Name:      "status-client",
			Port:      "http",
			Domains:   b.Domains,
		},
	}
}

func (b *BackupClient) Applier() (world.Applier, error) {
	return nil, nil
}

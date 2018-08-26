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
	"github.com/pkg/errors"
)

type MonitoringConfig struct {
	Name       k8s.MetadataName
	Subject    string
	GitRepoUrl string
}

func (m *MonitoringConfig) Validate(ctx context.Context) error {
	if m.Subject == "" {
		return errors.New("Subject missing")
	}
	if m.GitRepoUrl == "" {
		return errors.New("GitRepoUrl missing")
	}
	return validation.Validate(
		ctx,
		m.Name,
	)
}

type Monitoring struct {
	Cluster         cluster.Cluster
	SmtpPassword    deployer.SecretValue
	GitSyncPassword deployer.SecretValue
	GitSyncVersion  docker.Tag
	Configs         []MonitoringConfig
}

func (w *Monitoring) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Cluster,
		w.SmtpPassword,
		w.GitSyncPassword,
		w.GitSyncVersion,
	)
}

func (m *Monitoring) Children() []world.Configuration {
	configurations := []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "monitoring",
		},
		&deployer.SecretDeployer{
			Context:   m.Cluster.Context,
			Namespace: "monitoring",
			Name:      "monitoring",
			Secrets: deployer.Secrets{
				"git-sync-password": m.GitSyncPassword,
				"smtp-password":     m.SmtpPassword,
			},
		},
	}
	for _, config := range m.Configs {
		configurations = append(configurations, &MonitoringDeployment{
			Context:        m.Cluster.Context,
			Name:           config.Name,
			Subject:        config.Subject,
			GitRepoUrl:     config.GitRepoUrl,
			GitSyncVersion: m.GitSyncVersion,
		})
	}
	return configurations
}

func (m *Monitoring) Applier() (world.Applier, error) {
	return nil, nil
}

type MonitoringDeployment struct {
	Context        k8s.Context
	Name           k8s.MetadataName
	Subject        string
	GitRepoUrl     string
	GitSyncVersion docker.Tag
}

func (m *MonitoringDeployment) Validate(ctx context.Context) error {
	if m.Subject == "" {
		return errors.New("Subject missing")
	}
	if m.GitRepoUrl == "" {
		return errors.New("GitRepoUrl missing")
	}
	return validation.Validate(
		ctx,
		m.GitSyncVersion,
		m.Context,
		m.Name,
	)
}

func (m *MonitoringDeployment) Children() []world.Configuration {
	gitSyncImage := docker.Image{
		Repository: "bborbe/git-sync",
		Tag:        m.GitSyncVersion,
	}
	monitoringImage := docker.Image{
		Repository: "bborbe/monitoring",
		Tag:        "1.2.0",
	}
	mountName := k8s.MountName("data")
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   m.Context,
			Namespace: "monitoring",
			Name:      m.Name,
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "monitoring",
					Image: monitoringImage,
					Requirement: &build.Monitoring{
						Image: monitoringImage,
					},
					Args: []k8s.Arg{
						"-logtostderr",
						"-v=1",
					},
					Env: []k8s.Env{
						{
							Name:  "CONFIG",
							Value: "/data/config.xml",
						},
						{
							Name:  "DRIVER",
							Value: "phantomjs",
						},
						{
							Name:  "DELAY",
							Value: "5m",
						},
						{
							Name:  "CONCURRENT",
							Value: "1",
						},
						{
							Name:  "ONE_TIME",
							Value: "false",
						},
						{
							Name:  "LOCK",
							Value: "/monitoring.lock",
						},
						{
							Name:  "SMTP_TLS",
							Value: "true",
						},
						{
							Name:  "SMTP_TLS_SKIP_VERIFY",
							Value: "true",
						},
						{
							Name:  "SMTP_USER",
							Value: "monitoring@benjamin-borbe.de",
						},
						{
							Name: "SMTP_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "smtp-password",
									Name: "monitoring",
								},
							},
						},
						{
							Name:  "SMTP_HOST",
							Value: "mail.benjamin-borbe.de",
						},
						{
							Name:  "SMTP_PORT",
							Value: "465",
						},
						{
							Name:  "SENDER",
							Value: "monitoring@benjamin-borbe.de",
						},
						{
							Name:  "RECIPIENT",
							Value: "bborbe@rocketnews.de",
						},
						{
							Name:  "SUBJECT",
							Value: m.Subject,
						},
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     mountName,
							Path:     "/data",
							ReadOnly: true,
						},
					},
				},
				{
					Name:  "git-sync",
					Image: gitSyncImage,
					Requirement: &build.GitSync{
						Image: gitSyncImage,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "50m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args: []k8s.Arg{
						"-logtostderr",
						"-v=2",
					},
					Env: []k8s.Env{
						{
							Name:  "GIT_SYNC_REPO",
							Value: m.GitRepoUrl,
						},
						{
							Name:  "GIT_SYNC_DEST",
							Value: "/data",
						},
						{
							Name:  "GIT_SYNC_USERNAME",
							Value: "bborbereadonly",
						},
						{
							Name: "GIT_SYNC_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "git-sync-password",
									Name: "monitoring",
								},
							},
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: mountName,
							Path: "/data",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     mountName,
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
			},
		},
	}
}

func (m *MonitoringDeployment) Applier() (world.Applier, error) {
	return nil, nil
}

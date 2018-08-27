package app

import (
	"context"

	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Prometheus struct {
	Cluster            cluster.Cluster
	PrometheusDomain   k8s.IngressHost
	AlertmanagerDomain k8s.IngressHost
	Secret             deployer.SecretValue
	LdapUsername       deployer.SecretValue
	LdapPassword       deployer.SecretValue
}

func (t *Prometheus) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Cluster,
		t.PrometheusDomain,
		t.AlertmanagerDomain,
		t.Secret,
		t.LdapUsername,
		t.LdapPassword,
	)
}

func (p *Prometheus) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, p.nodeExporter()...)
	result = append(result, p.prometheus()...)
	result = append(result, p.alertmanager()...)
	return result
}

func (p *Prometheus) prometheus() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/prometheus",
		Tag:        "v1.8.2",
	}
	prometheusPort := deployer.Port{
		Port:     9090,
		Name:     "http",
		Protocol: "TCP",
	}
	authPort := deployer.Port{
		Port:     9091,
		Name:     "http-auth",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "prometheus",
					Image: image,
					Args: []k8s.Arg{
						"-config.file=/config/prometheus.yaml",
						"-storage.local.retention=48h",
						"-storage.local.path=/prometheus",
						"-storage.local.target-heap-size=500000000",
						"-web.console.libraries=/etc/prometheus/console_libraries",
						"-web.console.templates=/etc/prometheus/consoles",
						k8s.Arg(fmt.Sprintf("-web.external-url=https://%s", p.PrometheusDomain)),
						"-alertmanager.url=http://prometheus-alertmanager:9093",
						"-log.level=info",
					},
					Requirement: &build.Prometheus{
						Image: image,
					},
					Ports: []deployer.Port{
						prometheusPort,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "1000m",
							Memory: "800Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "400Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "config",
							Path: "/config",
						},
						{
							Name: "prometheus",
							Path: "/prometheus",
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   prometheusPort.Port,
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
							Port:   prometheusPort.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
					},
				},
				&container.Auth{
					Context:      p.Cluster.Context,
					Namespace:    "kube-system",
					Port:         authPort.Port,
					TargetPort:   prometheusPort.Port,
					Secret:       p.Secret,
					LdapUsername: p.LdapUsername,
					LdapPassword: p.LdapPassword,
				},
				&container.GitSync{
					MountName:  "config",
					GitRepoUrl: "https://github.com/bborbe/prometheus-nc.git",
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "config",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
				{
					Name: "prometheus",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/prometheus",
						Server: k8s.PodNfsServer(p.Cluster.NfsServer),
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus",
			Ports: []deployer.Port{
				prometheusPort,
				authPort,
			},
			Annotations: k8s.Annotations{
				"prometheus.io/path":   "/metrics",
				"prometheus.io/port":   "9090",
				"prometheus.io/scheme": "http",
				"prometheus.io/scrape": "true",
			},
		},
		&deployer.IngressDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus",
			Port:      "http-auth",
			Domains:   k8s.IngressHosts{p.PrometheusDomain},
		},
	}
}

func (p *Prometheus) alertmanager() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/prometheus-alertmanager",
		Tag:        "v0.14.0",
	}
	alertmanagerPort := deployer.Port{
		Port:     9093,
		Name:     "http",
		Protocol: "TCP",
	}
	authPort := deployer.Port{
		Port:     9094,
		Name:     "http-auth",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus-alertmanager",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "alertmanager",
					Image: image,
					Args: []k8s.Arg{
						"--config.file=/config/alertmanager.yaml",
						k8s.Arg(fmt.Sprintf("--web.external-url=https://%s", p.AlertmanagerDomain)),
					},
					Requirement: &build.PrometheusAlertmanager{
						Image: image,
					},
					Ports: []deployer.Port{
						alertmanagerPort,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "50Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "config",
							Path: "/config",
						},
						{
							Name: "alertmanager",
							Path: "/alertmanager",
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   alertmanagerPort.Port,
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
							Port:   alertmanagerPort.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
					},
				},
				&container.Auth{
					Context:      p.Cluster.Context,
					Namespace:    "kube-system",
					Port:         authPort.Port,
					TargetPort:   alertmanagerPort.Port,
					Secret:       p.Secret,
					LdapUsername: p.LdapUsername,
					LdapPassword: p.LdapPassword,
				},
				&container.GitSync{
					MountName:  "config",
					GitRepoUrl: "https://github.com/bborbe/prometheus-nc.git",
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "alertmanager",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
				{
					Name:     "config",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus-alertmanager",
			Ports: []deployer.Port{
				alertmanagerPort,
				authPort,
			},
		},
		&deployer.IngressDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus-alertmanager",
			Port:      "http-auth",
			Domains:   k8s.IngressHosts{p.AlertmanagerDomain},
		},
	}
}

func (p *Prometheus) nodeExporter() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/prometheus-node-exporter",
		Tag:        "v0.14.0",
	}
	return []world.Configuration{
		&deployer.DaemonSetDeployer{
			Requirements: []world.Configuration{
				&build.PrometheusNodeExporter{
					Image: image,
				},
			},
			Context: p.Cluster.Context,
			DaemonSet: k8s.DaemonSet{
				ApiVersion: "apps/v1",
				Kind:       "DaemonSet",
				Metadata: k8s.Metadata{
					Name:      "prometheus-node-exporter",
					Namespace: "kube-system",
					Labels: k8s.Labels{
						"app": "prometheus-node-exporter",
					},
				},
				Spec: k8s.DaemonSetSpec{
					Selector: k8s.Selector{
						MatchLabels: k8s.Labels{
							"app": "prometheus-node-exporter",
						},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Annotations: k8s.Annotations{
								"prometheus.io/path":   "/metrics",
								"prometheus.io/port":   "9100",
								"prometheus.io/scheme": "http",
								"prometheus.io/scrape": "true",
							},
							Labels: k8s.Labels{
								"app": "prometheus-node-exporter",
							},
						},
						Spec: k8s.PodSpec{
							HostPid: true,
							Containers: []k8s.Container{
								{
									Args: []k8s.Arg{
										"-collector.procfs=/host/proc",
										"-collector.sysfs=/host/sys",
										"-collector.filesystem.ignored-mount-points='^/(sys|proc|dev|host|etc)($|/)'",
									},
									Image: k8s.Image(image.String()),
									Name:  "node-exporter",
									Ports: []k8s.ContainerPort{
										{
											Name:          "http",
											ContainerPort: 9100,
										},
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
									SecurityContext: k8s.SecurityContext{
										Privileged: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (p *Prometheus) Applier() (world.Applier, error) {
	return nil, nil
}

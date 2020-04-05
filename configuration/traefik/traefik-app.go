// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package traefik

import (
	"context"

	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type App struct {
	Context      k8s.Context
	Domains      k8s.IngressHosts
	SSL          bool
	DisableRBAC  bool
	Requirements []world.Configuration
}

func (a *App) Validate(ctx context.Context) error {
	if a.SSL {
		return validation.Validate(
			ctx,
			a.Context,
			a.Domains,
		)
	}
	return validation.Validate(
		ctx,
		a.Context,
		a.Domains,
	)
}

func (a *App) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, a.Requirements...)
	result = append(result, a.traefik()...)
	return result
}

func (a *App) traefik() []world.Configuration {
	traefikImage := docker.Image{
		Repository: "bborbe/traefik",
		Tag:        "1.7.24-alpine", // https://hub.docker.com/_/traefik?tab=tags
	}
	httpPort := deployer.Port{
		Port:     80,
		HostPort: 80,
		Name:     "http",
		Protocol: "TCP",
	}
	httpsPort := deployer.Port{
		Port:     443,
		HostPort: 443,
		Name:     "https",
		Protocol: "TCP",
	}
	dashboardPort := deployer.Port{
		Port:     8080,
		Name:     "dashboard",
		Protocol: "TCP",
	}
	ports := deployer.Ports{
		httpPort,
		dashboardPort,
	}
	if a.SSL {
		ports = append(ports, httpsPort)
	}
	exporterImage := docker.Image{
		Repository: "bborbe/traefik-certificate-extractor",
		Tag:        "v1.2.2",
	}
	var acmeVolume k8s.PodVolume
	if a.SSL {
		acmeVolume = k8s.PodVolume{
			Name: "acme",
			Host: k8s.PodVolumeHost{
				Path: "/data/traefik-acme",
			},
		}
	} else {
		acmeVolume = k8s.PodVolume{
			Name:     "acme",
			EmptyDir: &k8s.PodVolumeEmptyDir{},
		}
	}
	result := []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: a.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "traefik",
					Name:      "traefik",
				},
			},
		},
		&k8s.ServiceaccountConfiguration{
			Context: a.Context,
			Serviceaccount: k8s.ServiceAccount{
				ApiVersion: "v1",
				Kind:       "ServiceAccount",
				Metadata: k8s.Metadata{
					Namespace: "traefik",
					Name:      "traefik",
				},
			},
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   a.Context,
				Namespace: "traefik",
				Name:      "traefik",
				ConfigValues: map[string]deployer.ConfigValue{
					"config": deployer.ConfigValueFunc(func(ctx context.Context) (string, error) {
						if a.SSL {
							return traefikConfigWithHttps, nil
						}
						return traefikConfigWithoutHttps, nil
					}),
				},
			},
		),
		&BuildApp{
			Image: traefikImage,
		},
		&k8s.DeploymentConfiguration{
			Context: a.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "traefik",
					Name:      "traefik",
				},
				Spec: k8s.DeploymentSpec{
					Replicas: 1,
					Strategy: k8s.DeploymentStrategy{
						Type: "RollingUpdate",
						RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
							MaxUnavailable: 1,
						},
					},
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"app": "traefik",
						},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "traefik",
							},
						},
						Spec: k8s.PodSpec{
							ServiceAccountName:            "traefik",
							TerminationGracePeriodSeconds: 60,
							Containers: []k8s.Container{
								{
									Name:  "traefik",
									Image: k8s.Image(traefikImage.String()),
									Ports: ports.ContainerPort(),
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "200m",
											Memory: "100Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "100m",
											Memory: "25Mi",
										},
									},
									Args: []k8s.Arg{
										"--configfile=/config/traefik.toml",
										"--logLevel=INFO", // "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC"
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Name: "config",
											Path: "/config",
										},
										{
											Name: "acme",
											Path: "/acme",
										},
									},
									LivenessProbe: k8s.Probe{
										TcpSocket: k8s.TcpSocket{
											Port: httpPort.Port,
										},
										FailureThreshold:    3,
										InitialDelaySeconds: 10,
										PeriodSeconds:       10,
										SuccessThreshold:    1,
										TimeoutSeconds:      2,
									},
									ReadinessProbe: k8s.Probe{
										TcpSocket: k8s.TcpSocket{
											Port: httpPort.Port,
										},
										FailureThreshold:    1,
										InitialDelaySeconds: 10,
										PeriodSeconds:       10,
										SuccessThreshold:    1,
										TimeoutSeconds:      2,
									},
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "config",
									ConfigMap: k8s.PodVolumeConfigMap{
										Name: "traefik",
										Items: []k8s.PodConfigMapItem{
											{
												Key:  "config",
												Path: "traefik.toml",
											},
										},
									},
								},
								acmeVolume,
							},
						},
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   a.Context,
			Namespace: "traefik",
			Name:      "traefik",
			Ports:     ports,
			Annotations: k8s.Annotations{
				"prometheus.io/path":   "/metrics",
				"prometheus.io/port":   "8080",
				"prometheus.io/scheme": "http",
				"prometheus.io/scrape": "true",
			},
		},
		&deployer.IngressDeployer{
			Context:   a.Context,
			Namespace: "traefik",
			Name:      "traefik",
			Port:      "dashboard",
			Domains:   a.Domains,
		},
	}
	if a.SSL {
		result = append(result,
			&deployer.DeploymentDeployer{
				Context:   a.Context,
				Namespace: "traefik",
				Name:      "traefik-extract",
				Strategy: k8s.DeploymentStrategy{
					Type: "RollingUpdate",
					RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
						MaxSurge:       1,
						MaxUnavailable: 1,
					},
				},
				Containers: []deployer.HasContainer{
					&deployer.DeploymentDeployerContainer{
						Name:  "traefik-extract",
						Image: exporterImage,
						Requirement: &BuildCertificateExtractor{
							Image: exporterImage,
						},
						Resources: k8s.Resources{
							Limits: k8s.ContainerResource{
								Cpu:    "200m",
								Memory: "100Mi",
							},
							Requests: k8s.ContainerResource{
								Cpu:    "100m",
								Memory: "25Mi",
							},
						},
						Mounts: []k8s.ContainerMount{
							{
								Name:     "acme",
								Path:     "/app/data",
								ReadOnly: true,
							},
							{
								Name: "certs",
								Path: "/app/certs",
							},
						},
					},
				},
				Volumes: []k8s.PodVolume{
					{
						Name: "acme",
						Host: k8s.PodVolumeHost{
							Path: "/data/traefik-acme",
						},
					},
					{
						Name: "certs",
						Host: k8s.PodVolumeHost{
							Path: "/data/traefik-extract",
						},
					},
				},
			})
	}

	if !a.DisableRBAC {
		result = append(result,
			&k8s.ClusterRoleConfiguration{
				Context: a.Context,
				ClusterRole: k8s.ClusterRole{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Kind:       "ClusterRole",
					Metadata: k8s.Metadata{
						Namespace: "",
						Name:      "traefik",
					},
					Rules: []k8s.PolicyRule{
						{
							ApiGroups: []string{
								"",
							},
							Resources: []string{
								"pods",
								"services",
								"endpoints",
								"secrets",
							},
							Verbs: []string{
								"get",
								"list",
								"watch",
							},
						},
						{
							ApiGroups: []string{
								"extensions",
							},
							Resources: []string{
								"ingresses",
							},
							Verbs: []string{
								"get",
								"list",
								"watch",
							},
						},
						{
							ApiGroups: []string{
								"extensions",
							},
							Resources: []string{
								"ingresses/status",
							},
							Verbs: []string{
								"update",
							},
						},
					},
				},
			},
			&k8s.ClusterRoleBindingConfiguration{
				Context: a.Context,
				ClusterRoleBinding: k8s.ClusterRoleBinding{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Kind:       "ClusterRoleBinding",
					Metadata: k8s.Metadata{
						Name: "traefik",
					},
					Subjects: []k8s.Subject{
						{
							Kind:      "ServiceAccount",
							Name:      "traefik",
							Namespace: "traefik",
						},
					},
					RoleRef: k8s.RoleRef{
						Kind:     "ClusterRole",
						Name:     "traefik",
						ApiGroup: "rbac.authorization.k8s.io",
					},
				},
			},
		)
	}

	return result
}
func (a *App) Applier() (world.Applier,
	error) {
	return nil,
		nil
}

const traefikConfigWithHttps = `
graceTimeOut = 10
debug = false
logLevel = "INFO"
defaultEntryPoints = ["http","https"]
[entryPoints]
[entryPoints.http]
address = ":80"
compress = false
[entryPoints.http.redirect]
entryPoint = "https"
[entryPoints.https]
address = ":443"
compress = false
[entryPoints.https.tls]
cipherSuites = [
"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
]
[kubernetes]
[web]
address = ":8080"
[web.metrics.prometheus]
[acme]
email = "bborbe@rocketnews.de"
storage = "/acme/acme.json"
entryPoint = "https"
onHostRule = true
acmeLogging = true
[acme.httpChallenge]
entryPoint = "http"
`

const traefikConfigWithoutHttps = `
graceTimeOut = 10
debug = false
logLevel = "INFO"
defaultEntryPoints = ["http"]
[entryPoints]
[entryPoints.http]
address = ":80"
compress = false
[kubernetes]
[web]
address = ":8080"
[web.metrics.prometheus]
email = "bborbe@rocketnews.de"
entryPoint = "http"
onHostRule = true
`

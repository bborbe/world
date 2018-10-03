package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Traefik struct {
	Context   k8s.Context
	NfsServer k8s.PodNfsServer
	Domains   k8s.IngressHosts
	SSL       bool
}

func (t *Traefik) Validate(ctx context.Context) error {
	if t.SSL {
		return validation.Validate(
			ctx,
			t.Context,
			t.Domains,
			t.NfsServer,
		)
	}
	return validation.Validate(
		ctx,
		t.Context,
		t.Domains,
	)
}

func (t *Traefik) Children() []world.Configuration {
	traefikImage := docker.Image{
		Repository: "bborbe/traefik",
		Tag:        "1.7.0-alpine",
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
	ports := []deployer.Port{
		httpPort,
		dashboardPort,
	}
	if t.SSL {
		ports = append(ports, httpsPort)
	}
	exporterImage := docker.Image{
		Repository: "bborbe/traefik-certificate-extractor",
		Tag:        "latest",
	}
	var acmeVolume k8s.PodVolume
	if t.SSL {
		acmeVolume = k8s.PodVolume{
			Name: "acme",
			Nfs: k8s.PodVolumeNfs{
				Path:   "/data/traefik-acme",
				Server: t.NfsServer,
			},
		}
	} else {
		acmeVolume = k8s.PodVolume{
			Name:     "acme",
			EmptyDir: &k8s.PodVolumeEmptyDir{},
		}
	}
	result := []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   t.Context,
			Namespace: "traefik",
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   t.Context,
				Namespace: "traefik",
				Name:      "traefik",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "config",
						Value: t.traefikConfig(),
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   t.Context,
			Namespace: "traefik",
			Name:      "traefik",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "traefik",
					Image: traefikImage,
					Requirement: &build.Traefik{
						Image: traefikImage,
					},
					Ports: ports,
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
					Args: []k8s.Arg{"--configfile=/config/traefik.toml"},
					Mounts: []k8s.ContainerMount{
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
		&deployer.ServiceDeployer{
			Context:   t.Context,
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
			Context:   t.Context,
			Namespace: "traefik",
			Name:      "traefik",
			Port:      "dashboard",
			Domains:   t.Domains,
		},
	}
	if t.SSL {
		result = append(result, &deployer.DeploymentDeployer{
			Context:   t.Context,
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
					Requirement: &build.TraefikCertificateExtractor{
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
							Name: "acme",
							Path: "/app/data",
						},
						{
							Name:     "certs",
							Path:     "/app/certs",
							ReadOnly: true,
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "acme",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/traefik-acme",
						Server: t.NfsServer,
					},
				},
				{
					Name: "certs",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/traefik-extract",
						Server: t.NfsServer,
					},
				},
			},
		})
	}
	return result
}

func (t *Traefik) Applier() (world.Applier, error) {
	return nil, nil
}

func (t *Traefik) traefikConfig() string {
	if t.SSL {
		return traefikConfigWithHttps
	}
	return traefikConfigWithoutHttps
}

const traefikConfigWithHttps = `graceTimeOut = 10
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

const traefikConfigWithoutHttps = `graceTimeOut = 10
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

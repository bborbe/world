package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Traefik struct {
	Cluster cluster.Cluster
	Domains []k8s.IngressHost
}

func (t *Traefik) Children() []world.Configuration {
	traefikImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/traefik",
		Tag:        "1.5.3-alpine",
	}
	traefikPorts := []deployer.Port{
		{
			Port:     80,
			HostPort: 80,
			Name:     "http",
			Protocol: "TCP",
		},
		{
			Port:     443,
			HostPort: 443,
			Name:     "https",
			Protocol: "TCP",
		},
		{
			Port:     8080,
			Name:     "dashboard",
			Protocol: "TCP",
		},
	}
	exporterImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/traefik-certificate-extractor",
		Tag:        "latest",
	}
	return []world.Configuration{
		&deployer.ConfigMapDeployer{
			Context:   t.Cluster.Context,
			Namespace: "kube-system",
			Name:      "traefik",
			ConfigMap: k8s.ConfigMapData{
				"config": traefikConfig,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   t.Cluster.Context,
			Namespace: "kube-system",
			Name:      "traefik",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "traefik",
					Image: traefikImage,
					Requirement: &build.Traefik{
						Image: traefikImage,
					},
					Ports:         traefikPorts,
					CpuLimit:      "200m",
					MemoryLimit:   "100Mi",
					CpuRequest:    "100m",
					MemoryRequest: "25Mi",
					Args:          []k8s.Arg{"--configfile=/config/traefik.toml"},
					Mounts: []k8s.VolumeMount{
						{
							Name: "config",
							Path: "/config",
						},
						{
							Name: "acme",
							Path: "/acme",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "config",
					ConfigMap: k8s.PodConfigMap{
						Name: "traefik",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "config",
								Path: "traefik.toml",
							},
						},
					},
				},
				{
					Name: "acme",
					Nfs: k8s.PodNfs{
						Path:   "/data/traefik-acme",
						Server: t.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   t.Cluster.Context,
			Namespace: "kube-system",
			Name:      "traefik-extract",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "traefik-extract",
					Image: exporterImage,
					Requirement: &build.TraefikCertificateExtractor{
						Image: exporterImage,
					},
					CpuLimit:      "200m",
					MemoryLimit:   "100Mi",
					CpuRequest:    "100m",
					MemoryRequest: "25Mi",
					Mounts: []k8s.VolumeMount{
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
					Nfs: k8s.PodNfs{
						Path:   "/data/traefik-acme",
						Server: t.Cluster.NfsServer,
					},
				},
				{
					Name: "certs",
					Nfs: k8s.PodNfs{
						Path:   "/data/traefik-extract",
						Server: t.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   t.Cluster.Context,
			Namespace: "kube-system",
			Name:      "traefik",
			Ports:     traefikPorts,
			Annotations: k8s.Annotations{
				"prometheus.io/path":   "/metrics",
				"prometheus.io/port":   "8080",
				"prometheus.io/scheme": "http",
				"prometheus.io/scrape": "true",
			},
		},
		&deployer.IngressDeployer{
			Context:   t.Cluster.Context,
			Namespace: "kube-system",
			Name:      "traefik",
			Port:      "dashboard",
			Domains:   t.Domains,
		},
	}
}

func (t *Traefik) Applier() world.Applier {
	return nil
}

func (t *Traefik) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate traefik app ...")
	if err := t.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate traefik app failed")
	}
	if len(t.Domains) == 0 {
		return errors.New("domains empty")
	}
	return nil
}

const traefikConfig = `graceTimeOut = 10
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

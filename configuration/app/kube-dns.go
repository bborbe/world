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

type KubeDns struct {
	Cluster cluster.Cluster
}

func (d *KubeDns) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Cluster,
	)
}

func (d *KubeDns) Children() []world.Configuration {
	kubednsImage := docker.Image{
		Repository: "bborbe/kubedns",
		Tag:        "1.9",
	}
	kubednsPorts := []deployer.Port{
		{
			Name:     "dns-local",
			Port:     10053,
			Protocol: "UDP",
		},
		{
			Name:     "dns-tcp-local",
			Port:     10053,
			Protocol: "TCP",
		},
	}
	kubednsMasqImage := docker.Image{
		Repository: "bborbe/kube-dnsmasq",
		Tag:        "1.4.1",
	}
	kubednsMasqPorts := []deployer.Port{
		{
			Name:     "dns",
			Port:     53,
			Protocol: "UDP",
		},
		{
			Name:     "dns-tcp",
			Port:     53,
			Protocol: "TCP",
		},
	}
	healthzImage := docker.Image{
		Repository: "bborbe/exechealthz",
		Tag:        "v1.3.0",
	}
	healthzPorts := []deployer.Port{
		{
			Port:     8080,
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   d.Cluster.Context,
			Namespace: "kube-system",
			Name:      "kube-dns",
			DnsPolicy: "Default",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "kubedns",
					Image: kubednsImage,
					Requirement: &build.Kubedns{
						Image: kubednsImage,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/readiness",
							Port:   8081,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						SuccessThreshold:    1,
						TimeoutSeconds:      5,
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/healthz-kubedns",
							Port:   8080,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					Ports: kubednsPorts,
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "200Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "100Mi",
						},
					},
					Args: []k8s.Arg{
						"--domain=cluster.local.",
						"--dns-port=10053",
					},
				},
				&deployer.DeploymentDeployerContainer{
					Name:  "dnsmasq",
					Image: kubednsMasqImage,
					Requirement: &build.KubednsMasq{
						Image: kubednsMasqImage,
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/healthz-dnsmasq",
							Port:   8080,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					Ports: kubednsMasqPorts,
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "50Mi",
						},
					},
					Args: []k8s.Arg{
						"--cache-size=1000",
						"--no-resolv",
						"--server=127.0.0.1#10053",
						"--log-facility=-",
					},
				},
				&deployer.DeploymentDeployerContainer{
					Name:  "healthz",
					Image: healthzImage,
					Requirement: &build.Healthz{
						Image: healthzImage,
					},
					Ports: healthzPorts,
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "50Mi",
						},
					},
					Args: []k8s.Arg{
						"--cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1 >/dev/null",
						"--url=/healthz-dnsmasq",
						"--cmd=nslookup kubernetes.default.svc.cluster.local 127.0.0.1:10053 >/dev/null",
						"--url=/healthz-kubedns",
						"--port=8080",
						"--quiet",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   d.Cluster.Context,
			Namespace: "kube-system",
			Name:      "kube-dns",
			Ports:     kubednsMasqPorts,
			ClusterIP: "10.103.0.10",
			Labels: k8s.Labels{
				"k8s-app":                       "kube-dns",
				"kubernetes.io/cluster-service": "true",
				"kubernetes.io/name":            "KubeDNS",
			},
		},
	}
}

func (d *KubeDns) Applier() (world.Applier, error) {
	return nil, nil
}

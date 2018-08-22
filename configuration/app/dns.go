package app

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Dns struct {
	Cluster cluster.Cluster
}

func (d *Dns) Children() []world.Configuration {
	kubednsImage := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/kubedns",
		Tag:        "1.8",
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
		Registry:   "docker.io",
		Repository: "bborbe/kube-dnsmasq",
		Tag:        "1.4",
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
		Registry:   "docker.io",
		Repository: "bborbe/exechealthz",
		Tag:        "1.2",
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
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "kubedns",
					Image: kubednsImage,
					Ports: kubednsPorts,
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "500m",
							Memory: "200Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "100m",
							Memory: "100Mi",
						},
					},
					Args: []k8s.Arg{
						"--domain=cluster.local.",
						"--dns-port=10053",
					},
				},
				{
					Name:  "dnsmasq",
					Image: kubednsMasqImage,
					Ports: kubednsMasqPorts,
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "500m",
							Memory: "50Mi",
						},
						Requests: k8s.Resources{
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
				{
					Name:  "dnsmasq",
					Image: healthzImage,
					Ports: healthzPorts,
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "500m",
							Memory: "50Mi",
						},
						Requests: k8s.Resources{
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

func (d *Dns) Applier() (world.Applier, error) {
	return nil, nil
}

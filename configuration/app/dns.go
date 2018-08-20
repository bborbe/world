package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
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
					Name:          "kubedns",
					Image:         kubednsImage,
					Ports:         kubednsPorts,
					CpuLimit:      "500m",
					MemoryLimit:   "200Mi",
					CpuRequest:    "100m",
					MemoryRequest: "100Mi",
					Args: []k8s.Arg{
						"--domain=cluster.local.",
						"--dns-port=10053",
					},
				},
				{
					Name:          "dnsmasq",
					Image:         kubednsMasqImage,
					Ports:         kubednsMasqPorts,
					CpuLimit:      "500m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "50Mi",
					Args: []k8s.Arg{
						"--cache-size=1000",
						"--no-resolv",
						"--server=127.0.0.1#10053",
						"--log-facility=-",
					},
				},
				{
					Name:          "dnsmasq",
					Image:         healthzImage,
					Ports:         healthzPorts,
					CpuLimit:      "500m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "50Mi",
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

func (d *Dns) Applier() world.Applier {
	return nil
}

func (d *Dns) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate dns app ...")
	if err := d.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate dns app failed")
	}
	return nil
}

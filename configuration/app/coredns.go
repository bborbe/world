package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/docker"

	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type CoreDns struct {
	Cluster cluster.Cluster
}

func (c *CoreDns) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Cluster,
	)
}

func (c *CoreDns) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/coredns",
		Tag:        "1.2.2",
	}
	udpPort := deployer.Port{
		Name:     "dns",
		Port:     53,
		Protocol: "UDP",
	}
	tcpPort := deployer.Port{
		Name:     "dns-tcp",
		Port:     53,
		Protocol: "TCP",
	}
	metricsPort := deployer.Port{
		Name:     "metrics",
		Port:     9153,
		Protocol: "TCP",
	}
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   c.Cluster.Context,
				Namespace: "kube-system",
				Name:      "coredns",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "Corefile",
						Value: corefileConfig,
					},
				},
			},
		),
		&build.CoreDns{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: c.Cluster.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kube-system",
					Name:      "coredns",
					Labels: k8s.Labels{
						"k8s-app":            "kube-dns",
						"kubernetes.io/name": "CoreDNS",
					},
				},
				Spec: k8s.DeploymentSpec{
					Replicas: 2,
					Strategy: k8s.DeploymentStrategy{
						Type: "RollingUpdate",
						RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
							MaxUnavailable: 1,
						},
					},
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"k8s-app": "kube-dns",
						},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"k8s-app": "kube-dns",
							},
						},
						Spec: k8s.PodSpec{
							Tolerations: []k8s.Toleration{
								{
									Key:    "node-role.kubernetes.io/master",
									Effect: "NoSchedule",
								},
								{
									Key:      "CriticalAddonsOnly",
									Operator: "Exists",
								},
							},
							Containers: []k8s.Container{
								{
									Name:            "coredns",
									Image:           k8s.Image(image.String()),
									ImagePullPolicy: "IfNotPresent",
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "200m",
											Memory: "170Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "100m",
											Memory: "70Mi",
										},
									},
									Args: []k8s.Arg{"-conf", "/etc/coredns/Corefile"},
									VolumeMounts: []k8s.ContainerMount{
										{
											Name:     "config-volume",
											Path:     "/etc/coredns",
											ReadOnly: true,
										},
									},
									Ports: []k8s.ContainerPort{
										udpPort.ContainerPort(),
										tcpPort.ContainerPort(),
										metricsPort.ContainerPort(),
									},
									SecurityContext: k8s.SecurityContext{
										AllowPrivilegeEscalation: false,
										Capabilities: map[string][]string{
											"add":  {"NET_BIND_SERVICE"},
											"drop": {"all"},
										},
										ReadOnlyRootFilesystem: true,
									},
									LivenessProbe: k8s.Probe{
										HttpGet: k8s.HttpGet{
											Path:   "/health",
											Port:   8080,
											Scheme: "HTTP",
										},
										InitialDelaySeconds: 60,
										TimeoutSeconds:      5,
										SuccessThreshold:    1,
										FailureThreshold:    5,
									},
								},
							},
							DnsPolicy: "Default",
							Volumes: []k8s.PodVolume{
								{
									Name: "config-volume",
									ConfigMap: k8s.PodVolumeConfigMap{
										Name: "coredns",
										Items: []k8s.PodConfigMapItem{
											{
												Key:  "Corefile",
												Path: "Corefile",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		&k8s.ServiceConfiguration{
			Context: c.Cluster.Context,
			Service: k8s.Service{
				ApiVersion: "v1",
				Kind:       "Service",
				Metadata: k8s.Metadata{
					Namespace: "kube-system",
					Name:      "kube-dns",
					Annotations: k8s.Annotations{
						"prometheus.io/port":   "9153",
						"prometheus.io/scrape": "true",
					},
					Labels: k8s.Labels{
						"k8s-app":                       "kube-dns",
						"kubernetes.io/cluster-service": "true",
						"kubernetes.io/name":            "CoreDNS",
					},
				},
				Spec: k8s.ServiceSpec{
					ClusterIP: "10.103.0.10",
					Ports: []k8s.ServicePort{
						udpPort.ServicePort(),
						tcpPort.ServicePort(),
					},
					Selector: k8s.ServiceSelector{
						"k8s-app": "kube-dns",
					},
				},
			},
		},
	}
}

func (c *CoreDns) Applier() (world.Applier, error) {
	return nil, nil
}

const corefileConfig = `.:53 {
    errors
    health
    kubernetes cluster.local in-addr.arpa ip6.arpa {
      pods insecure
      upstream
      fallthrough in-addr.arpa ip6.arpa
    }
    prometheus :9153
    proxy . /etc/resolv.conf
    cache 30
    loop
    reload
    loadbalance
}
`
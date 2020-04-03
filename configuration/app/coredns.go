// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type CoreDns struct {
	Context     k8s.Context
	DisableRBAC bool
}

func (c *CoreDns) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Context,
	)
}
func (c *CoreDns) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/coredns",
		Tag:        "1.6.9", // https://hub.docker.com/r/coredns/coredns/tags
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
	result := []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   c.Context,
				Namespace: "kube-system",
				Name:      "coredns",
				ConfigValues: map[string]deployer.ConfigValue{
					"Corefile": deployer.ConfigValueStatic(corefileConfig),
				},
			},
		),
		&build.CoreDns{
			Image: image,
		},
		&k8s.ServiceaccountConfiguration{
			Context: c.Context,
			Serviceaccount: k8s.ServiceAccount{
				ApiVersion: "v1",
				Kind:       "ServiceAccount",
				Metadata: k8s.Metadata{
					Namespace: "kube-system",
					Name:      "coredns",
				},
			},
		},
		&k8s.DeploymentConfiguration{
			Context: c.Context,
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
							ServiceAccountName: "coredns",
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
									Args: []k8s.Arg{
										"-conf",
										"/etc/coredns/Corefile",
									},
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
											"add": {
												"NET_BIND_SERVICE",
											},
											"drop": {
												"all",
											},
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
			Context: c.Context,
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
	if !c.DisableRBAC {
		result = append(result,
			&k8s.ClusterRoleConfiguration{
				Context: c.Context,
				ClusterRole: k8s.ClusterRole{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Kind:       "ClusterRole",
					Metadata: k8s.Metadata{
						Name: "system:coredns",
						Labels: k8s.Labels{
							"kubernetes.io/bootstrapping": "rbac-defaults",
						},
						Annotations: k8s.Annotations(nil)},
					Rules: []k8s.PolicyRule{
						{
							ApiGroups: []string{
								"",
							},
							Resources: []string{
								"endpoints",
								"services",
								"pods",
								"namespaces",
							},
							Verbs: []string{
								"list",
								"watch",
							},
						},
						{
							ApiGroups: []string{
								"",
							},
							Resources: []string{
								"nodes",
							},
							Verbs: []string{
								"get",
							},
						},
					},
				},
			},
			&k8s.ClusterRoleBindingConfiguration{
				Context: c.Context,
				ClusterRoleBinding: k8s.ClusterRoleBinding{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Kind:       "ClusterRoleBinding",
					Metadata: k8s.Metadata{
						Name: "system:coredns",
						Labels: k8s.Labels{
							"kubernetes.io/bootstrapping": "rbac-defaults",
						},
						Annotations: k8s.Annotations{
							"rbac.authorization.kubernetes.io/autoupdate": "true",
						},
					},
					Subjects: []k8s.Subject{
						{
							Kind:      "ServiceAccount",
							Name:      "coredns",
							Namespace: "kube-system",
						},
					},
					RoleRef: k8s.RoleRef{
						Kind:     "ClusterRole",
						Name:     "system:coredns",
						ApiGroup: "rbac.authorization.k8s.io",
					},
				},
			},
		)
	}
	return result
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
    forward . /etc/resolv.conf
    cache 30
    loop
    reload
    loadbalance
}
`

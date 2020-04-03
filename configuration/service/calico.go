// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package service

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

type Calico struct {
	Context   k8s.Context
	ClusterIP network.IP
}

func (c *Calico) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Context,
		c.ClusterIP,
	)
}
func (c *Calico) Children() []world.Configuration {
	var version docker.Tag = "release-v3.13"
	cniImage := docker.Image{
		Repository: "bborbe/calico-cni",
		Tag:        version,
	}
	kubeControllersImage := docker.Image{
		Repository: "bborbe/calico-kube-controllers",
		Tag:        version,
	}
	nodeImage := docker.Image{
		Repository: "bborbe/calico-node",
		Tag:        version,
	}
	return []world.Configuration{
		&k8s.ClusterRoleConfiguration{
			Context: c.Context,
			ClusterRole: k8s.ClusterRole{
				ApiVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
				Metadata: k8s.Metadata{
					Namespace: "",
					Name:      "calico-kube-controllers",
				},
				Rules: []k8s.PolicyRule{
					{
						ApiGroups: []string{
							"",
							"extensions",
						},
						Resources: []string{
							"pods",
							"namespaces",
							"networkpolicies",
							"nodes",
							"serviceaccounts",
						},
						Verbs: []string{
							"watch",
							"list",
						},
					},
					{
						ApiGroups: []string{
							"networking.k8s.io",
						},
						Resources: []string{
							"networkpolicies",
						},
						Verbs: []string{
							"watch",
							"list",
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
					Namespace: "",
					Name:      "calico-kube-controllers",
				},
				Subjects: []k8s.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "calico-kube-controllers",
						ApiGroup:  "",
						Namespace: "kube-system",
					},
				},
				RoleRef: k8s.RoleRef{
					Kind:     "ClusterRole",
					Name:     "calico-kube-controllers",
					ApiGroup: "rbac.authorization.k8s.io",
				},
			},
		},
		&k8s.ClusterRoleConfiguration{
			Context: c.Context,
			ClusterRole: k8s.ClusterRole{
				ApiVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
				Metadata: k8s.Metadata{
					Namespace: "",
					Name:      "calico-node",
				},
				Rules: []k8s.PolicyRule{
					{
						ApiGroups: []string{
							"",
						},
						Resources: []string{
							"pods",
							"nodes",
							"namespaces",
							"configmaps",
						},
						Verbs: []string{
							"get",
						},
					},
					{
						ApiGroups: []string{
							"",
						},
						Resources: []string{
							"nodes/status",
						},
						Verbs: []string{
							"patch",
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
					Namespace: "",
					Name:      "calico-node",
				},
				Subjects: []k8s.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "calico-node",
						ApiGroup:  "",
						Namespace: "kube-system",
					},
				},
				RoleRef: k8s.RoleRef{
					Kind:     "ClusterRole",
					Name:     "calico-node",
					ApiGroup: "rbac.authorization.k8s.io",
				},
			},
		},
		&k8s.ServiceaccountConfiguration{
			Context: c.Context,
			Serviceaccount: k8s.ServiceAccount{
				ApiVersion: "v1",
				Kind:       "ServiceAccount",
				Metadata: k8s.Metadata{
					Namespace: "kube-system",
					Name:      "calico-kube-controllers",
				},
			},
		},
		&k8s.ServiceaccountConfiguration{
			Context: c.Context,
			Serviceaccount: k8s.ServiceAccount{
				ApiVersion: "v1",
				Kind:       "ServiceAccount",
				Metadata: k8s.Metadata{
					Namespace: "kube-system",
					Name:      "calico-node",
				},
			},
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   c.Context,
				Namespace: "kube-system",
				Name:      "calico-config",
				ConfigValues: map[string]deployer.ConfigValue{
					"etcd_ca":   deployer.ConfigValueStatic(""),
					"etcd_cert": deployer.ConfigValueStatic(""),
					"etcd_key":  deployer.ConfigValueStatic(""),
					"etcd_endpoints": deployer.ConfigValueFunc(func(ctx context.Context) (string, error) {
						ip, err := c.ClusterIP.IP(ctx)
						if err != nil {
							return "", errors.Wrap(err, "get ip failed")
						}
						return fmt.Sprintf("http://%s:2379", ip), nil
					}),
					"calico_backend":     deployer.ConfigValueStatic("bird"),
					"veth_mtu":           deployer.ConfigValueStatic("1440"),
					"cni_network_config": deployer.ConfigValueStatic("{\n  \"name\": \"k8s-pod-network\",\n  \"cniVersion\": \"0.3.0\",\n  \"plugins\": [\n    {\n      \"type\": \"calico\",\n      \"log_level\": \"info\",\n      \"etcd_endpoints\": \"__ETCD_ENDPOINTS__\",\n      \"etcd_key_file\": \"__ETCD_KEY_FILE__\",\n      \"etcd_cert_file\": \"__ETCD_CERT_FILE__\",\n      \"etcd_ca_cert_file\": \"__ETCD_CA_CERT_FILE__\",\n      \"mtu\": __CNI_MTU__,\n      \"ipam\": {\n          \"type\": \"calico-ipam\"\n      },\n      \"policy\": {\n          \"type\": \"k8s\"\n      },\n      \"kubernetes\": {\n          \"kubeconfig\": \"__KUBECONFIG_FILEPATH__\"\n      }\n    },\n    {\n      \"type\": \"portmap\",\n      \"snat\": true,\n      \"capabilities\": {\"portMappings\": true}\n    }\n  ]\n}"),
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.SecretApplier{
				Context:   c.Context,
				Namespace: "kube-system",
				Name:      "calico-etcd-secrets",
				Secrets:   deployer.Secrets{},
			},
		),
		&k8s.DaemonSetConfiguration{
			Requirements: []world.Configuration{
				&build.CalicoCNI{
					Image: cniImage,
				},
				&build.CalicoNode{
					Image: nodeImage,
				},
			},
			Context: c.Context,
			DaemonSet: k8s.DaemonSet{
				ApiVersion: "apps/v1",
				Kind:       "DaemonSet",
				Metadata: k8s.Metadata{
					Namespace: "kube-system",
					Name:      "calico-node",
					Labels: k8s.Labels{
						"k8s-app": "calico-node",
					},
				},
				Spec: k8s.DaemonSetSpec{
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"k8s-app": "calico-node",
						},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"k8s-app": "calico-node",
							},
							Annotations: k8s.Annotations{
								"scheduler.alpha.kubernetes.io/critical-pod": "",
							},
						},
						Spec: k8s.PodSpec{
							Tolerations: []k8s.Toleration{
								{
									Key:      "",
									Effect:   "NoSchedule",
									Operator: "Exists",
								},
								{
									Key:      "CriticalAddonsOnly",
									Effect:   "",
									Operator: "Exists",
								},
								{
									Key:      "",
									Effect:   "NoExecute",
									Operator: "Exists",
								},
							},
							TerminationGracePeriodSeconds: 0,
							HostNetwork:                   true,
							ServiceAccountName:            "calico-node",
							Containers: []k8s.Container{
								{
									Name:  "calico-node",
									Image: k8s.Image(nodeImage.String()),
									Env: []k8s.Env{
										{
											Name: "ETCD_ENDPOINTS",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "etcd_endpoints",
												},
											},
										},
										{
											Name: "ETCD_CA_CERT_FILE",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "etcd_ca",
												},
											},
										},
										{
											Name: "ETCD_KEY_FILE",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "etcd_key",
												},
											},
										},
										{
											Name: "ETCD_CERT_FILE",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "etcd_cert",
												},
											},
										},
										{
											Name: "CALICO_K8S_NODE_REF",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "spec.nodeName",
												},
											},
										},
										{
											Name: "CALICO_NETWORKING_BACKEND",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "calico_backend",
												},
											},
										},
										{
											Name:  "CLUSTER_TYPE",
											Value: "k8s,bgp",
										},
										{
											Name:  "IP",
											Value: "autodetect",
										},
										{
											Name:  "CALICO_IPV4POOL_IPIP",
											Value: "Always",
										},
										{
											Name: "FELIX_IPINIPMTU",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "veth_mtu",
												},
											},
										},
										{
											Name:  "CALICO_IPV4POOL_CIDR",
											Value: "10.103.0.0/16",
										},
										{
											Name:  "CALICO_DISABLE_FILE_LOGGING",
											Value: "true",
										},
										{
											Name:  "FELIX_DEFAULTENDPOINTTOHOSTACTION",
											Value: "ACCEPT",
										},
										{
											Name:  "FELIX_IPV6SUPPORT",
											Value: "false",
										},
										{
											Name:  "FELIX_LOGSEVERITYSCREEN",
											Value: "info",
										},
										{
											Name:  "FELIX_HEALTHENABLED",
											Value: "true",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "500m",
											Memory: "200Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "250m",
											Memory: "10Mi",
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path:     "/lib/modules",
											Name:     "lib-modules",
											ReadOnly: true,
										},
										{
											Path: "/run/xtables.lock",
											Name: "xtables-lock",
										},
										{
											Path: "/var/run/calico",
											Name: "var-run-calico",
										},
										{
											Path: "/var/lib/calico",
											Name: "var-lib-calico",
										},
										{
											Path: "/calico-secrets",
											Name: "etcd-certs",
										},
									},
									ReadinessProbe: k8s.Probe{
										Exec: k8s.Exec{
											Command: []k8s.Command{
												"/bin/calico-node",
												"-bird-ready",
												"-felix-ready",
											},
										},
										PeriodSeconds: 10,
									},
									LivenessProbe: k8s.Probe{
										HttpGet: k8s.HttpGet{
											Host: "localhost",
											Path: "/liveness",
											Port: 9099,
										},
										InitialDelaySeconds: 10,
										FailureThreshold:    6,
										PeriodSeconds:       10,
									},
									SecurityContext: k8s.SecurityContext{
										Privileged: true,
									},
								},
								{
									Name:  "install-cni",
									Image: k8s.Image(cniImage.String()),
									Command: []k8s.Command{
										"/install-cni.sh",
									},
									Env: []k8s.Env{
										{
											Name:  "CNI_CONF_NAME",
											Value: "10-calico.conflist",
										},
										{
											Name: "ETCD_ENDPOINTS",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "etcd_endpoints",
												},
											},
										},
										{
											Name: "CNI_NETWORK_CONFIG",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "cni_network_config",
												},
											},
										},
										{
											Name: "CNI_MTU",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "veth_mtu",
												},
											},
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "500m",
											Memory: "200Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/host/opt/cni/bin",
											Name: "cni-bin-dir",
										},
										{
											Path: "/host/etc/cni/net.d",
											Name: "cni-net-dir",
										},
										{
											Path: "/calico-secrets",
											Name: "etcd-certs",
										},
									},
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "lib-modules",
									Host: k8s.PodVolumeHost{
										Path: "/lib/modules",
									},
								},
								{
									Name: "var-run-calico",
									Host: k8s.PodVolumeHost{
										Path: "/var/run/calico",
									},
								},
								{
									Name: "var-lib-calico",
									Host: k8s.PodVolumeHost{
										Path: "/var/lib/calico",
									},
								},
								{
									Name: "xtables-lock",
									Host: k8s.PodVolumeHost{
										Path: "/run/xtables.lock",
									},
								},
								{
									Name: "cni-bin-dir",
									Host: k8s.PodVolumeHost{
										Path: "/opt/cni/bin",
									},
								},
								{
									Name: "cni-net-dir",
									Host: k8s.PodVolumeHost{
										Path: "/etc/cni/net.d",
									},
								},
								{
									Name: "etcd-certs",
									Secret: k8s.PodVolumeSecret{
										Name: "calico-etcd-secrets",
									},
								},
							},
						},
					},
				},
			},
		},
		&k8s.DeploymentConfiguration{
			Requirements: []world.Configuration{
				&build.CalicoKubeController{
					Image: kubeControllersImage,
				},
			},
			Context: c.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "kube-system",
					Name:      "calico-kube-controllers",
					Labels: k8s.Labels{
						"k8s-app": "calico-kube-controllers",
					},
					Annotations: k8s.Annotations{
						"scheduler.alpha.kubernetes.io/critical-pod": "",
					},
				},
				Spec: k8s.DeploymentSpec{
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"k8s-app": "calico-kube-controllers",
						},
					},
					Replicas: 1,
					Strategy: k8s.DeploymentStrategy{
						Type: "Recreate",
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Namespace: "kube-system",
							Name:      "calico-kube-controllers",
							Labels: k8s.Labels{
								"k8s-app": "calico-kube-controllers",
							},
						},
						Spec: k8s.PodSpec{
							Tolerations: []k8s.Toleration{
								{
									Key:      "CriticalAddonsOnly",
									Effect:   "",
									Operator: "Exists",
								},
								{
									Key:      "node-role.kubernetes.io/master",
									Effect:   "NoSchedule",
									Operator: "",
								},
							},
							Containers: []k8s.Container{
								{
									Name:  "calico-kube-controllers",
									Image: k8s.Image(kubeControllersImage.String()),
									Env: []k8s.Env{
										{
											Name: "ETCD_ENDPOINTS",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "etcd_endpoints",
												},
											},
										},
										{
											Name: "ETCD_CA_CERT_FILE",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "etcd_ca",
												},
											},
										},
										{
											Name: "ETCD_KEY_FILE",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "etcd_key",
												},
											},
										},
										{
											Name: "ETCD_CERT_FILE",
											ValueFrom: k8s.ValueFrom{
												ConfigMapKeyRef: k8s.ConfigMapKeyRef{
													Name: "calico-config",
													Key:  "etcd_cert",
												},
											},
										},
										{
											Name:  "ENABLED_CONTROLLERS",
											Value: "policy,namespace,serviceaccount,workloadendpoint,node",
										},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Path: "/calico-secrets",
											Name: "etcd-certs",
										},
									},
									ReadinessProbe: k8s.Probe{
										Exec: k8s.Exec{
											Command: []k8s.Command{
												"/usr/bin/check-status",
												"-r",
											},
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "500m",
											Memory: "200Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
									},
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "etcd-certs",
									Secret: k8s.PodVolumeSecret{
										Name: "calico-etcd-secrets",
									},
								},
							},
							HostNetwork:        true,
							ServiceAccountName: "calico-kube-controllers",
						},
					},
				},
			},
		},
	}
}
func (c *Calico) Applier() (world.Applier, error) {
	return nil, nil
}

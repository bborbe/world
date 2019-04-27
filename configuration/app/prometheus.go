// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Prometheus struct {
	Context            k8s.Context
	PrometheusDomain   k8s.IngressHost
	AlertmanagerDomain k8s.IngressHost
	Secret             deployer.SecretValue
	LdapUsername       deployer.SecretValue
	LdapPassword       deployer.SecretValue
}

func (p *Prometheus) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.Context,
		p.PrometheusDomain,
		p.AlertmanagerDomain,
		p.Secret,
		p.LdapUsername,
		p.LdapPassword,
	)
}

func (p *Prometheus) Children() []world.Configuration {
	result := []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: p.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "prometheus",
					Name:      "prometheus",
				},
			},
		},
	}
	result = append(result, p.prometheus()...)
	result = append(result, p.alertmanager()...)
	result = append(result, p.nodeExporter()...)
	result = append(result, p.kubeStateMetrics()...)
	return result
}

func (p *Prometheus) prometheus() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/prometheus",
		Tag:        "v2.4.2",
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
		&k8s.ServiceaccountConfiguration{
			Context: p.Context,
			Serviceaccount: k8s.ServiceAccount{
				ApiVersion: "v1",
				Kind:       "ServiceAccount",
				Metadata: k8s.Metadata{
					Namespace: "prometheus",
					Name:      "prometheus",
				},
			},
		},
		&k8s.ClusterRoleConfiguration{
			Context: p.Context,
			ClusterRole: k8s.ClusterRole{
				ApiVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
				Metadata: k8s.Metadata{
					Namespace: "",
					Name:      "prometheus",
				},
				Rules: []k8s.PolicyRule{
					{
						ApiGroups: []string{""},
						Resources: []string{
							"nodes",
							"services",
							"endpoints",
							"pods",
						},
						Verbs: []string{
							"get",
							"list",
							"watch",
						},
					},
					{
						ApiGroups: []string{""},
						Resources: []string{
							"configmaps",
						},
						Verbs: []string{
							"get",
						},
					},
					{
						NonResourceURLs: []string{
							"/metrics",
						},
						Verbs: []string{
							"get",
						},
					},
				},
			},
		},
		&k8s.ClusterRoleBindingConfiguration{
			Context: p.Context,
			ClusterRoleBinding: k8s.ClusterRoleBinding{
				ApiVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
				Metadata: k8s.Metadata{
					Name: "prometheus",
				},
				Subjects: []k8s.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "prometheus",
						Namespace: "prometheus",
					},
				},
				RoleRef: k8s.RoleRef{
					Kind:     "ClusterRole",
					Name:     "prometheus",
					ApiGroup: "rbac.authorization.k8s.io",
				},
			},
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   p.Context,
				Namespace: "prometheus",
				Name:      "prometheus",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "prometheus.yaml",
						Value: prometheusConfig,
					},
					deployer.ConfigEntry{
						Key:   "alert.rules.yaml",
						Value: prometheusAlertRulesConfig,
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   p.Context,
			Namespace: "prometheus",
			Name:      "prometheus",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			ServiceAccountName: "prometheus",
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "prometheus",
					Image: image,
					Args: []k8s.Arg{
						"--config.file=/config/prometheus.yaml",
						"--storage.tsdb.retention=48h",
						"--storage.tsdb.path=/prometheus",
						"--web.console.libraries=/etc/prometheus/console_libraries",
						"--web.console.templates=/etc/prometheus/consoles",
						k8s.Arg(fmt.Sprintf("--web.external-url=https://%s", p.PrometheusDomain)),
						"--web.enable-lifecycle",
						"--log.level=info",
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
					Context:      p.Context,
					Namespace:    "prometheus",
					Port:         authPort.Port,
					TargetPort:   prometheusPort.Port,
					Secret:       p.Secret,
					LdapUsername: p.LdapUsername,
					LdapPassword: p.LdapPassword,
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "prometheus",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "prometheus.yaml",
								Path: "prometheus.yaml",
							},
							{
								Key:  "alert.rules.yaml",
								Path: "alert.rules.yaml",
							},
						},
					},
				},
				{
					Name: "prometheus",
					Host: k8s.PodVolumeHost{
						Path: "/data/prometheus",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Context,
			Namespace: "prometheus",
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
			Context:   p.Context,
			Namespace: "prometheus",
			Name:      "prometheus",
			Port:      "http-auth",
			Domains:   k8s.IngressHosts{p.PrometheusDomain},
		},
	}
}

func (p *Prometheus) alertmanager() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/prometheus-alertmanager",
		Tag:        "v0.15.2",
	}
	alertmanagerPort := deployer.Port{
		Port:     9093,
		Name:     "http",
		Protocol: "TCP",
	}
	authPort := deployer.Port{
		Port:     9095,
		Name:     "http-auth",
		Protocol: "TCP",
	}
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   p.Context,
				Namespace: "prometheus",
				Name:      "alertmanager",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "alertmanager.yaml",
						Value: prometheusAlertmanagerConfig,
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   p.Context,
			Namespace: "prometheus",
			Name:      "alertmanager",
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
						"--storage.path=/alertmanager",
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
					Context:      p.Context,
					Namespace:    "prometheus",
					Port:         authPort.Port,
					TargetPort:   alertmanagerPort.Port,
					Secret:       p.Secret,
					LdapUsername: p.LdapUsername,
					LdapPassword: p.LdapPassword,
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "alertmanager",
					Host: k8s.PodVolumeHost{
						Path: "/data/alertmanager",
					},
				},
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "alertmanager",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "alertmanager.yaml",
								Path: "alertmanager.yaml",
							},
						},
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Context,
			Namespace: "prometheus",
			Name:      "alertmanager",
			Ports: []deployer.Port{
				alertmanagerPort,
				authPort,
			},
		},
		&deployer.IngressDeployer{
			Context:   p.Context,
			Namespace: "prometheus",
			Name:      "alertmanager",
			Port:      "http-auth",
			Domains:   k8s.IngressHosts{p.AlertmanagerDomain},
		},
	}
}

func (p *Prometheus) nodeExporter() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/prometheus-node-exporter",
		Tag:        "v0.16.0",
	}
	return []world.Configuration{
		&k8s.DaemonSetConfiguration{
			Requirements: []world.Configuration{
				&build.PrometheusNodeExporter{
					Image: image,
				},
			},
			Context: p.Context,
			DaemonSet: k8s.DaemonSet{
				ApiVersion: "apps/v1",
				Kind:       "DaemonSet",
				Metadata: k8s.Metadata{
					Name:      "node-exporter",
					Namespace: "prometheus",
					Labels: k8s.Labels{
						"app": "node-exporter",
					},
				},
				Spec: k8s.DaemonSetSpec{
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"app": "node-exporter",
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
								"app": "node-exporter",
							},
						},
						Spec: k8s.PodSpec{
							HostPid:   true,
							DnsPolicy: "ClusterFirst",
							Containers: []k8s.Container{
								{
									Args: []k8s.Arg{
										"--path.procfs=/proc_host",
										"--path.sysfs=/host_sys",
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
									VolumeMounts: []k8s.ContainerMount{
										{
											Name:     "proc",
											Path:     "/proc_host",
											ReadOnly: true,
										},
										{
											Name:     "sys",
											Path:     "/host_sys",
											ReadOnly: true,
										},
									},
								},
							},
							Volumes: []k8s.PodVolume{
								{
									Name: "proc",
									Host: k8s.PodVolumeHost{
										Path: "/proc",
									},
								},
								{
									Name: "sys",
									Host: k8s.PodVolumeHost{
										Path: "/sys",
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

func (p *Prometheus) kubeStateMetrics() []world.Configuration {
	port := deployer.Port{
		Port:     8080,
		Protocol: "TCP",
		Name:     "http",
	}
	image := docker.Image{
		Repository: "bborbe/kube-state-metrics",
		Tag:        "v1.4.0",
	}
	return []world.Configuration{
		&build.KubeStateMetrics{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: p.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "prometheus",
					Name:      "kube-state-metrics",
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
							"app": "kube-state-metrics",
						},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "kube-state-metrics",
							},
							Annotations: k8s.Annotations{
								"prometheus.io/path":   "/metrics",
								"prometheus.io/port":   port.Port.String(),
								"prometheus.io/scheme": "http",
								"prometheus.io/scrape": "true",
							},
						},
						Spec: k8s.PodSpec{
							Containers: []k8s.Container{
								{
									Name:            "kube-state-metrics",
									Image:           k8s.Image(image.String()),
									ImagePullPolicy: "IfNotPresent",
									Ports: []k8s.ContainerPort{
										port.ContainerPort(),
									},
									LivenessProbe: k8s.Probe{
										HttpGet: k8s.HttpGet{
											Path:   "/",
											Port:   port.Port,
											Scheme: "HTTP",
										},
										InitialDelaySeconds: 20,
										SuccessThreshold:    1,
										FailureThreshold:    5,
										TimeoutSeconds:      5,
									},
									ReadinessProbe: k8s.Probe{
										HttpGet: k8s.HttpGet{
											Path:   "/",
											Port:   port.Port,
											Scheme: "HTTP",
										},
										InitialDelaySeconds: 3,
										TimeoutSeconds:      5,
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "2000m",
											Memory: "2000Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "10m",
											Memory: "10Mi",
										},
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

const prometheusAlertRulesConfig = `
groups:
- name: alert.rules
  rules:
  - alert: InstanceDown
    expr: up == 0
    for: 5m
    labels:
      severity: critical
    annotations:
      description: '{{ $labels.instance }} of job {{ $labels.job }} has been down
        for more than 5 minutes.'
      summary: Instance {{ $labels.instance }} down
  - alert: DiskOutOfSpace
    expr: (max by(app) (diskstatus_bytesfree / diskstatus_bytestotal)) < 0.2
    for: 1m
    labels:
      severity: warning
    annotations:
      description: '{{ $labels.app }} is below 20% free space for more than 1 minutes.'
      summary: Disk of {{ $labels.app }} space
  - alert: DiskOutOfSpace
    expr: (max by(app) (diskstatus_bytesfree / diskstatus_bytestotal)) < 0.1
    for: 1m
    labels:
      severity: critical
    annotations:
      description: '{{ $labels.app }} is below 10% free space for more than 1 minutes.'
      summary: Disk of {{ $labels.app }} space
  - alert: DiskOutOfInodes
    expr: (max by(app) (diskstatus_inodesfree / diskstatus_inodestotal)) < 0.2
    for: 1m
    labels:
      severity: warning
    annotations:
      description: '{{ $labels.app }} is below 20% free inodes for more than 1 minutes.'
      summary: Disk of {{ $labels.app }} inodes
  - alert: DiskOutOfInodes
    expr: (max by(app) (diskstatus_inodesfree / diskstatus_inodestotal)) < 0.1
    for: 1m
    labels:
      severity: critical
    annotations:
      description: '{{ $labels.app }} is below 10% free inodes for more than 1 minutes.'
      summary: Disk of {{ $labels.app }} inodes
`

const prometheusAlertmanagerConfig = `
global:
  resolve_timeout: 1m
  smtp_smarthost: 'mail.benjamin-borbe.de:25'
  smtp_from: 'alertmanager@rocketnews.de'
  smtp_require_tls: false

# The directory from which notification templates are read.
templates:
- '/etc/alertmanager/template/*.tmpl'

# The root route on which each incoming alert enters.
route:
  # The labels by which incoming alerts are grouped together. For example,
  # multiple alerts coming in for cluster=A and alertname=LatencyHigh would
  # be batched into a single group.
  group_by: ['alertname', 'cluster', 'service']

  # When a new group of alerts is created by an incoming alert, wait at
  # least 'group_wait' to send the initial notification.
  # This way ensures that you get multiple alerts for the same group that start
  # firing shortly after another are batched together on the first
  # notification.
  group_wait: 30s

  # When the first notification was sent, wait 'group_interval' to send a batch
  # of new alerts that started firing for that group.
  group_interval: 5m

  # If an alert has successfully been sent, wait 'repeat_interval' to
  # resend them.
  repeat_interval: 3h

  # A default receiver
  receiver: nc-monitoring

# Inhibition rules allow to mute a set of alerts given that another alert is
# firing.
# We use this to mute any warning-level notifications if the same alert is
# already critical.
inhibit_rules:
- source_match:
    severity: 'critical'
  target_match:
    severity: 'warning'
  # Apply inhibition if the alertname is the same.
  equal: ['alertname', 'cluster', 'service']


receivers:
- name: 'nc-monitoring'
  email_configs:
  - to: 'bborbe@rocketnews.de'
`

const prometheusConfig = `
global:
  # How frequently to scrape targets by default.
  scrape_interval: 1m

  # How long until a scrape request times out.
  scrape_timeout: 10s

  # How frequently to evaluate rules.
  evaluation_interval: 1m

  # The labels to add to any time series or alerts when communicating with
  # external systems (federation, remote storage, Alertmanager).
  external_labels:
    region: 'nc'

alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - alertmanager:9093
    scheme: http
    timeout: 10s

# Rule files specifies a list of globs. Rules and alerts are read from
# all matching files.
rule_files:
- 'alert.rules.yaml'

# A scrape configuration for running Prometheus on a Kubernetes cluster.
# This uses separate scrape configs for cluster components (i.e. API server, node)
# and services to allow each to use different authentication configs.
#
# Kubernetes labels will be added as Prometheus labels on metrics via the
# 'labelmap' relabeling action.

# Scrape config for API servers.
scrape_configs:
- job_name: 'kubernetes-apiservers'

  # Default to scraping over https. If required, just disable this or change to
  # 'http'.
  scheme: https

  # This TLS & bearer token file config is used to connect to the actual scrape
  # endpoints for cluster components. This is separate to discovery auth
  # configuration ('in_cluster' below) because discovery & scraping are two
  # separate concerns in Prometheus.
  tls_config:
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    # If your node certificates are self-signed or use a different CA to the
    # master CA, then disable certificate verification below. Note that
    # certificate verification is an integral part of a secure infrastructure
    # so this should only be disabled in a controlled environment. You can
    # disable certificate verification by uncommenting the line below.
    #
    insecure_skip_verify: true
  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

  # Keep only the default/kubernetes service endpoints for the https port. This
  # will add targets for each API server which Kubernetes adds an endpoint to
  # the default/kubernetes service.
  relabel_configs:
  - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
    action: keep
    regex: default;kubernetes;https

- job_name: 'kubernetes-nodes'

  # Default to scraping over https. If required, just disable this or change to
  # 'http'.
  scheme: https



  # This TLS & bearer token file config is used to connect to the actual scrape
  # endpoints for cluster components. This is separate to discovery auth
  # configuration ('in_cluster' below) because discovery & scraping are two
  # separate concerns in Prometheus.
  tls_config:
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    # If your node certificates are self-signed or use a different CA to the
    # master CA, then disable certificate verification below. Note that
    # certificate verification is an integral part of a secure infrastructure
    # so this should only be disabled in a controlled environment. You can
    # disable certificate verification by uncommenting the line below.
    #
    insecure_skip_verify: true
  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

  kubernetes_sd_configs:
  - role: node

  relabel_configs:
  - action: labelmap
    regex: __meta_kubernetes_node_label_(.+)
  - source_labels: [__address__]
    action: replace
    target_label: __address__
    regex: ([^:;]+):(\d+)
    replacement: ${1}:10255
  - source_labels: [__scheme__]
    action: replace
    target_label: __scheme__
    regex: https
    replacement: http


- job_name: 'kubernetes-cadvisor'

  # Default to scraping over https. If required, just disable this or change to
  # 'http'.
  scheme: https
  metrics_path: /metrics/cadvisor

  # This TLS & bearer token file config is used to connect to the actual scrape
  # endpoints for cluster components. This is separate to discovery auth
  # configuration ('in_cluster' below) because discovery & scraping are two
  # separate concerns in Prometheus.
  tls_config:
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    # If your node certificates are self-signed or use a different CA to the
    # master CA, then disable certificate verification below. Note that
    # certificate verification is an integral part of a secure infrastructure
    # so this should only be disabled in a controlled environment. You can
    # disable certificate verification by uncommenting the line below.
    #
    insecure_skip_verify: true
  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

  kubernetes_sd_configs:
  - role: node

  relabel_configs:
  - action: labelmap
    regex: __meta_kubernetes_node_label_(.+)
  - source_labels: [__address__]
    action: replace
    target_label: __address__
    regex: ([^:;]+):(\d+)
    replacement: ${1}:10255
  - source_labels: [__scheme__]
    action: replace
    target_label: __scheme__
    regex: https
    replacement: http

# Scrape config for service endpoints.
#
# The relabeling allows the actual service scrape endpoint to be configured
# via the following annotations:
#
# * 'prometheus.io/scrape': Only scrape services that have a value of 'true'
# * 'prometheus.io/scheme': If the metrics endpoint is secured then you will need
# to set this to 'https' & most likely set the 'tls_config' of the scrape config.
# * 'prometheus.io/path': If the metrics path is not '/metrics' override this.
# * 'prometheus.io/port': If the metrics are exposed on a different port to the
# service then set this appropriately.
- job_name: 'kubernetes-service-endpoints'

  kubernetes_sd_configs:
  - role: endpoints

  relabel_configs:
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
    action: keep
    regex: true
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
    action: replace
    target_label: __scheme__
    regex: (https?)
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
    action: replace
    target_label: __metrics_path__
    regex: (.+)
  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
    action: replace
    target_label: __address__
    regex: (.+)(?::\d+);(\d+)
    replacement: $1:$2
  - action: labelmap
    regex: __meta_kubernetes_service_label_(.+)
  - source_labels: [__meta_kubernetes_service_namespace]
    action: replace
    target_label: kubernetes_namespace
  - source_labels: [__meta_kubernetes_service_name]
    action: replace
    target_label: kubernetes_name

# Example scrape config for probing services via the Blackbox Exporter.
#
# The relabeling allows the actual service scrape endpoint to be configured
# via the following annotations:
#
# * 'prometheus.io/probe': Only probe services that have a value of 'true'
- job_name: 'kubernetes-services'

  metrics_path: /probe
  params:
    module: [http_2xx]

  kubernetes_sd_configs:
  - role: service

  relabel_configs:
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_probe]
    action: keep
    regex: true
  - source_labels: [__address__]
    target_label: __param_target
  - target_label: __address__
    replacement: blackbox
  - source_labels: [__param_target]
    target_label: instance
  - action: labelmap
    regex: __meta_kubernetes_service_label_(.+)
  - source_labels: [__meta_kubernetes_service_namespace]
    target_label: kubernetes_namespace
  - source_labels: [__meta_kubernetes_service_name]
    target_label: kubernetes_name


# Scrape config for probing services via the Blackbox Exporter.
#
# The relabeling allows the actual service scrape endpoint to be configured
# via the following annotations:
#
# * 'prometheus.io/probehttp': Only probe services that have a value of 'true'
# * 'prometheus.io/fqdn': The external address to check
- job_name: 'kubernetes-services-http'

  metrics_path: /probe
  params:
    module: [http_2xx]

  kubernetes_sd_configs:
  - role: service

  relabel_configs:
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_probehttp]
    action: keep
    regex: true
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_target]
    target_label: __param_target
  - source_labels: [__meta_kubernetes_service_name]
    target_label: __address__
  - source_labels: [__param_target]
    target_label: instance
  - action: labelmap
    regex: __meta_kubernetes_service_label_(.+)
  - source_labels: [__meta_kubernetes_service_namespace]
    target_label: kubernetes_namespace
  - source_labels: [__meta_kubernetes_service_name]
    target_label: kubernetes_name

# Scrape config for probing services via the Blackbox Exporter.
#
# The relabeling allows the actual service scrape endpoint to be configured
# via the following annotations:
#
# * 'prometheus.io/probehttp': Only probe services that have a value of 'true'
# * 'prometheus.io/fqdn': The external address to check
- job_name: 'kubernetes-services-https'

  metrics_path: /probe
  params:
    module: [https_2xx]

  kubernetes_sd_configs:
  - role: service

  relabel_configs:
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_probehttps]
    action: keep
    regex: true
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_target]
    target_label: __param_target
  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_service_namespace]
    target_label: __address__
    action: replace
    regex: (.+);(.+)
    replacement: $1.$2.svc
  - source_labels: [__param_target]
    target_label: instance
  - action: labelmap
    regex: __meta_kubernetes_service_label_(.+)
  - source_labels: [__meta_kubernetes_service_namespace]
    target_label: kubernetes_namespace
  - source_labels: [__meta_kubernetes_service_name]
    target_label: kubernetes_name

# Example scrape config for pods
#
# The relabeling allows the actual pod scrape endpoint to be configured via the
# following annotations:
#
# * 'prometheus.io/scrape': Only scrape pods that have a value of 'true'
# * 'prometheus.io/path': If the metrics path is not '/metrics' override this.
# * 'prometheus.io/port': Scrape the pod on the indicated port instead of the default of '9102'.
- job_name: 'kubernetes-pods'

  kubernetes_sd_configs:
  - role: pod

  relabel_configs:
  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
    action: keep
    regex: true
  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
    action: replace
    target_label: __metrics_path__
    regex: (.+)
  - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
    action: replace
    regex: (.+):(?:\d+);(\d+)
    replacement: ${1}:${2}
    target_label: __address__
  - action: labelmap
    regex: __meta_kubernetes_pod_label_(.+)
  - source_labels: [__meta_kubernetes_pod_namespace]
    action: replace
    target_label: kubernetes_namespace
  - source_labels: [__meta_kubernetes_pod_name]
    action: replace
    target_label: kubernetes_pod_name

`

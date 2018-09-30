package app

import (
	"context"

	"github.com/bborbe/world/pkg/configuration"

	"fmt"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Prometheus struct {
	Cluster            cluster.Cluster
	PrometheusDomain   k8s.IngressHost
	AlertmanagerDomain k8s.IngressHost
	Secret             deployer.SecretValue
	LdapUsername       deployer.SecretValue
	LdapPassword       deployer.SecretValue
}

func (t *Prometheus) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Cluster,
		t.PrometheusDomain,
		t.AlertmanagerDomain,
		t.Secret,
		t.LdapUsername,
		t.LdapPassword,
	)
}

func (p *Prometheus) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, p.nodeExporter()...)
	result = append(result, p.prometheus()...)
	result = append(result, p.alertmanager()...)
	return result
}

func (p *Prometheus) prometheus() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/prometheus",
		Tag:        "v1.8.2",
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
		configuration.New().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   p.Cluster.Context,
				Namespace: "kube-system",
				Name:      "prometheus",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "prometheus.yaml",
						Value: prometheusConfig,
					},
					deployer.ConfigEntry{
						Key:   "alert.rules",
						Value: prometheusAlertRulesConfig,
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "prometheus",
					Image: image,
					Args: []k8s.Arg{
						"-config.file=/config/prometheus.yaml",
						"-storage.local.retention=48h",
						"-storage.local.path=/prometheus",
						"-storage.local.target-heap-size=500000000",
						"-web.console.libraries=/etc/prometheus/console_libraries",
						"-web.console.templates=/etc/prometheus/consoles",
						k8s.Arg(fmt.Sprintf("-web.external-url=https://%s", p.PrometheusDomain)),
						"-alertmanager.url=http://prometheus-alertmanager:9093",
						"-log.level=info",
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
					Context:      p.Cluster.Context,
					Namespace:    "kube-system",
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
								Key:  "alert.rules",
								Path: "alert.rules",
							},
						},
					},
				},
				{
					Name: "prometheus",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/prometheus",
						Server: k8s.PodNfsServer(p.Cluster.NfsServer),
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
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
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus",
			Port:      "http-auth",
			Domains:   k8s.IngressHosts{p.PrometheusDomain},
		},
	}
}

func (p *Prometheus) alertmanager() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/prometheus-alertmanager",
		Tag:        "v0.14.0",
	}
	alertmanagerPort := deployer.Port{
		Port:     9093,
		Name:     "http",
		Protocol: "TCP",
	}
	authPort := deployer.Port{
		Port:     9094,
		Name:     "http-auth",
		Protocol: "TCP",
	}
	return []world.Configuration{
		configuration.New().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   p.Cluster.Context,
				Namespace: "kube-system",
				Name:      "prometheus-alertmanager",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "alertmanager.yaml",
						Value: prometheusAlertmanagerConfig,
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus-alertmanager",
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
					Context:      p.Cluster.Context,
					Namespace:    "kube-system",
					Port:         authPort.Port,
					TargetPort:   alertmanagerPort.Port,
					Secret:       p.Secret,
					LdapUsername: p.LdapUsername,
					LdapPassword: p.LdapPassword,
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "alertmanager",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "prometheus-alertmanager",
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
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus-alertmanager",
			Ports: []deployer.Port{
				alertmanagerPort,
				authPort,
			},
		},
		&deployer.IngressDeployer{
			Context:   p.Cluster.Context,
			Namespace: "kube-system",
			Name:      "prometheus-alertmanager",
			Port:      "http-auth",
			Domains:   k8s.IngressHosts{p.AlertmanagerDomain},
		},
	}
}

func (p *Prometheus) nodeExporter() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/prometheus-node-exporter",
		Tag:        "v0.14.0",
	}
	return []world.Configuration{
		&k8s.DaemonSetConfiguration{
			Requirements: []world.Configuration{
				&build.PrometheusNodeExporter{
					Image: image,
				},
			},
			Context: p.Cluster.Context,
			DaemonSet: k8s.DaemonSet{
				ApiVersion: "apps/v1",
				Kind:       "DaemonSet",
				Metadata: k8s.Metadata{
					Name:      "prometheus-node-exporter",
					Namespace: "kube-system",
					Labels: k8s.Labels{
						"app": "prometheus-node-exporter",
					},
				},
				Spec: k8s.DaemonSetSpec{
					Selector: k8s.Selector{
						MatchLabels: k8s.Labels{
							"app": "prometheus-node-exporter",
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
								"app": "prometheus-node-exporter",
							},
						},
						Spec: k8s.PodSpec{
							HostPid: true,
							Containers: []k8s.Container{
								{
									Args: []k8s.Arg{
										"-collector.procfs=/host/proc",
										"-collector.sysfs=/host/sys",
										"-collector.filesystem.ignored-mount-points='^/(sys|proc|dev|host|etc)($|/)'",
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
									SecurityContext: k8s.SecurityContext{
										Privileged: true,
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
# Alert for any instance that is unreachable for >5 minutes.
ALERT InstanceDown
	IF up == 0
	FOR 5m
	LABELS {
		severity = "critical",
	}
	ANNOTATIONS {
		summary = "Instance {{ $labels.instance }} down",
		description = "{{ $labels.instance }} of job {{ $labels.job }} has been down for more than 5 minutes.",
	}

# Alert disk space below 20% free
ALERT DiskOutOfSpace
	IF (max(diskstatus_bytesfree / diskstatus_bytestotal) by (app)) < 0.2
	FOR 1m
	LABELS {
		severity = "warning",
	}
	ANNOTATIONS {
		summary = "Disk of {{ $labels.app }} space",
		description = "{{ $labels.app }} is below 20% free space for more than 1 minutes.",
	}

# Alert disk space below 10% free
ALERT DiskOutOfSpace
	IF (max(diskstatus_bytesfree / diskstatus_bytestotal) by (app)) < 0.1
	FOR 1m
	LABELS {
		severity = "critical",
	}
	ANNOTATIONS {
		summary = "Disk of {{ $labels.app }} space",
		description = "{{ $labels.app }} is below 10% free space for more than 1 minutes.",
	}

# Alert disk space below 20% free
ALERT DiskOutOfInodes
	IF (max(diskstatus_inodesfree / diskstatus_inodestotal) by (app)) < 0.2
	FOR 1m
	LABELS {
		severity = "warning",
	}
	ANNOTATIONS {
		summary = "Disk of {{ $labels.app }} inodes",
		description = "{{ $labels.app }} is below 20% free inodes for more than 1 minutes.",
	}

# Alert disk space below 10% free
ALERT DiskOutOfInodes
	IF (max(diskstatus_inodesfree / diskstatus_inodestotal) by (app)) < 0.1
	FOR 1m
	LABELS {
		severity = "critical",
	}
	ANNOTATIONS {
		summary = "Disk of {{ $labels.app }} inodes",
		description = "{{ $labels.app }} is below 10% free inodes for more than 1 minutes.",
	}
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

# Rule files specifies a list of globs. Rules and alerts are read from
# all matching files.
rule_files:
- 'alert.rules'

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

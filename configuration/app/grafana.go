package app

import (
	"bytes"
	"context"
	"text/template"

	"github.com/pkg/errors"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Grafana struct {
	Cluster      cluster.Cluster
	Domain       k8s.IngressHost
	LdapUsername deployer.SecretValue
	LdapPassword deployer.SecretValue
	Requirements []world.Configuration
}

func (g *Grafana) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		g.Cluster,
		g.Domain,
		g.LdapPassword,
		g.LdapUsername,
	)
}

func (g *Grafana) Applier() (world.Applier, error) {
	return nil, nil
}

func (g *Grafana) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, g.Requirements...)
	result = append(result, g.grafana()...)
	return result
}

func (g *Grafana) grafana() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/grafana",
		Tag:        "5.2.4",
	}
	port := deployer.Port{
		Port:     3000,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   g.Cluster.Context,
			Namespace: "grafana",
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   g.Cluster.Context,
				Namespace: "grafana",
				Name:      "config",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "grafana.ini",
						Value: grafanaIni,
					},
					deployer.ConfigEntry{
						Key:       "ldap.toml",
						ValueFrom: g.generateLdapToml,
					},
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   g.Cluster.Context,
				Namespace: "grafana",
				Name:      "datasources",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "all.yml",
						Value: datasourceYaml,
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   g.Cluster.Context,
			Namespace: "grafana",
			Name:      "grafana",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "grafana",
					Image: image,
					Requirement: &build.Grafana{
						Image: image,
					},
					Ports: []deployer.Port{port},
					Env: []k8s.Env{
						{
							Name:  "GF_PATHS_CONFIG",
							Value: "/config/grafana.ini",
						},
						{
							Name:  "GF_PATHS_DATA",
							Value: "/var/lib/grafana",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "config",
							Path: "/config",
						},
						{
							Name: "datasources",
							Path: "/etc/grafana/provisioning/datasources",
						},
						{
							Name: "data",
							Path: "/var/lib/grafana",
						},
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "25Mi",
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 10,
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
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "config",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "grafana.ini",
								Path: "grafana.ini",
							},
							{
								Key:  "ldap.toml",
								Path: "ldap.toml",
							},
						},
					},
				},
				{
					Name: "datasources",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "datasources",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "all.yml",
								Path: "all.yml",
							},
						},
					},
				},
				{
					Name: "data",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/grafana",
						Server: g.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   g.Cluster.Context,
			Namespace: "grafana",
			Name:      "grafana",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   g.Cluster.Context,
			Namespace: "grafana",
			Name:      "grafana",
			Port:      port.Name,
			Domains:   k8s.IngressHosts{g.Domain},
		},
	}
}

func (g *Grafana) generateLdapToml(ctx context.Context) (string, error) {
	t, err := template.New("template").Parse(ldapToml)
	if err != nil {
		return "", errors.Wrap(err, "parse ldap toml failed")
	}
	b := &bytes.Buffer{}
	user, err := g.LdapUsername.Value()
	if err != nil {
		return "", errors.Wrap(err, "get ldap username failed")
	}
	password, err := g.LdapPassword.Value()
	if err != nil {
		return "", errors.Wrap(err, "get ldap password failed")
	}
	err = t.Execute(b, struct {
		LdapHost              string
		LdapPort              int
		LdapBindUsename       string
		LdapBindPassword      string
		LdapUserBaseDn        string
		LdapUserSearchFilter  string
		LdapGroupBaseDn       string
		LdapGroupSearchFilter string
		LdapAdminGroupDn      string
		LdapEditorGroupDn     string
		LdapViewerGroupDn     string
	}{
		LdapHost:              "ldap.ldap.svc.cluster.local",
		LdapPort:              389,
		LdapBindUsename:       string(user),
		LdapBindPassword:      string(password),
		LdapUserBaseDn:        "ou=users,dc=benjamin-borbe,dc=de",
		LdapUserSearchFilter:  "(uid=%s)",
		LdapGroupBaseDn:       "ou=groups,dc=benjamin-borbe,dc=de",
		LdapGroupSearchFilter: "(member=uid=%s,ou=users,dc=benjamin-borbe,dc=de)",
		LdapAdminGroupDn:      "Admins",
		LdapEditorGroupDn:     "Admins",
		LdapViewerGroupDn:     "Employees",
	})
	if err != nil {
		return "", errors.Wrap(err, "parse ldapToml failed")
	}
	return b.String(), nil
}

const ldapToml = `
[[servers]]
host = "{{ .LdapHost }}"
port = {{ .LdapPort }}
use_ssl = false
start_tls = false
ssl_skip_verify = false

bind_dn = '{{ .LdapBindUsename }}'
bind_password = '{{ .LdapBindPassword }}'
search_filter = "{{ .LdapUserSearchFilter }}"
search_base_dns = ["{{ .LdapUserBaseDn }}"]

group_search_base_dns = ["{{ .LdapGroupBaseDn }}"]
group_search_filter = "{{ .LdapGroupSearchFilter }}"

[servers.attributes]
name = "givenName"
surname = "sn"
username = "uid"
member_of = "cn"
email =  "mail"

[[servers.group_mappings]]
group_dn = "{{ .LdapAdminGroupDn }}"
org_role = "Admin"

[[servers.group_mappings]]
group_dn = "{{ .LdapAdminGroupDn }}"
org_role = "Admin"

[[servers.group_mappings]]
group_dn = "{{ .LdapEditorGroupDn }}"
org_role = "Editor"

[[servers.group_mappings]]
group_dn = "{{ .LdapViewerGroupDn }}"
org_role = "Viewer"
`

const grafanaIni = `
[users]
allow_sign_up = false

[auth.ldap]
enabled = true
config_file = /config/ldap.toml
`

const datasourceYaml = `
apiVersion: 1

deleteDatasources:
- name: Prometheus
  orgId: 1

datasources:
- name: Prometheus
  type: prometheus
  access: proxy
  orgId: 1
  url: http://prometheus.prometheus.svc.cluster.local:9090
  isDefault: true
`
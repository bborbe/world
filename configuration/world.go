// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"context"

	"github.com/bborbe/world/configuration/app"
	"github.com/bborbe/world/configuration/cluster"
	service "github.com/bborbe/world/configuration/serivce"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/hetzner"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/secret"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type AppName string

type ClusterName string

type World struct {
	App              AppName
	Cluster          ClusterName
	TeamvaultSecrets *secret.Teamvault
}

func (w *World) Children() []world.Configuration {
	var result []world.Configuration
	for clusterName, configurations := range w.configurations() {
		if clusterName != w.Cluster && w.Cluster != "" {
			continue
		}
		for appName, configuration := range configurations {
			if appName != w.App && w.App != "" {
				continue
			}
			result = append(result, configuration)
		}
	}
	return result
}

func (w *World) Applier() (world.Applier, error) {
	return nil, nil
}

func (w *World) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.TeamvaultSecrets,
	)
}

func (w *World) configurations() map[ClusterName]map[AppName]world.Configuration {
	return map[ClusterName]map[AppName]world.Configuration{
		"cluster-1": w.cluster1(),
		"cluster-2": w.cluster2(),
		"netcup":    w.netcup(),
		"sun":       w.sun(),
		"fire":      w.fire(),
		"nuke":      w.nuke(),
		"hetzner-1": w.hetzner1(),
	}
}

func (w *World) hetzner1() map[AppName]world.Configuration {
	context := k8s.Context("hetzner-1")
	apiKey := w.TeamvaultSecrets.Password("kLolmq")
	ip := &hetzner.IP{
		ApiKey: apiKey,
		Name:   context,
	}
	return map[AppName]world.Configuration{
		"cluster": &cluster.Hetzner{
			Context:    context,
			ApiKey:     apiKey,
			IP:         ip,
			ServerType: "cx11",
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: context,
		},
		"calico": &service.Calico{
			Context:   context,
			ClusterIP: ip,
		},
		"dns": &app.CoreDns{
			Context: context,
		},
		"kubeless": world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: dns.Host("kubeless.hetzner-1.benjamin-borbe.de"),
						IP:   ip,
					},
				},
			},
		),
		"hostpath": &app.HostPathProvisioner{
			Context:             context,
			HostPath:            "/data",
			DefaultStorageClass: true,
		},
		"traefik": &app.Traefik{
			Context: context,
			Domains: k8s.IngressHosts{
				"traefik.hetzner-1.benjamin-borbe.de",
			},
		},
		"ip": &app.Ip{
			Context: context,
			IP:      ip,
			Tag:     "1.1.0",
			Domain:  "ip.hetzner-1.benjamin-borbe.de",
		},
		"openfaas": &app.OpenFaas{
			Context: context,
			IP:      ip,
			Domain:  "openfaas.hetzner-1.benjamin-borbe.de",
		},
	}
}

func (w *World) cluster1() map[AppName]world.Configuration {
	context := k8s.Context("gke_smedia-kubernetes_europe-west1-d_cluster-1")
	nfsServer := k8s.PodNfsServer("10.15.48.11")
	return map[AppName]world.Configuration{
		"kafka": &app.Kafka{
			AccessMode:        "ReadWriteOnce",
			Context:           context,
			DisableConnect:    true,
			DisableRest:       true,
			KafkaReplicas:     1,
			KafkaStorage:      "20Gi",
			StorageClass:      "standard",
			ZookeeperReplicas: 1,
			ZookeeperStorage:  "5Gi",
			Version:           "5.0.1",
		},
		"kafka-sample": &app.KafkaSample{
			Context: context,
			Domain:  "kafka-sample.lab.seibert-media.net",
		},
		"kafka-status": &app.KafkaStatus{
			Context:  context,
			Replicas: 1,
			Domain:   "kafka-status.lab.seibert-media.net",
		},
		"debug": &app.Debug{
			Context: context,
			Domain:  "debug.lab.seibert-media.net",
		},
		"metabase": &app.Metabase{
			Context:          context,
			NfsServer:        nfsServer,
			Domain:           "metabase.lab.seibert-media.net",
			DatabasePassword: w.TeamvaultSecrets.Password("dwkWAw"),
		},
	}
}

func (w *World) cluster2() map[AppName]world.Configuration {
	context := k8s.Context("gke_smedia-kubernetes_europe-west1-d_cluster-2")
	return map[AppName]world.Configuration{
		"kafka": &app.Kafka{
			AccessMode:        "ReadWriteOnce",
			Context:           context,
			DisableConnect:    true,
			DisableRest:       true,
			KafkaReplicas:     1,
			KafkaStorage:      "20Gi",
			StorageClass:      "standard",
			ZookeeperReplicas: 1,
			ZookeeperStorage:  "5Gi",
			Version:           "5.0.1",
		},
	}
}

func (w *World) fire() map[AppName]world.Configuration {
	context := k8s.Context("fire")
	nfsServer := k8s.PodNfsServer("192.168.178.3")
	return map[AppName]world.Configuration{
		"cluster": &cluster.Fire{
			Context:   context,
			ClusterIP: dns.IPStatic("192.168.178.3"),
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: context,
		},
		"calico": &service.Calico{
			Context:   context,
			ClusterIP: dns.IPStatic("192.168.178.3"),
		},
		"dns": &app.CoreDns{
			Context: context,
		},
		"traefik": &app.Traefik{
			Context:   context,
			NfsServer: nfsServer,
			Domains: k8s.IngressHosts{
				"traefik.fire.hm.benjamin-borbe.de",
			},
		},
		"backup": &app.BackupClient{
			Context:   context,
			NfsServer: nfsServer,
			Domains: k8s.IngressHosts{
				"backup.fire.hm.benjamin-borbe.de",
			},
			GitSyncPassword: w.TeamvaultSecrets.Password("YLb4wV"),
			BackupSshKey:    w.TeamvaultSecrets.File("8q1bJw"),
			GitRepoUrl:      "https://bborbereadonly@bitbucket.org/bborbe/backup_config_fire.git",
		},
	}
}

func (w *World) nuke() map[AppName]world.Configuration {
	context := k8s.Context("nuke")
	nfsServer := k8s.PodNfsServer("192.168.178.5")
	return map[AppName]world.Configuration{
		"cluster": &cluster.Nuke{
			Context:   context,
			ClusterIP: dns.IPStatic("192.168.178.5"),
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: context,
		},
		"calico": &service.Calico{
			Context:   context,
			ClusterIP: dns.IPStatic("192.168.178.5"),
		},
		"dns": &app.CoreDns{
			Context: context,
		},
		"traefik": &app.Traefik{
			Context:   context,
			NfsServer: nfsServer,
			Domains: k8s.IngressHosts{
				"traefik.nuke.hm.benjamin-borbe.de",
			},
		},
		"backup": &app.BackupClient{
			NfsServer: nfsServer,
			Context:   context,
			Domains: k8s.IngressHosts{
				"backup.nuke.hm.benjamin-borbe.de",
			},
			GitSyncPassword: w.TeamvaultSecrets.Password("YLb4wV"),
			BackupSshKey:    w.TeamvaultSecrets.File("8q1bJw"),
			GitRepoUrl:      "https://bborbereadonly@bitbucket.org/bborbe/backup_config_nuke.git",
		},
	}
}

func (w *World) sun() map[AppName]world.Configuration {
	context := k8s.Context("sun")
	nfsServer := k8s.PodNfsServer("192.168.2.3")
	return map[AppName]world.Configuration{
		"cluster": &cluster.Sun{
			Context:   context,
			ClusterIP: dns.IPStatic("192.168.2.3"),
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: context,
		},
		"calico": &service.Calico{
			Context:   context,
			ClusterIP: dns.IPStatic("192.168.2.3"),
		},
		"dns": &app.CoreDns{
			Context: context,
		},
		"monitoring": &app.Monitoring{
			Context:         context,
			GitSyncPassword: w.TeamvaultSecrets.Password("YLb4wV"),
			SmtpPassword:    w.TeamvaultSecrets.Password("QL3VQO"),
			Configs: []app.MonitoringConfig{
				{
					Name:       "nc",
					Subject:    "Monitoring Result: Netcup",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_nc.git",
				},
				{
					Name:       "pn-intern",
					Subject:    "Monitoring Result: PN-Intern",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_pn_intern.git",
				},
				{
					Name:       "hm",
					Subject:    "Monitoring Result: HM",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_hm.git",
				},
			},
		},
		"traefik": &app.Traefik{
			Context:   context,
			NfsServer: nfsServer,
			Domains: k8s.IngressHosts{
				"traefik.sun.pn.benjamin-borbe.de",
			},
		},
		"backup": &app.BackupClient{
			Context:   context,
			NfsServer: nfsServer,
			Domains: k8s.IngressHosts{
				"backup.sun.pn.benjamin-borbe.de",
			},
			GitSyncPassword: w.TeamvaultSecrets.Password("YLb4wV"),
			BackupSshKey:    w.TeamvaultSecrets.File("8q1bJw"),
			GitRepoUrl:      "https://bborbereadonly@bitbucket.org/bborbe/backup_config_sun.git",
		},
	}
}

func (w *World) netcup() map[AppName]world.Configuration {
	context := k8s.Context("netcup")
	nfsServer := k8s.PodNfsServer("185.170.112.48")
	ip := dns.IPStatic("185.170.112.48")
	return map[AppName]world.Configuration{
		"cluster": &cluster.Netcup{
			Context: context,
			IP:      ip,
		},
		"cluster-admin": &service.ClusterAdmin{
			Context: context,
		},
		"calico": &service.Calico{
			Context:   context,
			ClusterIP: ip,
		},
		"debug": &app.Debug{
			Context: context,
			Domain:  "debug.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: dns.Host("debug.benjamin-borbe.de"),
								IP:   ip,
							},
						},
					},
				),
			},
		},
		"metabase": &app.Metabase{
			Context:          context,
			NfsServer:        nfsServer,
			NfsPrefix:        "/data",
			Domain:           "metabase.benjamin-borbe.de",
			DatabasePassword: w.TeamvaultSecrets.Password("dwkWAw"),
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: dns.Host("metabase.benjamin-borbe.de"),
								IP:   ip,
							},
						},
					},
				),
			},
		},
		"grafana": &app.Grafana{
			Context:      context,
			NfsServer:    nfsServer,
			Domain:       "grafana.benjamin-borbe.de",
			LdapUsername: w.TeamvaultSecrets.Username("MOPMLG"),
			LdapPassword: w.TeamvaultSecrets.Password("MOPMLG"),
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: dns.Host("grafana.benjamin-borbe.de"),
								IP:   ip,
							},
						},
					},
				),
			},
		},
		//"nfs": &app.NfsProvisioner{
		//	Context:  context,
		//	HostPath: "/data/nfs-provisioner",
		//},
		"hostpath": &app.HostPathProvisioner{
			Context:             context,
			HostPath:            "/data/hostpath-provisioner",
			DefaultStorageClass: true,
		},
		"kafka": &app.Kafka{
			AccessMode:        "ReadWriteMany",
			Context:           context,
			DisableConnect:    true,
			DisableRest:       true,
			KafkaReplicas:     1,
			KafkaStorage:      "5Gi",
			StorageClass:      "hostpath",
			ZookeeperReplicas: 1,
			ZookeeperStorage:  "5Gi",
			Version:           "5.1.0",
		},
		"kafka-status": &app.KafkaStatus{
			Context:  context,
			Replicas: 1,
			Domain:   "kafka-status.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: dns.Host("kafka-status.benjamin-borbe.de"),
								IP:   ip,
							},
						},
					},
				),
			},
		},
		"kafka-latest-versions": &app.KafkaLatestVersions{
			Context:      context,
			Replicas:     2,
			AccessMode:   "ReadWriteMany",
			StorageClass: "hostpath",
			Domain:       "versions.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: dns.Host("versions.benjamin-borbe.de"),
								IP:   ip,
							},
						},
					},
				),
			},
		},
		"kafka-update-available": &app.KafkaUpdateAvailable{
			Context:      context,
			Replicas:     2,
			AccessMode:   "ReadWriteMany",
			StorageClass: "hostpath",
			Domain:       "updates.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: dns.Host("updates.benjamin-borbe.de"),
								IP:   ip,
							},
						},
					},
				),
			},
		},
		"kafka-k8s-version-collector": &app.KafkaK8sVersionCollector{
			Context: context,
		},
		"kafka-regex-version-collector": &app.KafkaRegexVersionCollector{
			Context:     context,
			Application: "Golang",
			Url:         "https://golang.org/dl/",
			Regex:       `https://dl.google.com/go/go([\d\.]+)\.src\.tar\.gz`,
		},
		"kafka-atlassian-version-collector": &app.KafkaAtlassianVersionCollector{
			Context: context,
		},
		"kafka-installed-version-collector": &app.KafkaInstalledVersionCollector{
			Context: context,
			Apps: []struct {
				Name  string
				Regex string
				Url   string
			}{
				{
					Name:  "Confluence",
					Regex: `<meta\s+name="ajs-version-number"\s+content="([^"]+)">`,
					Url:   "https://confluence.benjamin-borbe.de",
				},
				{
					Name:  "Jira",
					Regex: `<meta\s+name="ajs-version-number"\s+content="([^"]+)">`,
					Url:   "https://jira.benjamin-borbe.de",
				},
			},
		},
		"kafka-sample": &app.KafkaSample{
			Context: context,
			Domain:  "kafka-sample.benjamin-borbe.de",
			Requirements: []world.Configuration{
				world.NewConfiguraionBuilder().WithApplier(
					&dns.Server{
						Host:    "ns.rocketsource.de",
						KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
						List: []dns.Entry{
							{
								Host: dns.Host("kafka-sample.benjamin-borbe.de"),
								IP:   ip,
							},
						},
					},
				),
			},
		},
		//"erpnext": &app.ErpNext{
		//	Cluster:              c,
		//	Domain:               "erpnext.benjamin-borbe.de",
		//	DatabaseRootPassword: w.TeamvaultSecrets.Password("dqDzmO"),
		//	DatabaseName:         w.TeamvaultSecrets.Username("MOPGMw"),
		//	DatabasePassword:     w.TeamvaultSecrets.Password("MOPGMw"),
		//	AdminPassword:        w.TeamvaultSecrets.Password("AwJndw"),
		//},
		"dns": &app.CoreDns{
			Context: context,
		},
		"traefik": &app.Traefik{
			Context:   context,
			NfsServer: nfsServer,
			Domains: k8s.IngressHosts{
				"traefik.benjamin-borbe.de",
			},
			SSL: true,
		},
		"prometheus": &app.Prometheus{
			Context:            context,
			NfsServer:          nfsServer,
			PrometheusDomain:   "prometheus.benjamin-borbe.de",
			AlertmanagerDomain: "prometheus-alertmanager.benjamin-borbe.de",
			Secret:             w.TeamvaultSecrets.Password("aqMr6w"),
			LdapUsername:       w.TeamvaultSecrets.Username("MOPMLG"),
			LdapPassword:       w.TeamvaultSecrets.Password("MOPMLG"),
		},
		"ldap": &app.Ldap{
			Context:      context,
			NfsServer:    nfsServer,
			Tag:          "1.3.0",
			LdapPassword: w.TeamvaultSecrets.Password("MOPMLG"),
		},
		"teamvault": &app.Teamvault{
			Context:          context,
			NfsServer:        nfsServer,
			Domain:           "teamvault.benjamin-borbe.de",
			DatabasePassword: w.TeamvaultSecrets.Password("VO0W5w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("3OlNaq"),
			SmtpPassword:     w.TeamvaultSecrets.Password("3OlNaq"),
			LdapPassword:     w.TeamvaultSecrets.Password("MOPMLG"),
			SecretKey:        w.TeamvaultSecrets.Password("NqA68w"),
			FernetKey:        w.TeamvaultSecrets.Password("5wYZ2O"),
			Salt:             w.TeamvaultSecrets.Password("Rwg74w"),
		},
		"monitoring": &app.Monitoring{
			Context:         context,
			GitSyncPassword: w.TeamvaultSecrets.Password("YLb4wV"),
			SmtpPassword:    w.TeamvaultSecrets.Password("QL3VQO"),
			Configs: []app.MonitoringConfig{
				{
					Name:       "pn",
					Subject:    "Monitoring Result: PN",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_pn.git",
				},
				{
					Name:       "hm",
					Subject:    "Monitoring Result: HM",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_hm.git",
				},
				{
					Name:       "work",
					Subject:    "Monitoring Result: Work",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_work.git",
				},
			},
		},
		//"proxy": &app.Proxy{
		//	Cluster:  netcup,
		//	Password: w.TeamvaultSecrets.Htpasswd("zL89oq"),
		//},
		"confluence": &app.Confluence{
			Context:          context,
			NfsServer:        nfsServer,
			Domain:           "confluence.benjamin-borbe.de",
			Version:          "6.13.2",
			DatabasePassword: w.TeamvaultSecrets.Password("3OlaLn"),
			SmtpUsername:     w.TeamvaultSecrets.Username("nOeNjL"),
			SmtpPassword:     w.TeamvaultSecrets.Password("nOeNjL"),
		},
		"jira": &app.Jira{
			Context:          context,
			NfsServer:        nfsServer,
			Domain:           "jira.benjamin-borbe.de",
			Version:          "7.13.1",
			DatabasePassword: w.TeamvaultSecrets.Password("eOB12w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("MwmE0w"),
			SmtpPassword:     w.TeamvaultSecrets.Password("MwmE0w"),
		},
		"backup": &app.BackupServer{
			Context:   context,
			NfsServer: nfsServer,
		},
		"poste": &app.Poste{
			Context:      context,
			NfsServer:    nfsServer,
			PosteVersion: "2.1.0", // https://hub.docker.com/r/analogic/poste.io/tags
			Domains: k8s.IngressHosts{
				"mail.benjamin-borbe.de",
			},
		},
		"maven": &app.Maven{
			Context:   context,
			NfsServer: nfsServer,
			Domains: k8s.IngressHosts{
				"maven.benjamin-borbe.de",
			},
			MavenRepoVersion: "1.0.0",
		},
		"portfolio": &app.Portfolio{
			Context: context,
			Domains: k8s.IngressHosts{
				"benjamin-borbe.de",
				"www.benjamin-borbe.de",
				"benjaminborbe.de",
				"www.benjaminborbe.de",
			},
			OverlayServerVersion: "1.0.0",
			GitSyncPassword:      w.TeamvaultSecrets.Password("YLb4wV"),
		},
		"webdav": &app.Webdav{
			Context:   context,
			NfsServer: nfsServer,
			Domains: k8s.IngressHosts{
				"webdav.benjamin-borbe.de",
			},
			Tag:      "1.0.1",
			Password: w.TeamvaultSecrets.Password("VOzvAO"),
		},
		"bind": &app.Bind{
			Context:   context,
			NfsServer: nfsServer,
			Tag:       "1.0.1",
		},
		"download": &app.Download{
			Context:   context,
			NfsServer: nfsServer,
			Domains: k8s.IngressHosts{
				"dl.benjamin-borbe.de",
			},
		},
		"mumble": &app.Mumble{
			Context: context,
			Tag:     "1.1.1",
		},
		"ip": &app.Ip{
			Context: context,
			IP:      ip,
			Tag:     "1.1.0",
			Domain:  "ip.benjamin-borbe.de",
		},
		"password": &app.Password{
			Context: context,
			Tag:     "1.1.0",
			Domains: k8s.IngressHosts{
				"password.benjamin-borbe.de",
			},
		},
		"now": &app.Now{
			Context: context,
			Tag:     "1.3.0",
			Domains: k8s.IngressHosts{
				"now.benjamin-borbe.de",
			},
		},
		"helloworld": &app.HelloWorld{
			Context: context,
			Tag:     "1.0.1",
			Domains: k8s.IngressHosts{
				"rocketsource.de",
				"www.rocketsource.de",
				"rocketnews.de",
				"www.rocketnews.de",
			},
		},
		"slideshow": &app.Slideshow{
			Context: context,
			Domains: k8s.IngressHosts{
				"slideshow.benjamin-borbe.de",
			},
		},
		"kickstart": &app.Kickstart{
			Context: context,
			Domains: k8s.IngressHosts{
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			},
		},
	}
}

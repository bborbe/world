package configuration

import (
	"context"

	"github.com/bborbe/world/configuration/app"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/server"
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
		"netcup": w.netcup(),
		"sun":    w.sun(),
		"fire":   w.fire(),
		"nuke":   w.nuke(),
	}
}

func (w *World) fire() map[AppName]world.Configuration {
	fire := cluster.Cluster{
		Context:   "fire",
		NfsServer: "172.16.22.1",
	}
	return map[AppName]world.Configuration{
		"server": &server.Fire{
			Context:   fire.Context,
			ClusterIP: "192.168.178.3",
		},
		"dns": &app.KubeDns{
			Cluster: fire,
		},
		"traefik": &app.Traefik{
			Cluster: fire,
			Domains: k8s.IngressHosts{
				"traefik.fire.hm.benjamin-borbe.de",
			},
		},
		"backup": &app.BackupClient{
			Cluster: fire,
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
	nuke := cluster.Cluster{
		Context:   "nuke",
		NfsServer: "172.16.24.1",
	}
	return map[AppName]world.Configuration{
		"server": &server.Nuke{
			Context:   nuke.Context,
			ClusterIP: "192.168.178.5",
		},
		"dns": &app.KubeDns{
			Cluster: nuke,
		},
		"traefik": &app.Traefik{
			Cluster: nuke,
			Domains: k8s.IngressHosts{
				"traefik.nuke.hm.benjamin-borbe.de",
			},
		},
		"backup": &app.BackupClient{
			Cluster: nuke,
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
	sun := cluster.Cluster{
		Context:   "sun",
		NfsServer: "172.16.72.1",
	}
	return map[AppName]world.Configuration{
		"server": &server.Sun{
			Context:   sun.Context,
			ClusterIP: "192.168.2.3",
		},
		"dns": &app.KubeDns{
			Cluster: sun,
		},
		"monitoring": &app.Monitoring{
			Cluster:         sun,
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
			Cluster: sun,
			Domains: k8s.IngressHosts{
				"traefik.sun.pn.benjamin-borbe.de",
			},
		},
		"backup": &app.BackupClient{
			Cluster: sun,
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
	netcup := cluster.Cluster{
		Context:   "netcup",
		NfsServer: "185.170.112.48",
	}
	return map[AppName]world.Configuration{
		"server": &server.Netcup{},
		"kafka": &app.Kafka{
			Cluster: netcup,
		},
		"erpnext": &app.ErpNext{
			Cluster:              netcup,
			Domain:               "erpnext.benjamin-borbe.de",
			DatabaseRootPassword: w.TeamvaultSecrets.Password("dqDzmO"),
			DatabaseName:         w.TeamvaultSecrets.Username("MOPGMw"),
			DatabasePassword:     w.TeamvaultSecrets.Password("MOPGMw"),
			AdminPassword:        w.TeamvaultSecrets.Password("AwJndw"),
		},
		"dns": &app.CoreDns{
			Cluster: netcup,
		},
		"traefik": &app.Traefik{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"traefik.benjamin-borbe.de",
			},
			SSL: true,
		},
		"prometheus": &app.Prometheus{
			Cluster:            netcup,
			PrometheusDomain:   "prometheus.benjamin-borbe.de",
			AlertmanagerDomain: "prometheus-alertmanager.benjamin-borbe.de",
			Secret:             w.TeamvaultSecrets.Password("aqMr6w"),
			LdapUsername:       w.TeamvaultSecrets.Username("MOPMLG"),
			LdapPassword:       w.TeamvaultSecrets.Password("MOPMLG"),
		},
		"ldap": &app.Ldap{
			Cluster:      netcup,
			Tag:          "1.1.0",
			LdapPassword: w.TeamvaultSecrets.Password("MOPMLG"),
		},
		"teamvault": &app.Teamvault{
			Cluster:          netcup,
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
			Cluster:         netcup,
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
			},
		},
		//"proxy": &app.Proxy{
		//	Cluster:  netcup,
		//	Password: w.TeamvaultSecrets.Htpasswd("zL89oq"),
		//},
		"confluence": &app.Confluence{
			Cluster:          netcup,
			Domain:           "confluence.benjamin-borbe.de",
			Version:          "6.10.2",
			DatabasePassword: w.TeamvaultSecrets.Password("3OlaLn"),
			SmtpUsername:     w.TeamvaultSecrets.Username("nOeNjL"),
			SmtpPassword:     w.TeamvaultSecrets.Password("nOeNjL"),
		},
		"jira": &app.Jira{
			Cluster:          netcup,
			Domain:           "jira.benjamin-borbe.de",
			Version:          "7.11.2",
			DatabasePassword: w.TeamvaultSecrets.Password("eOB12w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("MwmE0w"),
			SmtpPassword:     w.TeamvaultSecrets.Password("MwmE0w"),
		},
		"backup": &app.BackupServer{
			Cluster: netcup,
		},
		"poste": &app.Poste{
			Cluster:      netcup,
			PosteVersion: "1.0.7",
			Domains: k8s.IngressHosts{
				"mail.benjamin-borbe.de",
			},
		},
		"maven": &app.Maven{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"maven.benjamin-borbe.de",
			},
			MavenRepoVersion: "1.0.0",
		},
		"portfolio": &app.Portfolio{
			Cluster: netcup,
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
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"webdav.benjamin-borbe.de",
			},
			Tag:      "1.0.1",
			Password: w.TeamvaultSecrets.Password("VOzvAO"),
		},
		"bind": &app.Bind{
			Cluster: netcup,
			Tag:     "1.0.1",
		},
		"download": &app.Download{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"dl.benjamin-borbe.de",
			},
		},
		"mumble": &app.Mumble{
			Cluster: netcup,
			Tag:     "1.1.1",
		},
		"ip": &app.Ip{
			Cluster: netcup,
			Tag:     "1.1.0",
			Domains: k8s.IngressHosts{
				"ip.benjamin-borbe.de",
			},
		},
		"password": &app.Password{
			Cluster: netcup,
			Tag:     "1.1.0",
			Domains: k8s.IngressHosts{
				"password.benjamin-borbe.de",
			},
		},
		"now": &app.Now{
			Cluster: netcup,
			Tag:     "1.1.0",
			Domains: k8s.IngressHosts{
				"now.benjamin-borbe.de",
			},
		},
		"helloworld": &app.HelloWorld{
			Cluster: netcup,
			Tag:     "1.0.1",
			Domains: k8s.IngressHosts{
				"rocketsource.de",
				"www.rocketsource.de",
				"rocketnews.de",
				"www.rocketnews.de",
			},
		},
		"slideshow": &app.Slideshow{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"slideshow.benjamin-borbe.de",
			},
		},
		"kickstart": &app.Kickstart{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			},
		},
	}

}

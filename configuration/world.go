package configuration

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/app"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/secret"
	"github.com/pkg/errors"
)

type World struct {
	Name             string
	TeamvaultSecrets *secret.Teamvault
}

func (w *World) Configuration() (world.Configuration, error) {
	if w.Name == "" {
		configuration := world.NewConfiguration()
		for _, c := range w.configurations() {
			configuration.AddChildConfiguration(c)
		}
		return configuration, nil
	}
	configuration, ok := w.configurations()[w.Name]
	if !ok {
		return nil, errors.New("Configurations not found")
	}
	return configuration, nil
}

func (w *World) configurations() map[string]world.Configuration {
	netcup := cluster.Cluster{
		Context:   "netcup",
		NfsServer: "185.170.112.48",
	}
	sun := cluster.Cluster{
		Context:   "sun",
		NfsServer: "172.16.72.1",
	}
	var gitSyncVersion docker.Tag = "1.3.0"
	return map[string]world.Configuration{
		"dns": &app.Dns{
			Cluster: netcup,
		},
		"traefik": &app.Traefik{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"traefik.benjamin-borbe.de",
			},
		},
		"prometheus": &app.Prometheus{
			Cluster: netcup,
		},
		"ldap": &app.Ldap{
			Cluster:    netcup,
			Tag:        "1.1.0",
			LdapSecret: w.TeamvaultSecrets.Password("MOPMLG"),
		},
		"teamvault": &app.Teamvault{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"teamvault.benjamin-borbe.de",
			},
			DatabasePassword: w.TeamvaultSecrets.Password("VO0W5w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("3OlNaq"),
			SmtpPassword:     w.TeamvaultSecrets.Password("3OlNaq"),
			LdapPassword:     w.TeamvaultSecrets.Password("MOPMLG"),
			SecretKey:        w.TeamvaultSecrets.Password("NqA68w"),
			FernetKey:        w.TeamvaultSecrets.Password("5wYZ2O"),
			Salt:             w.TeamvaultSecrets.Password("Rwg74w"),
		},
		"monitoring-nc": &app.Monitoring{
			Cluster:         netcup,
			GitSyncVersion:  gitSyncVersion,
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
		"monitoring-sun": &app.Monitoring{
			Cluster:         sun,
			GitSyncVersion:  gitSyncVersion,
			GitSyncPassword: w.TeamvaultSecrets.Password("YLb4wV"),
			SmtpPassword:    w.TeamvaultSecrets.Password("QL3VQO"),
			Configs: []app.MonitoringConfig{
				{
					Name:       "nc",
					Subject:    "Monitoring Result: Netcup",
					GitRepoUrl: "https://bborbereadonly@bitbucket.org/bborbe/monitoring_nc.git",
				},
			},
		},
		"proxy": &app.Proxy{
			Cluster:  netcup,
			Password: w.TeamvaultSecrets.Htpasswd("zL89oq"),
		},
		"confluence": &app.Confluence{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"confluence.benjamin-borbe.de",
			},
			Version:          "6.10.2",
			DatabasePassword: w.TeamvaultSecrets.Password("3OlaLn"),
			SmtpUsername:     w.TeamvaultSecrets.Username("nOeNjL"),
			SmtpPassword:     w.TeamvaultSecrets.Password("nOeNjL"),
		},
		"jira": &app.Jira{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"jira.benjamin-borbe.de",
			},
			Version:          "7.11.2",
			DatabasePassword: w.TeamvaultSecrets.Password("eOB12w"),
			SmtpUsername:     w.TeamvaultSecrets.Username("MwmE0w"),
			SmtpPassword:     w.TeamvaultSecrets.Password("MwmE0w"),
		},
		"backup": &app.Backup{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"backup.benjamin-borbe.de",
			},
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
			GitSyncVersion:       gitSyncVersion,
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
			Tag:     "1.0.2",
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
			Tag:     "1.0.1",
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
			GitSyncVersion: gitSyncVersion,
		},
		"kickstart": &app.Kickstart{
			Cluster: netcup,
			Domains: k8s.IngressHosts{
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			},
			GitSyncVersion: gitSyncVersion,
		},
	}

}

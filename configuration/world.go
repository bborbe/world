package configuration

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/app"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/secret"
	"github.com/pkg/errors"
)

type World struct {
	TeamvaultSecrets *secret.Teamvault
}

func (c *World) Applier() world.Applier {
	return nil
}

func (c *World) Validate(ctx context.Context) error {
	if c.TeamvaultSecrets == nil {
		return errors.New("Teamvault missing")
	}
	return nil
}

func (c *World) Children() []world.Configuration {
	netcup := cluster.Cluster{
		Context:   "netcup",
		NfsServer: "185.170.112.48",
	}
	var gitSyncVersion docker.Tag = "1.3.0"
	return []world.Configuration{
		//&app.Dns{
		//	Cluster: netcup,
		//},
		&app.Ldap{
			Cluster:    netcup,
			Tag:        "1.1.0",
			LdapSecret: c.TeamvaultSecrets.Password("MOPMLG"),
		},
		&app.Teamvault{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"teamvault.benjamin-borbe.de",
			},
			DatabasePassword: c.TeamvaultSecrets.Password("VO0W5w"),
			SmtpUsername:     c.TeamvaultSecrets.Username("3OlNaq"),
			SmtpPassword:     c.TeamvaultSecrets.Password("3OlNaq"),
			LdapPassword:     c.TeamvaultSecrets.Password("MOPMLG"),
			SecretKey:        c.TeamvaultSecrets.Password("NqA68w"),
			FernetKey:        c.TeamvaultSecrets.Password("5wYZ2O"),
			Salt:             c.TeamvaultSecrets.Password("Rwg74w"),
		},
		&app.Traefik{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"traefik.benjamin-borbe.de",
			},
		},
		&app.Confluence{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"confluence.benjamin-borbe.de",
			},
			Version:          "6.10.2",
			DatabasePassword: c.TeamvaultSecrets.Password("3OlaLn"),
			SmtpUsername:     c.TeamvaultSecrets.Username("nOeNjL"),
			SmtpPassword:     c.TeamvaultSecrets.Password("nOeNjL"),
		},
		&app.Jira{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"jira.benjamin-borbe.de",
			},
			Version:          "7.11.2",
			DatabasePassword: c.TeamvaultSecrets.Password("eOB12w"),
			SmtpUsername:     c.TeamvaultSecrets.Username("MwmE0w"),
			SmtpPassword:     c.TeamvaultSecrets.Password("MwmE0w"),
		},
		&app.Backup{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"backup.benjamin-borbe.de",
			},
		},
		&app.Poste{
			Cluster:      netcup,
			PosteVersion: "1.0.7",
			Domains: []k8s.IngressHost{
				"mail.benjamin-borbe.de",
			},
		},
		&app.Maven{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"maven.benjamin-borbe.de",
			},
			MavenRepoVersion: "1.0.0",
		},
		&app.Monitoring{
			Cluster: netcup,
		},
		&app.Portfolio{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"benjamin-borbe.de",
				"www.benjamin-borbe.de",
				"benjaminborbe.de",
				"www.benjaminborbe.de",
			},
			OverlayServerVersion: "1.0.0",
			GitSyncVersion:       gitSyncVersion,
			GitSyncPassword:      c.TeamvaultSecrets.Password("YLb4wV"),
		},
		&app.Prometheus{
			Cluster: netcup,
		},
		&app.Proxy{
			Cluster: netcup,
		},
		&app.Webdav{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"webdav.benjamin-borbe.de",
			},
			Tag:      "1.0.1",
			Password: c.TeamvaultSecrets.Password("VOzvAO"),
		},
		&app.Bind{
			Cluster: netcup,
			Tag:     "1.0.1",
		},
		&app.Download{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"dl.benjamin-borbe.de",
			},
		},
		&app.Mumble{
			Cluster: netcup,
			Tag:     "1.0.2",
		},
		&app.Ip{
			Cluster: netcup,
			Tag:     "1.1.0",
			Domains: []k8s.IngressHost{
				"ip.benjamin-borbe.de",
			},
		},
		&app.Password{
			Cluster: netcup,
			Tag:     "1.1.0",
			Domains: []k8s.IngressHost{
				"password.benjamin-borbe.de",
			},
		},
		&app.Now{
			Cluster: netcup,
			Tag:     "1.0.1",
			Domains: []k8s.IngressHost{
				"now.benjamin-borbe.de",
			},
		},
		&app.HelloWorld{
			Cluster: netcup,
			Tag:     "1.0.1",
			Domains: []k8s.IngressHost{
				"rocketsource.de",
				"www.rocketsource.de",
				"rocketnews.de",
				"www.rocketnews.de",
			},
		},
		&app.Slideshow{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"slideshow.benjamin-borbe.de",
			},
			GitSyncVersion: gitSyncVersion,
		},
		&app.Kickstart{
			Cluster: netcup,
			Domains: []k8s.IngressHost{
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			},
			GitSyncVersion: gitSyncVersion,
		},
	}
}

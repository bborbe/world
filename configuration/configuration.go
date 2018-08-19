package configuration

import (
	"context"

	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/app"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Configuration struct {
	TeamvaultConnector teamvault.Connector
}

func (c *Configuration) Applier() world.Applier {
	return nil
}

func (c *Configuration) Validate(ctx context.Context) error {
	if c.TeamvaultConnector == nil {
		return errors.New("configuration.teamvault-connector missing")
	}
	return nil
}

func (c *Configuration) Children() []world.Configuration {
	netcup := cluster.Cluster{
		Context:   "netcup",
		NfsServer: "185.170.112.48",
	}
	var gitSyncVersion docker.Tag = "1.3.0"
	return []world.Configuration{
		&app.Jira{
			Cluster: netcup,
			Domains: []deployer.Domain{
				"jira.benjamin-borbe.de",
			},
			Version:          "7.11.2",
			DatabasePassword: c.teamvaultPassword("eOB12w"),
			SmtpUsername:     c.teamvaultUsername("MwmE0w"),
			SmtpPassword:     c.teamvaultPassword("MwmE0w"),
		},
		&app.Confluence{
			Cluster: netcup,
			Domains: []deployer.Domain{
				"confluence.benjamin-borbe.de",
			},
			Version:          "6.8.1",
			DatabasePassword: c.teamvaultPassword("3OlaLn"),
			SmtpUsername:     c.teamvaultUsername("nOeNjL"),
			SmtpPassword:     c.teamvaultPassword("nOeNjL"),
		},
		&app.Ldap{
			Cluster:    netcup,
			Tag:        "1.1.0",
			LdapSecret: c.teamvaultPassword("MOPMLG"),
		},
		&app.Backup{
			Cluster: netcup,
			Domains: []deployer.Domain{
				"backup.benjamin-borbe.de",
			},
		},
		&app.Poste{
			Cluster:      netcup,
			PosteVersion: "1.0.7",
			Domains: []deployer.Domain{
				"mail.benjamin-borbe.de",
			},
		},
		&app.Maven{
			Cluster: netcup,
			Domains: []deployer.Domain{
				"maven.benjamin-borbe.de",
			},
			MavenRepoVersion: "1.0.0",
		},
		&app.Monitoring{
			Cluster: netcup,
		},
		&app.Portfolio{
			Cluster: netcup,
			Domains: []deployer.Domain{
				"benjamin-borbe.de",
				"www.benjamin-borbe.de",
				"benjaminborbe.de",
				"www.benjaminborbe.de",
			},
			OverlayServerVersion: "1.0.0",
			GitSyncVersion:       gitSyncVersion,
			GitSyncPassword:      c.teamvaultPassword("YLb4wV"),
		},
		&app.Prometheus{
			Cluster: netcup,
		},
		&app.Proxy{
			Cluster: netcup,
		},
		&app.Teamvault{
			Cluster: netcup,
		},
		&app.Traefik{
			Cluster: netcup,
		},
		&app.Webdav{
			Cluster: netcup,
			Domains: []deployer.Domain{
				"webdav.benjamin-borbe.de",
			},
			Tag:      "1.0.1",
			Password: c.teamvaultPassword("VOzvAO"),
		},
		&app.Bind{
			Cluster: netcup,
			Tag:     "1.0.1",
		},
		&app.Download{
			Cluster: netcup,
			Domains: []deployer.Domain{
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
			Domains: []deployer.Domain{
				"ip.benjamin-borbe.de",
			},
		},
		&app.Password{
			Cluster: netcup,
			Tag:     "1.1.0",
			Domains: []deployer.Domain{
				"password.benjamin-borbe.de",
			},
		},
		&app.Now{
			Cluster: netcup,
			Tag:     "1.0.1",
			Domains: []deployer.Domain{
				"now.benjamin-borbe.de",
			},
		},
		&app.HelloWorld{
			Cluster: netcup,
			Tag:     "1.0.1",
			Domains: []deployer.Domain{
				"rocketsource.de",
				"www.rocketsource.de",
				"rocketnews.de",
				"www.rocketnews.de",
			},
		},
		&app.Slideshow{
			Cluster: netcup,
			Domains: []deployer.Domain{
				"slideshow.benjamin-borbe.de",
			},
			GitSyncVersion: gitSyncVersion,
		},
		&app.Kickstart{
			Cluster: netcup,
			Domains: []deployer.Domain{
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			},
			GitSyncVersion: gitSyncVersion,
		},
		//&app.Dns{
		//	Cluster: netcup,
		//},
	}
}

func (c *Configuration) teamvaultPassword(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultPassword{
		TeamvaultConnector: c.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

func (c *Configuration) teamvaultUsername(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultUser{
		TeamvaultConnector: c.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

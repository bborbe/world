package configuration

import (
	"context"

	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/teamvault-utils/model"
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/app"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/pkg/docker"
	"github.com/pkg/errors"
)

type Configuration struct {
	TeamvaultConnector connector.Connector
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

func (c *Configuration) Childs() []world.Configuration {
	netcup := cluster.Cluster{
		Context:   "netcup",
		NfsServer: "185.170.112.48",
	}
	var gitSyncVersion docker.Tag = "1.3.0"
	return []world.Configuration{
		&app.Backup{
			Cluster: netcup,
		},
		&app.Dns{
			Cluster: netcup,
		},
		&app.Jenkins{
			Cluster: netcup,
		},
		&app.Jira{
			Cluster: netcup,
		},
		&app.Maven{
			Cluster: netcup,
			Domains: []world.Domain{
				"maven.benjamin-borbe.de",
			},
			MavenRepoVersion: "1.0.0",
		},
		&app.Monitoring{
			Cluster: netcup,
		},
		&app.Portfolio{
			Cluster: netcup,
			Domains: []world.Domain{
				"benjamin-borbe.de",
				"www.benjamin-borbe.de",
				"benjaminborbe.de",
				"www.benjaminborbe.de",
			},
			OverlayServerVersion: "1.0.0",
			GitSyncVersion:       gitSyncVersion,
			GitSyncPassword:      c.teamvaultPassword("YLb4wV"),
		},
		&app.Poste{
			Cluster: netcup,
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
			Domains: []world.Domain{
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
			Domains: []world.Domain{
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
			Domains: []world.Domain{
				"ip.benjamin-borbe.de",
			},
		},
		&app.Password{
			Cluster: netcup,
			Tag:     "1.1.0",
			Domains: []world.Domain{
				"password.benjamin-borbe.de",
			},
		},
		&app.Now{
			Cluster: netcup,
			Tag:     "1.0.1",
			Domains: []world.Domain{
				"now.benjamin-borbe.de",
			},
		},
		&app.HelloWorld{
			Cluster: netcup,
			Tag:     "1.0.1",
			Domains: []world.Domain{
				"rocketsource.de",
				"www.rocketsource.de",
				"rocketnews.de",
				"www.rocketnews.de",
			},
		},
		&app.Slideshow{
			Cluster: netcup,
			Domains: []world.Domain{
				"slideshow.benjamin-borbe.de",
			},
			GitSyncVersion: gitSyncVersion,
		},
		&app.Kickstart{
			Cluster: netcup,
			Domains: []world.Domain{
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			},
			GitSyncVersion: gitSyncVersion,
		},
		//&app.Ldap{
		//Cluster: netcup,
		//	Tag:                "1.1.0",
		//	TeamvaultConnector: c.TeamvaultConnector,
		//},
		//&app.Confluence{
		//Cluster: netcup,
		//	Domains: []world.Domain{
		//		"confluence.benjamin-borbe.de",
		//	},
		//	Tag: "6.9.3",
		//},
	}
}

func (c *Configuration) teamvaultPassword(key model.TeamvaultKey) world.SecretValue {
	return &world.SecretFromTeamvault{
		TeamvaultConnector: c.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

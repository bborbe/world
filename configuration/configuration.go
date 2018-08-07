package configuration

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/app"
)

type Configuration struct {
}

func (c *Configuration) Applier() world.Applier {
	return nil
}

func (c *Configuration) Validate(ctx context.Context) error {
	return nil
}

func (c *Configuration) Childs() []world.Configuration {
	return []world.Configuration{
		&app.Download{
			Context: "netcup",
			Domains: []world.Domain{
				"dl.benjamin-borbe.de",
			},
			NfsServer: "185.170.112.48",
		},
		&app.Mumble{
			Context: "netcup",
			Tag:     "1.0.2",
		},
		&app.Ip{
			Context: "netcup",
			Tag:     "1.1.0",
			Domains: []world.Domain{
				"ip.benjamin-borbe.de",
			},
		},
		&app.Password{
			Context: "netcup",
			Tag:     "1.1.0",
			Domains: []world.Domain{
				"password.benjamin-borbe.de",
			},
		},
		&app.Now{
			Context: "netcup",
			Tag:     "1.0.1",
			Domains: []world.Domain{
				"now.benjamin-borbe.de",
			},
		},
		&app.HelloWorld{
			Context: "netcup",
			Tag:     "1.0.1",
			Domains: []world.Domain{
				"rocketsource.de",
				"www.rocketsource.de",
				"rocketnews.de",
				"www.rocketnews.de",
			},
		},
		&app.Slideshow{
			Context: "netcup",
			Domains: []world.Domain{
				"slideshow.benjamin-borbe.de",
			},
		},
		&app.Kickstart{
			Context: "netcup",
			Domains: []world.Domain{
				"kickstart.benjamin-borbe.de",
				"ks.benjamin-borbe.de",
			},
		},
		//&app.Ldap{
		//	Context:   "netcup",
		//	Tag:       "1.1.0",
		//	NfsServer: "185.170.112.48",
		//},
		//&app.Confluence{
		//	Context: "netcup",
		//	Domains: []world.Domain{
		//		"confluence.benjamin-borbe.de",
		//	},
		//	Tag: "6.9.3",
		//},
	}
}

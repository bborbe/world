package configuration

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/app"
)

type Configuration struct {
}

func (c *Configuration) Applier() world.Applier {
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
	}
}

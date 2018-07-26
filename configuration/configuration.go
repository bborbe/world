package configuration

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/download"
	"github.com/bborbe/world/configuration/hello_world"
	"github.com/bborbe/world/configuration/ip"
	"github.com/bborbe/world/configuration/mumble"
	"github.com/bborbe/world/configuration/password"
)

func Apps() world.Apps {
	return world.Apps{
		&download.App{
			Context: "netcup",
			Tag:     "latest",
			Domains: []world.Domain{
				"dl.benjamin-borbe.de",
			},
			NfsServer: "185.170.112.48",
		},
		&mumble.App{
			Context: "netcup",
			Tag:     "1.0.2",
		},
		&ip.App{
			Context: "netcup",
			Tag:     "1.1.0",
			Domains: []world.Domain{
				"ip.benjamin-borbe.de",
			},
		},
		&password.App{
			Context: "netcup",
			Tag:     "1.1.0",
			Domains: []world.Domain{
				"password.benjamin-borbe.de",
			},
		},
		&hello_world.App{
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

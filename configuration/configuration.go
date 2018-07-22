package configuration

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/builder/docker"
	"github.com/bborbe/world/pkg/builder/golang"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/uploader"
)

func Apps() world.Apps {
	return world.Apps{
		IpApp(),
		HelloWorldApp(),
	}
}

func IpApp() world.App {
	ipImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/ip",
		Tag:        "1.1.0",
	}
	return world.App{
		Name: "ip",
		Builder: &golang.Builder{
			Name:            "ip",
			GitRepo:         "https://github.com/bborbe/ip.git",
			SourceDirectory: "github.com/bborbe/ip",
			Package:         "github.com/bborbe/ip/cmd/ip-server",
			Image:           ipImage,
		},
		Uploader: &uploader.Uploader{
			Image: ipImage,
		},
		Deployer: &k8s.Deployer{
			Name:    "ip",
			Context: "netcup",
			Image:   ipImage,
			Domains: []world.Domain{
				"ip.benjamin-borbe.de",
			},
			Args: []world.Arg{"-logtostderr", "-v=2"},
			Port: 8080,
		},
	}
}

func HelloWorldApp() world.App {
	helloWorldImage := world.Image{
		Registry:   "docker.io",
		Repository: "bborbe/hello-world",
		Tag:        "1.0.1",
	}
	return world.App{
		Name: "hello-world",
		Builder: &docker.Builder{
			GitRepo: "https://github.com/bborbe/hello-world.git",
			Image:   helloWorldImage,
		},
		Uploader: &uploader.Uploader{
			Image: helloWorldImage,
		},
		Deployer: &k8s.Deployer{
			Name:    "hello-world",
			Context: "netcup",
			Image:   helloWorldImage,
			Domains: []world.Domain{
				"rocketsource.de",
				"www.rocketsource.de",
				"rocketnews.de",
				"www.rocketnews.de",
			},
			Port: 80,
		},
	}
}

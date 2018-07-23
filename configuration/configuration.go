package configuration

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/builder/docker"
	"github.com/bborbe/world/pkg/builder/golang"
	"github.com/bborbe/world/pkg/deploy/k8s"
	"github.com/bborbe/world/pkg/uploader"
)

func Apps() world.Apps {
	return world.Apps{
		{
			Name: "ip",
			Deployer: &k8s.Deployer{
				Namespace: "ip",
				Context:   "netcup",
				Domains: []world.Domain{
					"ip.benjamin-borbe.de",
				},
				Args: []world.Arg{"-logtostderr", "-v=2"},
				Port: 8080,
				Uploader: &uploader.Uploader{
					Builder: &golang.Builder{
						Name:            "ip",
						GitRepo:         "https://github.com/bborbe/ip.git",
						SourceDirectory: "github.com/bborbe/ip",
						Package:         "github.com/bborbe/ip/cmd/ip-server",
						Image: world.Image{
							Registry:   "docker.io",
							Repository: "bborbe/ip",
							Tag:        "1.1.0",
						},
					},
				},
			},
		},
		{
			Name: "password",
			Deployer: &k8s.Deployer{
				Namespace: "password",
				Context:   "netcup",
				Domains: []world.Domain{
					"password.benjamin-borbe.de",
				},
				Args: []world.Arg{"-logtostderr", "-v=2"},
				Port: 8080,
				Uploader: &uploader.Uploader{
					Builder: &golang.Builder{
						Name:            "password",
						GitRepo:         "https://github.com/bborbe/password.git",
						SourceDirectory: "github.com/bborbe/password",
						Package:         "github.com/bborbe/password/cmd/password-server",
						Image: world.Image{
							Registry:   "docker.io",
							Repository: "bborbe/password",
							Tag:        "1.1.0",
						},
					},
				},
			},
		},
		{
			Name: "hello-world",
			Deployer: &k8s.Deployer{
				Namespace: "hello-world",
				Context:   "netcup",
				Domains: []world.Domain{
					"rocketsource.de",
					"www.rocketsource.de",
					"rocketnews.de",
					"www.rocketnews.de",
				},
				Port: 80,
				Uploader: &uploader.Uploader{
					Builder: &docker.Builder{
						GitRepo: "https://github.com/bborbe/hello-world.git",
						Image: world.Image{
							Registry:   "docker.io",
							Repository: "bborbe/hello-world",
							Tag:        "1.0.1",
						},
					},
				},
			},
		},
	}
}

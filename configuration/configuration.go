package configuration

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/builder/clone"
	"github.com/bborbe/world/pkg/builder/docker"
	"github.com/bborbe/world/pkg/builder/golang"
	"github.com/bborbe/world/pkg/deploy/k8s"
	"github.com/bborbe/world/pkg/uploader"
)

func Apps() world.Apps {
	return world.Apps{
		{
			Name: "download",
			Deployer: &k8s.Deployer{
				Namespace:     "download",
				Context:       "netcup",
				Port:          80,
				CpuLimit:      "250m",
				MemoryLimit:   "25Mi",
				CpuRequest:    "10m",
				MemoryRequest: "10Mi",
				Domains: []world.Domain{
					"dl.benjamin-borbe.de",
				},
				Mounts: []world.Mount{
					{
						Name:      "download",
						Target:    "/usr/share/nginx/html",
						ReadOnly:  true,
						NfsPath:   "/data/download",
						NfsServer: "185.170.112.48",
					},
				},
				Uploader: &uploader.Uploader{
					Builder: &clone.Builder{
						SourceImage: world.Image{
							Registry:   "docker.io",
							Repository: "jrelva/nginx-autoindex",
							Tag:        "latest",
						},
						TargetImage: world.Image{
							Registry:   "docker.io",
							Repository: "bborbe/nginx-autoindex",
							Tag:        "latest",
						},
					},
				},
			},
		},
		{
			Name: "mumble",
			Deployer: &k8s.Deployer{
				Namespace:     "mumble",
				Context:       "netcup",
				Port:          64738,
				HostPort:      64738,
				CpuLimit:      "200m",
				MemoryLimit:   "100Mi",
				CpuRequest:    "100m",
				MemoryRequest: "25Mi",
				Uploader: &uploader.Uploader{
					Builder: &docker.Builder{
						GitRepo: "https://github.com/bborbe/mumble.git",
						Image: world.Image{
							Registry:   "docker.io",
							Repository: "bborbe/mumble",
							Tag:        "1.0.2",
						},
					},
				},
			},
		},
		{
			Name: "ip",
			Deployer: &k8s.Deployer{
				Namespace: "ip",
				Context:   "netcup",
				Domains: []world.Domain{
					"ip.benjamin-borbe.de",
				},
				CpuLimit:      "100",
				MemoryLimit:   "50Mi",
				CpuRequest:    "10m",
				MemoryRequest: "10Mi",
				Args:          []world.Arg{"-logtostderr", "-v=2"},
				Port:          8080,
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
				CpuLimit:      "100",
				MemoryLimit:   "50Mi",
				CpuRequest:    "10m",
				MemoryRequest: "10Mi",
				Args:          []world.Arg{"-logtostderr", "-v=2"},
				Port:          8080,
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
				CpuLimit:      "100",
				MemoryLimit:   "50Mi",
				CpuRequest:    "10m",
				MemoryRequest: "10Mi",
				Port:          80,
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

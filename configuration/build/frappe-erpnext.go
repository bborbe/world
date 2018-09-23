package build

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type FrappeErpnext struct {
	Image        docker.Image
	BenchVersion docker.Tag
}

func (f *FrappeErpnext) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.Image,
		f.BenchVersion,
	)
}

func (f *FrappeErpnext) Children() []world.Configuration {
	return []world.Configuration{
		&buildConfiguration{
			&docker.Builder{
				GitRepo:   "https://github.com/bborbe/frappe-erpnext.git",
				Image:     f.Image,
				GitBranch: docker.GitBranch(f.Image.Tag),
				BuildArgs: docker.BuildArgs{
					"FRAPPE_BENCH_VERSION": f.BenchVersion.String(),
				},
			},
		},
	}
}

func (f *FrappeErpnext) Applier() (world.Applier, error) {
	return &docker.Uploader{
		Image: f.Image,
	}, nil
}

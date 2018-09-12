package foo

import (
	"context"

	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Readme struct {
}

func (r *Readme) Children() []world.Configuration {
	return nil
}

func (r *Readme) Applier() (world.Applier, error) {
	return &remote.File{
		SSH: ssh.SSH{
			Host:           "pn.benjamin-borbe.de:22",
			User:           "bborbe",
			PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
		},
		Path:    "/tmp/readme.txt",
		Content: []byte("hello world\n"),
		User:    "root",
		Group:   "root",
		Perm:    0644,
	}, nil
}

func (r *Readme) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
	)
}

package service

import (
	"context"

	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type File struct {
	SSH     ssh.SSH
	Path    remote.Path
	Content remote.HasContent
	User    remote.User
	Group   remote.Group
	Perm    remote.Perm
}

func (f *File) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&remote.File{
			SSH:     f.SSH,
			Path:    f.Path,
			Content: f.Content,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chown{
			SSH:   f.SSH,
			Path:  f.Path,
			User:  f.User,
			Group: f.Group,
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chmod{
			SSH:  f.SSH,
			Path: f.Path,
			Perm: f.Perm,
		}),
	}
}

func (f *File) Applier() (world.Applier, error) {
	return nil, nil
}

func (f *File) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.SSH,
		f.Path,
		f.User,
		f.Group,
		f.Perm,
	)
}

package remote

import (
	"context"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
)

type File struct {
	SSH ssh.SSH

	Path    string
	Content []byte
	User    string
	Group   string
	Perm    int
}

func (f *File) Satisfied(ctx context.Context) (bool, error) {
	return f.SSH.Exists(ctx, f.Path)
}

func (f *File) Apply(ctx context.Context) error {
	return f.SSH.CreateFile(ctx, f.Path, f.Content)
}

func (f *File) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.SSH,
	)
}

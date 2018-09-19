package remote

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
)

type Directory struct {
	SSH ssh.SSH

	Path Path
}

func (f *Directory) Satisfied(ctx context.Context) (bool, error) {
	return f.SSH.RunCommand(ctx, fmt.Sprintf("test -d %s", f.Path)) == nil, nil
}

func (f *Directory) Apply(ctx context.Context) error {
	return errors.Wrap(f.SSH.RunCommand(ctx, fmt.Sprintf("mkdir -p %s", f.Path)), "mkdir failed")
}

func (f *Directory) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.SSH,
		f.Path,
	)
}

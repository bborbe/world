package remote

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type Chmod struct {
	SSH ssh.SSH

	Path Path
	Perm Perm
}

func (f *Chmod) Satisfied(ctx context.Context) (bool, error) {
	stdout, err := f.SSH.RunCommandStdout(ctx, "stat -c '%a' "+f.Path.String())
	if err != nil {
		return false, errors.Wrapf(err, "check stat of %s failed", f.Path)
	}
	return string(stdout) == fmt.Sprintf("%d", f.Perm), nil
}

func (f *Chmod) Apply(ctx context.Context) error {
	return errors.Wrap(f.SSH.RunCommand(ctx, fmt.Sprintf("chmod %s %s", f.Perm, f.Path)), "chown failed")
}

func (f *Chmod) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.SSH,
		f.Path,
		f.Perm,
	)
}

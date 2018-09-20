package remote

import (
	"context"
	"fmt"
	"strings"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type Chown struct {
	SSH ssh.SSH

	Path  Path
	User  User
	Group Group
}

func (f *Chown) Satisfied(ctx context.Context) (bool, error) {
	stdout, err := f.SSH.RunCommandStdout(ctx, "stat -c '%U:%G' "+f.Path.String())
	if err != nil {
		return false, errors.Wrapf(err, "check stat of %s failed", f.Path)
	}
	return strings.TrimSpace(string(stdout)) == fmt.Sprintf("%s:%s", f.User, f.Group), nil
}

func (f *Chown) Apply(ctx context.Context) error {
	return errors.Wrap(f.SSH.RunCommand(ctx, fmt.Sprintf("chown %s:%s %s", f.User, f.Group, f.Path)), "chown failed")
}

func (f *Chown) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.SSH,
		f.Path,
		f.User,
		f.Group,
	)
}

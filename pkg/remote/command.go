package remote

import (
	"context"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type Command struct {
	SSH ssh.SSH

	Command string
}

func (f *Command) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (f *Command) Apply(ctx context.Context) error {
	return f.SSH.RunCommand(ctx, f.Command)
}

func (f *Command) Validate(ctx context.Context) error {
	if f.Command == "" {
		return errors.New("Command missing")
	}
	return validation.Validate(
		ctx,
		f.SSH,
	)
}
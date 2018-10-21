package remote

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type File struct {
	SSH     ssh.SSH
	Path    Path
	Content HasContent
}

func (f *File) Satisfied(ctx context.Context) (bool, error) {
	content, err := f.Content.Content(ctx)
	if err != nil {
		return false, errors.Wrap(err, "get content failed")
	}
	h := md5.New()
	h.Write(content)
	return f.SSH.RunCommand(ctx, fmt.Sprintf(`echo "%s %s" | md5sum -c`, fmt.Sprintf("%x", h.Sum(nil)), f.Path)) == nil, nil
}

func (f *File) Apply(ctx context.Context) error {
	content, err := f.Content.Content(ctx)
	if err != nil {
		return errors.Wrap(err, "get content failed")
	}
	return errors.Wrap(f.SSH.RunCommandStdin(ctx, fmt.Sprintf("cat > %s", f.Path), content), "create file failed")
}

func (f *File) Validate(ctx context.Context) error {
	if f.Content == nil {
		return fmt.Errorf("Content missing of %s", f.Path)
	}
	return validation.Validate(
		ctx,
		f.SSH,
		f.Path,
	)
}

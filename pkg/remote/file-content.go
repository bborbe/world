// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"

	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type FileContent struct {
	SSH     *ssh.SSH
	Path    file.HasPath
	Content content.HasContent
}

func (f *FileContent) Satisfied(ctx context.Context) (bool, error) {
	content, err := f.Content.Content(ctx)
	if err != nil {
		return false, errors.Wrap(err, "get content failed")
	}
	h := md5.New()
	h.Write(content)
	path, err := f.Path.Path(ctx)
	if err != nil {
		return false, err
	}
	return f.SSH.RunCommand(ctx, fmt.Sprintf(`echo "%s %s" | md5sum -c`, fmt.Sprintf("%x", h.Sum(nil)), path)) == nil, nil
}

func (f *FileContent) Apply(ctx context.Context) error {
	content, err := f.Content.Content(ctx)
	if err != nil {
		return errors.Wrap(err, "get content failed")
	}
	path, err := f.Path.Path(ctx)
	if err != nil {
		return err
	}
	return errors.Wrap(f.SSH.RunCommandStdin(ctx, fmt.Sprintf("cat > %s", path), content), "create file failed")
}

func (f *FileContent) Validate(ctx context.Context) error {
	if f.Content == nil {
		return fmt.Errorf("Content missing of %s", f.Path)
	}
	if f.Path == nil {
		return fmt.Errorf("Path missing of %s", f.Path)
	}
	return validation.Validate(
		ctx,
		f.SSH,
	)
}

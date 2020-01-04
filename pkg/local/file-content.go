// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package local

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/file"
	"github.com/pkg/errors"
)

type FileContent struct {
	Path    file.HasPath
	Content content.HasContent
}

func (f *FileContent) Satisfied(ctx context.Context) (bool, error) {
	path, err := f.Path.Path(ctx)
	if err != nil {
		return false, err
	}
	if _, err = os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func (f *FileContent) Apply(ctx context.Context) error {
	path, err := f.Path.Path(ctx)
	if err != nil {
		return err
	}
	content, err := f.Content.Content(ctx)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, content, 0600)
}

func (f *FileContent) Validate(ctx context.Context) error {
	if f.Content == nil {
		return errors.New("Content missing")
	}
	if f.Path == nil {
		return errors.New("Path missing")
	}
	return nil
}

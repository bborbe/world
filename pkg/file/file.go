// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package file

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

type HasPath interface {
	Path(ctx context.Context) (string, error)
}

type PathFunc func(ctx context.Context) (string, error)

func (p PathFunc) Path(ctx context.Context) (string, error) {
	return p(ctx)
}

type Path string

func (p Path) Path(ctx context.Context) (string, error) {
	return p.String(), nil
}

func (p Path) String() string {
	return string(p)
}

func (p Path) Validate(ctx context.Context) error {
	if p == "" {
		return errors.New("Path missing")
	}
	return nil
}

type User string

func (f User) Validate(ctx context.Context) error {
	if f == "" {
		return errors.New("User missing")
	}
	return nil
}

type Group string

func (f Group) Validate(ctx context.Context) error {
	if f == "" {
		return errors.New("Group missing")
	}
	return nil
}

type Perm uint32

func (f Perm) Validate(ctx context.Context) error {
	if f == 0 {
		return errors.New("Perm missing")
	}
	return nil
}

func (f Perm) String() string {
	return fmt.Sprintf("%04o", uint32(f))
}

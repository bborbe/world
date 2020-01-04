// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package content

import (
	"context"
	"io/ioutil"
)

type HasContent interface {
	Content(ctx context.Context) ([]byte, error)
}

type Static []byte

func (s Static) Content(ctx context.Context) ([]byte, error) {
	return s, nil
}

type Func func(ctx context.Context) ([]byte, error)

func (f Func) Content(ctx context.Context) ([]byte, error) {
	return f(ctx)
}

type File string

func (f File) Content(ctx context.Context) ([]byte, error) {
	return ioutil.ReadFile(f.String())
}

func (f File) String() string {
	return string(f)
}

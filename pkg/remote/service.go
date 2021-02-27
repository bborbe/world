// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"context"

	"github.com/pkg/errors"
)

type ServiceName string

func (s ServiceName) String() string {
	return string(s)
}

func (h ServiceName) Validate(ctx context.Context) error {
	if h == "" {
		return errors.New("ServiceName missing")
	}
	return nil
}

// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"

	"github.com/pkg/errors"
)

type Kind string

func (k Kind) Validate(ctx context.Context) error {
	if k == "" {
		return errors.New("Kind missing")
	}
	return nil
}

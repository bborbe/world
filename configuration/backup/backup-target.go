// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package backup

import (
	"context"

	"github.com/pkg/errors"
)

type BackupTarget struct {
	User      string
	Host      string
	IP        string
	Port      int
	Excludes  []string
	Directory string
}

func (b BackupTarget) Validate(ctx context.Context) error {
	if b.Directory == "" {
		return errors.New("Directory missing")
	}
	if b.User == "" {
		return errors.New("User missing")
	}
	if b.Host == "" {
		return errors.New("Host missing")
	}
	if b.Port <= 0 || b.Port >= 65535 {
		return errors.Errorf("invalid port %d", b.Port)
	}
	return nil
}

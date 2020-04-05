// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"github.com/bborbe/world/pkg/world"
)

func Configuration(applier world.Applier) world.Configuration {
	return world.NewConfiguraionBuilder().WithApplier(applier)
}

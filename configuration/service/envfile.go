// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"fmt"
)

type EnvFile map[string]string

func (e EnvFile) Content(ctx context.Context) ([]byte, error) {
	result := &bytes.Buffer{}
	for k, v := range e {
		fmt.Fprintf(result, "%s=%s\n", k, v)
	}
	return result.Bytes(), nil
}

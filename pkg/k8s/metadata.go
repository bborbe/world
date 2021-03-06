// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/validation"
)

type MetadataName string

func (m MetadataName) String() string {
	return string(m)
}

func (m MetadataName) Validate(ctx context.Context) error {
	if m == "" {
		return errors.New("NamespaceName missing")
	}
	return nil
}

type Metadata struct {
	Namespace   NamespaceName `yaml:"namespace,omitempty"`
	Name        MetadataName  `yaml:"name,omitempty"`
	Labels      Labels        `yaml:"labels,omitempty"`
	Annotations Annotations   `yaml:"annotations,omitempty"`
}

func (m Metadata) String() string {
	return fmt.Sprintf("ns: %s name: %s", m.Namespace, m.Name)
}

func (m Metadata) Validate(ctx context.Context) error {
	return validation.Validate(ctx,
		m.Name,
		m.Namespace,
	)
}

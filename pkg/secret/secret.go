// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package secret

import (
	"context"

	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/pkg/errors"
)

type Teamvault struct {
	TeamvaultConnector teamvault.Connector
}

func (c *Teamvault) Password(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultPassword{
		TeamvaultConnector: c.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

func (c *Teamvault) Htpasswd(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultHtpasswd{
		TeamvaultConnector: c.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

func (c *Teamvault) Username(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultUser{
		TeamvaultConnector: c.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

func (c *Teamvault) File(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultFile{
		TeamvaultConnector: c.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

func (w *Teamvault) Validate(ctx context.Context) error {
	if w.TeamvaultConnector == nil {
		return errors.New("TeamvaultConnector missing")
	}
	return nil
}

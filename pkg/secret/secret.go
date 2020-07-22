// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package secret

import (
	"context"

	teamvault "github.com/bborbe/teamvault-utils"
	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/deployer"
)

type Teamvault struct {
	TeamvaultConnector teamvault.Connector
}

func (t *Teamvault) Password(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultPassword{
		TeamvaultConnector: t.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

func (t *Teamvault) Htpasswd(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultHtpasswd{
		TeamvaultConnector: t.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

func (t *Teamvault) Username(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultUser{
		TeamvaultConnector: t.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

func (t *Teamvault) File(key teamvault.Key) deployer.SecretValue {
	return &deployer.SecretFromTeamvaultFile{
		TeamvaultConnector: t.TeamvaultConnector,
		TeamvaultKey:       key,
	}
}

func (t *Teamvault) Validate(ctx context.Context) error {
	if t.TeamvaultConnector == nil {
		return errors.New("TeamvaultConnector missing")
	}
	return nil
}

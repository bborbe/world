package secret

import (
	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/world/configuration/deployer"
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

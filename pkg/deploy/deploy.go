package deploy

import (
	"context"

	"github.com/bborbe/world"
	"github.com/pkg/errors"
)

func DeployIfNeeded(ctx context.Context, deployer world.Deployer) error {
	ok, err := deployer.Satisfied(ctx)
	if err != nil {
		return errors.Wrap(err, "check satisfied failed")
	}
	if !ok {
		if err := deployer.Deploy(ctx); err != nil {
			return errors.Wrap(err, "deploy failed")
		}
	}
	return nil
}

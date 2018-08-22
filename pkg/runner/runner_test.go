package runner_test

import (
	"context"
	"testing"

	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/world/configuration"
	"github.com/bborbe/world/pkg/runner"
	"github.com/bborbe/world/pkg/secret"
)

func TestValidate(t *testing.T) {
	r, err := runner.New(
		&configuration.World{
			TeamvaultSecrets: &secret.Teamvault{
				TeamvaultConnector: connector.NewDummy(),
			},
		})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
}

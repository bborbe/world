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
	ctx := context.Background()
	b := runner.Builder{
		Configuration: &configuration.World{
			TeamvaultSecrets: &secret.Teamvault{
				TeamvaultConnector: connector.NewDummy(),
			},
		},
	}
	r, err := b.Build(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Validate(ctx); err != nil {
		t.Fatal(err)
	}

}

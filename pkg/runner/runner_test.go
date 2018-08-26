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
	world := &configuration.World{
		TeamvaultSecrets: &secret.Teamvault{
			TeamvaultConnector: connector.NewDummy(),
		},
	}
	conf, err := world.Configuration()
	if err != nil {
		t.Fatal(err)
	}
	builder := runner.Builder{
		Configuration: conf,
	}
	r, err := builder.Build(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Validate(ctx); err != nil {
		t.Fatal(err)
	}

}

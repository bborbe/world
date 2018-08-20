package runner_test

import (
	"context"
	"testing"

	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/world/configuration"
	"github.com/bborbe/world/pkg/runner"
)

func TestValidate(t *testing.T) {
	r := &runner.Runner{
		Configuration: &configuration.World{
			TeamvaultConnector: connector.NewDummy(),
		},
	}
	if err := r.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
}

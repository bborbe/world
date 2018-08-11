package configuration_test

import (
	"context"
	"testing"

	"github.com/bborbe/world/configuration"
	"github.com/bborbe/world"
	"github.com/bborbe/teamvault-utils/connector"
)

func TestValidate(t *testing.T) {
	c := &configuration.Configuration{
		TeamvaultConnector: connector.NewDummy(),
	}
	if err := world.Validate(context.Background(), c); err != nil {
		t.Fatal(err)
	}
}

package k8s_test

import (
	"context"
	"testing"

	"github.com/bborbe/world/pkg/k8s"
)

func TestPortNameValidate(t *testing.T) {
	if k8s.PortName("").Validate(context.Background()) == nil {
		t.Fatal("error expected")
	}
	if k8s.PortName("a_b").Validate(context.Background()) == nil {
		t.Fatal("error expected")
	}
	if k8s.PortName("abc").Validate(context.Background()) != nil {
		t.Fatal("no error expected")
	}
}

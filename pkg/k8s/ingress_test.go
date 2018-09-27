package k8s_test

import (
	"context"
	"testing"

	"github.com/bborbe/world/pkg/k8s"
)

func TestIngressHostsValidate(t *testing.T) {

	var tests = []struct {
		name         string
		ingressHosts k8s.IngressHosts
		expectError  bool
	}{
		{
			name:         "empty",
			ingressHosts: k8s.IngressHosts{},
			expectError:  true,
		},
		{
			name: "one",
			ingressHosts: k8s.IngressHosts{
				"a.example.com",
			},
			expectError: false,
		},
		{
			name: "two",
			ingressHosts: k8s.IngressHosts{
				"a.example.com",
				"b.example.com",
			},
			expectError: false,
		},
		{
			name: "invalid",
			ingressHosts: k8s.IngressHosts{
				"",
			},
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ingressHosts.Validate(context.Background())
			if tt.expectError {
				if err == nil {
					t.Fatal("expect error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestIngressHostValidate(t *testing.T) {
	var tests = []struct {
		name        string
		ingressHost k8s.IngressHost
		expectError bool
	}{
		{
			name:        "one",
			ingressHost: "a.example.com",
			expectError: false,
		},
		{
			name:        "invalid",
			ingressHost: "",
			expectError: true,
		},
		{
			name:        "one",
			ingressHost: "a-b.example.com",
			expectError: false,
		},
		{
			name:        "one",
			ingressHost: "a_b.example.com",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ingressHost.Validate(context.Background())
			if tt.expectError {
				if err == nil {
					t.Fatal("expect error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

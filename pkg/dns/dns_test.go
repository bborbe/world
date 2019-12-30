// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns_test

import (
	"context"
	"testing"

	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/network"
)

func TestApply(t *testing.T) {
	d := &dns.Server{
		Host:    "ns.rocketsource.de",
		KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
		List: []dns.Entry{
			{
				Host: "now.benjamin-borbe.de",
				IP:   network.IPStatic("185.170.112.48"),
			},
		},
	}
	err := d.Apply(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidateSuccess(t *testing.T) {
	d := &dns.Server{
		Host:    "ns.rocketsource.de",
		KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
		List: []dns.Entry{
			{
				Host: "now.benjamin-borbe.de",
				IP:   network.IPStatic("185.170.112.48"),
			},
		},
	}
	err := d.Validate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidateFailure(t *testing.T) {
	d := &dns.Server{
		Host:    "ns.rocketsource.de",
		KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
		List: []dns.Entry{
			{
				Host: "now.benjamin-borbe.de",
				IP:   network.IPStatic("2001:db8::68"),
			},
		},
	}
	err := d.Validate(context.Background())
	if err == nil {
		t.Fatal("error expected")
	}
}

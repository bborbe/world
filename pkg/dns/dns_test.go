package dns_test

import (
	"context"
	"net"
	"testing"

	"github.com/bborbe/world/pkg/dns"
)

func TestApply(t *testing.T) {
	d := &dns.Server{
		Host:    "ns.rocketsource.de",
		KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
		List: []dns.Entry{
			{
				Host: "now.benjamin-borbe.de",
				IP:   net.ParseIP("185.170.112.48"),
			},
		},
	}
	err := d.Apply(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

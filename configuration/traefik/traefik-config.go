// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package traefik

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/deployer"
)

type config struct {
	SSL   bool
	Debug bool
}

func (c config) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, `graceTimeOut = 10`)
	if c.Debug {
		fmt.Fprintln(buf, `debug = true`)
		fmt.Fprintln(buf, `logLevel = "DEBUG"`)
	} else {
		fmt.Fprintln(buf, `debug = false`)
		fmt.Fprintln(buf, `logLevel = "INFO"`)
	}
	if c.SSL {
		fmt.Fprintln(buf, `defaultEntryPoints = ["http","https"]`)
		fmt.Fprintln(buf, `[entryPoints]`)
		fmt.Fprintln(buf, `[entryPoints.http]`)
		fmt.Fprintln(buf, `address = ":80"`)
		fmt.Fprintln(buf, `compress = false`)
		fmt.Fprintln(buf, `[entryPoints.http.redirect]`)
		fmt.Fprintln(buf, `entryPoint = "https"`)
		fmt.Fprintln(buf, `[entryPoints.https]`)
		fmt.Fprintln(buf, `address = ":443"`)
		fmt.Fprintln(buf, `compress = false`)
		fmt.Fprintln(buf, `[entryPoints.https.tls]`)
		fmt.Fprintln(buf, `cipherSuites = [`)
		fmt.Fprintln(buf, `"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",`)
		fmt.Fprintln(buf, `"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",`)
		fmt.Fprintln(buf, `"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",`)
		fmt.Fprintln(buf, `]`)
		fmt.Fprintln(buf, `[acme]`)
		fmt.Fprintln(buf, `email = "bborbe@rocketnews.de"`)
		fmt.Fprintln(buf, `storage = "/acme/acme.json"`)
		fmt.Fprintln(buf, `entryPoint = "https"`)
		fmt.Fprintln(buf, `onHostRule = true`)
		fmt.Fprintln(buf, `acmeLogging = true`)
		fmt.Fprintln(buf, `[acme.httpChallenge]`)
		fmt.Fprintln(buf, `entryPoint = "http"`)
	} else {
		fmt.Fprintln(buf, `defaultEntryPoints = ["http"]`)
		fmt.Fprintln(buf, `[entryPoints]`)
		fmt.Fprintln(buf, `[entryPoints.http]`)
		fmt.Fprintln(buf, `address = ":80"`)
		fmt.Fprintln(buf, `compress = false`)
	}
	fmt.Fprintln(buf, `[kubernetes]`)
	fmt.Fprintln(buf, `[web]`)
	fmt.Fprintln(buf, `address = ":8080"`)
	fmt.Fprintln(buf, `[web.metrics.prometheus]`)
	return buf.String()
}

func (c config) ConfigValue() deployer.ConfigValue {
	return deployer.ConfigValueFunc(func(ctx context.Context) (string, error) {
		return c.String(), nil
	})
}

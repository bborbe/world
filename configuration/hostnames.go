// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import "github.com/bborbe/world/pkg/network"

const (
	DebugHostname        = network.Host("debug.benjamin-borbe.de")
	MetabaseHostname     = network.Host("metabase.benjamin-borbe.de")
	GrafanaHostname      = network.Host("grafana.benjamin-borbe.de")
	KafkaStatus          = network.Host("kafka-status.benjamin-borbe.de")
	VersionsHostname     = network.Host("versions.benjamin-borbe.de")
	UpdatesHostname      = network.Host("updates.benjamin-borbe.de")
	KafkaSampleHostname  = network.Host("kafka-sample.benjamin-borbe.de")
	PrometheusHostname   = network.Host("prometheus.benjamin-borbe.de")
	AlertmanagerHostname = network.Host("prometheus-alertmanager.benjamin-borbe.de")
	TeamvaultHostname    = network.Host("teamvault.benjamin-borbe.de")
	ConfluenceHostname   = network.Host("confluence.benjamin-borbe.de")
	JiraHostname         = network.Host("jira.benjamin-borbe.de")
	MailHostname         = network.Host("mail.benjamin-borbe.de")
	MavenHostname        = network.Host("maven.benjamin-borbe.de")
	IPHostname           = network.Host("ip.benjamin-borbe.de")
)

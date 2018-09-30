package app

import (
	"bytes"
	"context"
	"text/template"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/configuration"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Grafana struct {
	Cluster      cluster.Cluster
	Domain       k8s.IngressHost
	LdapUsername deployer.SecretValue
	LdapPassword deployer.SecretValue
	Requirements []world.Configuration
}

func (g *Grafana) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		g.Cluster,
		g.Domain,
		g.LdapPassword,
		g.LdapUsername,
	)
}

func (g *Grafana) Applier() (world.Applier, error) {
	return nil, nil
}

func (g *Grafana) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, g.Requirements...)
	result = append(result, g.grafana()...)
	return result
}

func (g *Grafana) grafana() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/grafana",
		Tag:        "5.2.4",
	}
	port := deployer.Port{
		Port:     3000,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   g.Cluster.Context,
			Namespace: "grafana",
		},
		configuration.New().WithApplier(
			&deployer.ConfigMapApplier{
				Context:   g.Cluster.Context,
				Namespace: "grafana",
				Name:      "grafana",
				ConfigEntryList: deployer.ConfigEntryList{
					deployer.ConfigEntry{
						Key:   "grafana.ini",
						Value: grafanaIni,
					},
					deployer.ConfigEntry{
						Key:       "ldap.toml",
						ValueFrom: g.generateLdapToml,
					},
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   g.Cluster.Context,
			Namespace: "grafana",
			Name:      "grafana",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "grafana",
					Image: image,
					Requirement: &build.Grafana{
						Image: image,
					},
					Ports: []deployer.Port{port},
					Env: []k8s.Env{
						{
							Name:  "GF_PATHS_CONFIG",
							Value: "/config/grafana.ini",
						},
						{
							Name:  "GF_PATHS_DATA",
							Value: "/var/lib/grafana",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "config",
							Path: "/config",
						},
						{
							Name: "data",
							Path: "/var/lib/grafana",
						},
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "25Mi",
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 10,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "grafana",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "grafana.ini",
								Path: "grafana.ini",
							},
							{
								Key:  "ldap.toml",
								Path: "ldap.toml",
							},
						},
					},
				},
				{
					Name: "data",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/grafana",
						Server: g.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   g.Cluster.Context,
			Namespace: "grafana",
			Name:      "grafana",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   g.Cluster.Context,
			Namespace: "grafana",
			Name:      "grafana",
			Port:      port.Name,
			Domains:   k8s.IngressHosts{g.Domain},
		},
	}
}

func (g *Grafana) generateLdapToml(ctx context.Context) (string, error) {
	t, err := template.New("template").Parse(ldapToml)
	if err != nil {
		return "", errors.Wrap(err, "parse ldap toml failed")
	}
	b := &bytes.Buffer{}
	user, err := g.LdapUsername.Value()
	if err != nil {
		return "", errors.Wrap(err, "get ldap username failed")
	}
	password, err := g.LdapPassword.Value()
	if err != nil {
		return "", errors.Wrap(err, "get ldap password failed")
	}
	err = t.Execute(b, struct {
		LdapHost              string
		LdapPort              int
		LdapBindUsename       string
		LdapBindPassword      string
		LdapUserBaseDn        string
		LdapUserSearchFilter  string
		LdapGroupBaseDn       string
		LdapGroupSearchFilter string
		LdapAdminGroupDn      string
		LdapEditorGroupDn     string
		LdapViewerGroupDn     string
	}{
		LdapHost:              "ldap.ldap.svc.cluster.local",
		LdapPort:              389,
		LdapBindUsename:       string(user),
		LdapBindPassword:      string(password),
		LdapUserBaseDn:        "ou=users,dc=benjamin-borbe,dc=de",
		LdapUserSearchFilter:  "(uid=%s)",
		LdapGroupBaseDn:       "ou=groups,dc=benjamin-borbe,dc=de",
		LdapGroupSearchFilter: "(member=uid=%s,ou=users,dc=benjamin-borbe,dc=de)",
		LdapAdminGroupDn:      "Admins",
		LdapEditorGroupDn:     "Admins",
		LdapViewerGroupDn:     "Employees",
	})
	if err != nil {
		return "", errors.Wrap(err, "parse ldapToml failed")
	}
	return b.String(), nil
}

const ldapToml = `
[[servers]]
host = "{{ .LdapHost }}"
port = {{ .LdapPort }}
use_ssl = false
start_tls = false
ssl_skip_verify = false

bind_dn = '{{ .LdapBindUsename }}'
bind_password = '{{ .LdapBindPassword }}'
search_filter = "{{ .LdapUserSearchFilter }}"
search_base_dns = ["{{ .LdapUserBaseDn }}"]

group_search_base_dns = ["{{ .LdapGroupBaseDn }}"]
group_search_filter = "{{ .LdapGroupSearchFilter }}"

[servers.attributes]
name = "givenName"
surname = "sn"
username = "uid"
member_of = "cn"
email =  "mail"

[[servers.group_mappings]]
group_dn = "{{ .LdapAdminGroupDn }}"
org_role = "Admin"

[[servers.group_mappings]]
group_dn = "{{ .LdapAdminGroupDn }}"
org_role = "Admin"

[[servers.group_mappings]]
group_dn = "{{ .LdapEditorGroupDn }}"
org_role = "Editor"

[[servers.group_mappings]]
group_dn = "{{ .LdapViewerGroupDn }}"
org_role = "Viewer"
`

const grafanaIni = `
##################### Grafana Configuration Example #####################
#
# Everything has defaults so you only need to uncomment things you want to
# change

# possible values : production, development
; app_mode = production

# instance name, defaults to HOSTNAME environment variable value or hostname if HOSTNAME var is empty
; instance_name = ${HOSTNAME}

#################################### Paths ####################################
[paths]
# Path to where grafana can store temp files, sessions, and the sqlite3 db (if that is used)
#
;data = /var/lib/grafana
#
# Directory where grafana can store logs
#
;logs = /var/log/grafana
#
# Directory where grafana will automatically scan and look for plugins
#
;plugins = /var/lib/grafana/plugins

#
#################################### Server ####################################
[server]
# Protocol (http, https, socket)
;protocol = http

# The ip address to bind to, empty will bind to all interfaces
;http_addr =

# The http port  to use
;http_port = 3000

# The public facing domain name used to access grafana from a browser
;domain = localhost

# Redirect to correct domain if host header does not match domain
# Prevents DNS rebinding attacks
;enforce_domain = false

# The full public facing url you use in browser, used for redirects and emails
# If you use reverse proxy and sub path specify full url (with sub path)
;root_url = http://localhost:3000

# Log web requests
;router_logging = false

# the path relative working path
;static_root_path = public

# enable gzip
;enable_gzip = false

# https certs & key file
;cert_file =
;cert_key =

# Unix socket path
;socket =

#################################### Database ####################################
[database]
# You can configure the database connection by specifying type, host, name, user and password
# as seperate properties or as on string using the url propertie.

# Either "mysql", "postgres" or "sqlite3", it's your choice
;type = sqlite3
;host = 127.0.0.1:3306
;name = grafana
;user = root
# If the password contains # or ; you have to wrap it with trippel quotes. Ex """#password;"""
;password =

# Use either URL or the previous fields to configure the database
# Example: mysql://user:secret@host:port/database
;url =

# For "postgres" only, either "disable", "require" or "verify-full"
;ssl_mode = disable

# For "sqlite3" only, path relative to data_path setting
;path = grafana.db

# Max idle conn setting default is 2
;max_idle_conn = 2

# Max conn setting default is 0 (mean not set)
;max_open_conn =


#################################### Session ####################################
[session]
# Either "memory", "file", "redis", "mysql", "postgres", default is "file"
;provider = file

# Provider config options
# memory: not have any config yet
# file: session dir path, is relative to grafana data_path
# redis: config like redis server e.g. addr=127.0.0.1:6379,pool_size=100,db=grafana
# mysql: go-sql-driver/mysql dsn config string, e.g. user:password@tcp(127.0.0.1:3306)/database_name
# postgres: user=a password=b host=localhost port=5432 dbname=c sslmode=disable
;provider_config = sessions

# Session cookie name
;cookie_name = grafana_sess

# If you use session in https only, default is false
;cookie_secure = false

# Session life time, default is 86400
;session_life_time = 86400

#################################### Data proxy ###########################
[dataproxy]

# This enables data proxy logging, default is false
;logging = false


#################################### Analytics ####################################
[analytics]
# Server reporting, sends usage counters to stats.grafana.org every 24 hours.
# No ip addresses are being tracked, only simple counters to track
# running instances, dashboard and error counts. It is very helpful to us.
# Change this option to false to disable reporting.
;reporting_enabled = true

# Set to false to disable all checks to https://grafana.net
# for new vesions (grafana itself and plugins), check is used
# in some UI views to notify that grafana or plugin update exists
# This option does not cause any auto updates, nor send any information
# only a GET request to http://grafana.com to get latest versions
;check_for_updates = true

# Google Analytics universal tracking code, only enabled if you specify an id here
;google_analytics_ua_id =

#################################### Security ####################################
[security]
# default admin user, created on startup
;admin_user = admin

# default admin password, can be changed before first start of grafana,  or in profile settings
;admin_password = admin

# used for signing
;secret_key = SW2YcwTIb9zpOOhoPsMm

# Auto-login remember days
;login_remember_days = 7
;cookie_username = grafana_user
;cookie_remember_name = grafana_remember

# disable gravatar profile images
;disable_gravatar = false

# data source proxy whitelist (ip_or_domain:port separated by spaces)
;data_source_proxy_whitelist =

[snapshots]
# snapshot sharing options
;external_enabled = true
;external_snapshot_url = https://snapshots-origin.raintank.io
;external_snapshot_name = Publish to snapshot.raintank.io

# remove expired snapshot
;snapshot_remove_expired = true

# remove snapshots after 90 days
;snapshot_TTL_days = 90

#################################### Users ####################################
[users]
# disable user signup / registration
allow_sign_up = false

# Allow non admin users to create organizations
;allow_org_create = true

# Set to true to automatically assign new users to the default organization (id 1)
;auto_assign_org = true

# Default role new users will be automatically assigned (if disabled above is set to true)
;auto_assign_org_role = Viewer

# Background text for the user field on the login page
;login_hint = email or username

# Default UI theme ("dark" or "light")
;default_theme = dark

# External user management, these options affect the organization users view
;external_manage_link_url =
;external_manage_link_name =
;external_manage_info =

[auth]
# Set to true to disable (hide) the login form, useful if you use OAuth, defaults to false
;disable_login_form = false

# Set to true to disable the signout link in the side menu. useful if you use auth.proxy, defaults to false
;disable_signout_menu = false

#################################### Anonymous Auth ##########################
[auth.anonymous]
# enable anonymous access
;enabled = false

# specify organization name that should be used for unauthenticated users
;org_name = Main Org.

# specify role for unauthenticated users
;org_role = Viewer

#################################### Github Auth ##########################
[auth.github]
;enabled = false
;allow_sign_up = true
;client_id = some_id
;client_secret = some_secret
;scopes = user:email,read:org
;auth_url = https://github.com/login/oauth/authorize
;token_url = https://github.com/login/oauth/access_token
;api_url = https://api.github.com/user
;team_ids =
;allowed_organizations =

#################################### Google Auth ##########################
[auth.google]
;enabled = false
;allow_sign_up = true
;client_id = some_client_id
;client_secret = some_client_secret
;scopes = https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email
;auth_url = https://accounts.google.com/o/oauth2/auth
;token_url = https://accounts.google.com/o/oauth2/token
;api_url = https://www.googleapis.com/oauth2/v1/userinfo
;allowed_domains =

#################################### Generic OAuth ##########################
[auth.generic_oauth]
;enabled = false
;name = OAuth
;allow_sign_up = true
;client_id = some_id
;client_secret = some_secret
;scopes = user:email,read:org
;auth_url = https://foo.bar/login/oauth/authorize
;token_url = https://foo.bar/login/oauth/access_token
;api_url = https://foo.bar/user
;team_ids =
;allowed_organizations =

#################################### Grafana.com Auth ####################
[auth.grafana_com]
;enabled = false
;allow_sign_up = true
;client_id = some_id
;client_secret = some_secret
;scopes = user:email
;allowed_organizations =

#################################### Auth Proxy ##########################
[auth.proxy]
;enabled = false
;header_name = X-WEBAUTH-USER
;header_property = username
;auto_sign_up = true
;ldap_sync_ttl = 60
;whitelist = 192.168.1.1, 192.168.2.1

#################################### Basic Auth ##########################
[auth.basic]
;enabled = true

#################################### Auth LDAP ##########################
[auth.ldap]
enabled = true
config_file = /config/ldap.toml
;allow_sign_up = false

#################################### SMTP / Emailing ##########################
[smtp]
;enabled = false
;host = localhost:25
;user =
# If the password contains # or ; you have to wrap it with trippel quotes. Ex """#password;"""
;password =
;cert_file =
;key_file =
;skip_verify = false
;from_address = admin@grafana.localhost
;from_name = Grafana
# EHLO identity in SMTP dialog (defaults to instance_name)
;ehlo_identity = dashboard.example.com

[emails]
;welcome_email_on_sign_up = false

#################################### Logging ##########################
[log]
# Either "console", "file", "syslog". Default is console and  file
# Use space to separate multiple modes, e.g. "console file"
;mode = console file

# Either "debug", "info", "warn", "error", "critical", default is "info"
;level = info

# optional settings to set different levels for specific loggers. Ex filters = sqlstore:debug
;filters =


# For "console" mode only
[log.console]
;level =

# log line format, valid options are text, console and json
;format = console

# For "file" mode only
[log.file]
;level =

# log line format, valid options are text, console and json
;format = text

# This enables automated log rotate(switch of following options), default is true
;log_rotate = true

# Max line number of single file, default is 1000000
;max_lines = 1000000

# Max size shift of single file, default is 28 means 1 << 28, 256MB
;max_size_shift = 28

# Segment log daily, default is true
;daily_rotate = true

# Expired days of log file(delete after max days), default is 7
;max_days = 7

[log.syslog]
;level =

# log line format, valid options are text, console and json
;format = text

# Syslog network type and address. This can be udp, tcp, or unix. If left blank, the default unix endpoints will be used.
;network =
;address =

# Syslog facility. user, daemon and local0 through local7 are valid.
;facility =

# Syslog tag. By default, the process' argv[0] is used.
;tag =


#################################### AMQP Event Publisher ##########################
[event_publisher]
;enabled = false
;rabbitmq_url = amqp://localhost/
;exchange = grafana_events

;#################################### Dashboard JSON files ##########################
[dashboards.json]
;enabled = false
;path = /var/lib/grafana/dashboards

#################################### Alerting ############################
[alerting]
# Disable alerting engine & UI features
;enabled = true
# Makes it possible to turn off alert rule execution but alerting UI is visible
;execute_alerts = true

#################################### Internal Grafana Metrics ##########################
# Metrics available at HTTP API Url /api/metrics
[metrics]
# Disable / Enable internal metrics
;enabled           = true

# Publish interval
;interval_seconds  = 10

# Send internal metrics to Graphite
[metrics.graphite]
# Enable by setting the address setting (ex localhost:2003)
;address =
;prefix = prod.grafana.%(instance_name)s.

#################################### Distributed tracing ############
[tracing.jaeger]
# Enable by setting the address sending traces to jaeger (ex localhost:6831)
;address = localhost:6831
# Tag that will always be included in when creating new spans. ex (tag1:value1,tag2:value2)
;always_included_tag = tag1:value1
# Type specifies the type of the sampler: const, probabilistic, rateLimiting, or remote
;sampler_type = const
# jaeger samplerconfig param
# for "const" sampler, 0 or 1 for always false/true respectively
# for "probabilistic" sampler, a probability between 0 and 1
# for "rateLimiting" sampler, the number of spans per second
# for "remote" sampler, param is the same as for "probabilistic"
# and indicates the initial sampling rate before the actual one
# is received from the mothership
;sampler_param = 1

#################################### Grafana.com integration  ##########################
# Url used to to import dashboards directly from Grafana.com
[grafana_com]
;url = https://grafana.com

#################################### External image storage ##########################
[external_image_storage]
# Used for uploading images to public servers so they can be included in slack/email messages.
# you can choose between (s3, webdav, gcs)
;provider =

[external_image_storage.s3]
;bucket =
;region =
;path =
;access_key =
;secret_key =

[external_image_storage.webdav]
;url =
;public_url =
;username =
;password =

[external_image_storage.gcs]
;key_file =
;bucket =
`

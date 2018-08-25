package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Proxy struct {
	Cluster  cluster.Cluster
	Password deployer.SecretValue
}

func (d *Proxy) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Cluster,
		d.Password,
	)
}

func (p *Proxy) Children() []world.Configuration {
	squidImage := docker.Image{
		Repository: "bborbe/squid",
		Tag:        "1.2.0",
	}
	privoxyImage := docker.Image{
		Repository: "bborbe/privoxy",
		Tag:        "1.2.0",
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "proxy",
		},
		&deployer.ConfigMapDeployer{
			Context:   p.Cluster.Context,
			Namespace: "proxy",
			Name:      "proxy",
			ConfigMapData: k8s.ConfigMapData{
				"user.action": privoxyUserAction,
				"user.filter": privoxyUserConfig,
			},
		},
		&deployer.SecretDeployer{
			Context:   p.Cluster.Context,
			Namespace: "proxy",
			Name:      "proxy",
			Secrets: deployer.Secrets{
				"htpasswd": p.Password,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   p.Cluster.Context,
			Namespace: "proxy",
			Name:      "proxy",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name:  "squid",
					Image: squidImage,
					Requirement: &build.Squid{
						Image: squidImage,
					},
					Ports: []deployer.Port{
						{
							Port:     3128,
							HostPort: 3128,
							Name:     "squid",
							Protocol: "TCP",
						},
					},
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "200m",
							Memory: "200Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "50m",
							Memory: "50Mi",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "auth",
							Path: "/etc/squid/auth",
						},
					},
				},
				{
					Name:  "privoxy",
					Image: privoxyImage,
					Requirement: &build.Privoxy{
						Image: privoxyImage,
					},
					Ports: []deployer.Port{
						{
							Port:     8118,
							Name:     "privoxy",
							Protocol: "TCP",
						},
					},
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "100m",
							Memory: "100Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "config",
							Path: "/etc/privoxy/user",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "auth",
					Secret: k8s.PodVolumeSecret{
						Name: "proxy",
						Items: []k8s.PodSecretItem{
							{
								Key:  "htpasswd",
								Path: "htpasswd",
							},
						},
					},
				},
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "proxy",
						Items: []k8s.PodConfigMapItem{
							{
								Key:  "user.action",
								Path: "user.action",
							},
							{
								Key:  "user.filter",
								Path: "user.filter",
							},
						},
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "proxy",
			Name:      "proxy",
			Ports: []deployer.Port{
				{
					Port:     3128,
					Name:     "squid",
					Protocol: "TCP",
				},
				{
					Port:     8118,
					Name:     "privoxy",
					Protocol: "TCP",
				},
			},
		},
	}
}

func (p *Proxy) Applier() (world.Applier, error) {
	return nil, nil
}

const privoxyUserAction = `
######################################################################
#
#  File        :  $Source: /cvsroot/ijbswa/current/user.action,v $
#
#  $Id: user.action,v 1.13 2011/11/06 11:36:01 fabiankeil Exp $
#
#  Purpose     :  User-maintained actions file, see
#                 http://www.privoxy.org/user-manual/actions-file.html
#
######################################################################

# This is the place to add your personal exceptions and additions to
# the general policies as defined in default.action. (Here they will be
# safe from updates to default.action.) Later defined actions always
# take precedence, so anything defined here should have the last word.

# See http://www.privoxy.org/user-manual/actions-file.html, or the
# comments in default.action, for an explanation of what an "action" is
# and what each action does.

# The examples included here either use bogus sites, or have the actual
# rules commented out (with the '#' character). Useful aliases are
# included in the top section as a convenience.

#############################################################################
# Aliases
#############################################################################
{{"{{"}}alias{{"}}"}}
#############################################################################
#
# You can define a short form for a list of permissions - e.g., instead
# of "-crunch-incoming-cookies -crunch-outgoing-cookies -filter -fast-redirects",
# you can just write "shop". This is called an alias.
#
# Currently, an alias can contain any character except space, tab, '=', '{'
# or '}'.
# But please use only 'a'-'z', '0'-'9', '+', and '-'.
#
# Alias names are not case sensitive.
#
# Aliases beginning with '+' or '-' may be used for system action names
# in future releases - so try to avoid alias names like this.  (e.g.
# "+crunch-all-cookies" below is not a good name)
#
# Aliases must be defined before they are used.
#
# These aliases just save typing later:
#
+crunch-all-cookies = +crunch-incoming-cookies +crunch-outgoing-cookies
-crunch-all-cookies = -crunch-incoming-cookies -crunch-outgoing-cookies
 allow-all-cookies  = -crunch-all-cookies -session-cookies-only -filter{content-cookies}
 allow-popups       = -filter{all-popups} -filter{unsolicited-popups}
+block-as-image     = +block{Blocked image request.} +handle-as-image
-block-as-image     = -block

# These aliases define combinations of actions
# that are useful for certain types of sites:
#
fragile     = -block -crunch-all-cookies -filter -fast-redirects -hide-referer -prevent-compression
shop        = -crunch-all-cookies allow-popups

# Your favourite blend of filters:
#
myfilters   = +filter{html-annoyances} +filter{js-annoyances} +filter{all-popups}\
              +filter{webbugs} +filter{banners-by-size}

# Allow ads for selected useful free sites:
#
allow-ads   = -block -filter{banners-by-size} -filter{banners-by-link}
#... etc.  Customize to your heart's content.

## end aliases ########################################################
#######################################################################

# Begin examples: #####################################################

# Say you have accounts on some sites that you visit regularly, and you
# don't want to have to log in manually each time. So you'd like to allow
# persistent cookies for these sites. The allow-all-cookies alias defined
# above does exactly that, i.e. it disables crunching of cookies in any
# direction, and the processing of cookies to make them only temporary.
#
{ allow-all-cookies }
#.sourceforge.net
#sunsolve.sun.com
#slashdot.org
#.yahoo.com
#.msdn.microsoft.com
#.redhat.com

# Say the site where you do your homebanking needs to open popup
# windows, but you have chosen to kill popups uncoditionally by default.
# This will allow it for your-example-bank.com:
#
{ -filter{all-popups} }
.banking.example.com

# Some hosts and some file types you may not want to filter for
# various reasons:
#
{ -filter }

# Technical documentation is likely to contain strings that might
# erroneously get altered by the JavaScript-oriented filters:
#
#.tldp.org
#/(.*/)?selfhtml/

# And this stupid host sends streaming video with a wrong MIME type,
# so that Privoxy thinks it is getting HTML and starts filtering:
#
stupid-server.example.com/

# Example of a simple "block" action. Say you've seen an ad on your
# favourite page on example.com that you want to get rid of. You have
# right-clicked the image, selected "copy image location" and pasted
# the URL below while removing the leading http://, into a { +block{reason} }
# section. Note that { +handle-as-image } need not be specified, since
# all URLs ending in .gif will be tagged as images by the general rules
# as set in default.action anyway:
#
{ +block{Nasty ads.} }
www.example.com/nasty-ads/sponsor.gif

# The URLs of dynamically generated banners, especially from large banner
# farms, often don't use the well-known image file name extensions, which
# makes it impossible for Privoxy to guess the file type just by looking
# at the URL.
# You can use the +block-as-image alias defined above for these cases.
# Note that objects which match this rule but then turn out NOT to be an
# image are typically rendered as a "broken image" icon by the browser.
# Use cautiously.
#
{ +block-as-image }
#.doubleclick.net
#/Realmedia/ads/
#ar.atwola.com/

# Now you noticed that the default configuration breaks Forbes
# Magazine, but you were too lazy to find out which action is the
# culprit, and you were again too lazy to give feedback, so you just
# used the fragile alias on the site, and -- whoa! -- it worked. The
# 'fragile' aliases disables those actions that are most likely to break
# a site. Also, good for testing purposes to see if it is Privoxy that
# is causing the problem or not.
#
{ fragile }
#.forbes.com

# Here are some sites we wish to support, and we will allow their ads
# through.
#
{ allow-ads }
#.sourceforge.net
#.slashdot.org
#.osdn.net
tracking.dpd.de

# user.action is generally the best place to define exceptions and
# additions to the default policies of default.action. Some actions are
# safe to have their default policies set here though. So let's set a
# default policy to have a 'blank' image as opposed to the checkerboard
# pattern for ALL sites. '/' of course matches all URLs.
# patterns:
#
{ +set-image-blocker{blank} }
#/

# Enable the following section (not the regression-test directives)
# to rewrite and redirect click-tracking URLs on news.google.com.
# Disabling JavaScript should work as well and probably works more reliably.
#
# Redirected URL = http://news.google.com/news/url?ct2=us%2F0_0_s_1_1_a&sa=t&usg=AFQjCNHJWPc7ffoSXPSqBRz55jDA0KgxOQ&cid=8797762374160&url=http%3A%2F%2Fonline.wsj.com%2Farticle%2FSB10001424052970204485304576640791304008536.html&ei=YcqeTsymCIjxggf8uQE&rt=HOMEPAGE&vm=STANDARD&bvm=section&did=-6537064229385238098
# Redirect Destination = http://online.wsj.com/article/SB10001424052970204485304576640791304008536.html
# Ignore = Yes
#
#{+fast-redirects{check-decoded-url}}
#news.google.com/news/url.*&url=http.*&

# Enable the following section (not the regression-test directives)
# to block various Facebook "like" and similar tracking URLs.  At the
# time this section was added it was reported to not break Facebook
# itself but this may have changed by the time you read this. This URL
# list is probably incomplete and if you don't have an account anyway,
# you may prefer to block the whole domain.
#
# Blocked URL = http://www.facebook.com/plugins/likebox.php?href=http%3A%2F%2Ffacebook.com%2Farstechnica&width=300&colorscheme=light&show_faces=false&stream=false&header=false&height=62&border_color=%23FFFFFF
# Ignore = Yes
# Blocked URL = http://www.facebook.com/plugins/activity.php?site=arstechnica.com&width=300&height=370&header=false&colorscheme=light&recommendations=false&border_color=%23FFFFFF
# Ignore = Yes
# Blocked URL = http://www.facebook.com/plugins/fan.php?api_key=368513495882&connections=10&height=250&id=8304333127&locale=en_US&sdk=joey&stream=false&width=377
# Ignore = Yes
# Blocked URL = http://www.facebook.com/plugins/like.php?api_key=368513495882&channel_url=http%3A%2F%2Fstatic.ak.fbcdn.net%2Fconnect%2Fxd_proxy.php%3Fversion%3D3%23cb%3Df13997452c%26origin%3Dhttp%253A%252F%252Fonline.wsj.com%252Ff1b037e354%26relation%3Dparent.parent%26transport%3Dpostmessage&extended_social_context=false&href=http%3A%2F%2Fonline.wsj.com%2Farticle%2FSB10001424052970204485304576640791304008536.html&layout=button_count&locale=en_US&node_type=link&ref=wsj_share_FB&sdk=joey&send=false&show_faces=false&width=90
# Ignore = Yes
#
#{+block{Facebook "like" and similar tracking URLs.}}
#www.facebook.com/(extern|plugins)/(login_status|like(box)?|activity|fan)\.php
{+block{Various domains as known adservers}}
.vgwort.de
.nuggad.net
.adcell.de
.adform.net
.adition.com
.chartbeat.com
.parsely.com
.parsely.com
.meetrics.net
.ads.linkedin.com
.clickcease.com
.crashlytics.com
.doubleclick.net
.facebook.com
.facebook.net
.google-analytics.com
.prophet.heise.de
.stats.wp.com
.yandex.ru
.userreplay.net
.script.hotjar.com
.clicktale.net
.smartlook.com
.decibelinsight.net
.quantummetric.com
.inspectlet.com
.mouseflow.com
.logrocket.com
.salemove.com
.parse.ly
.cxense.com
.hubspot.com
.computecmedia.de
.d10lpsik1i8c69.cloudfront.net
.insights.hotjar.com/api
fullstory.com/s/fs.js
ws.sessioncam.com/Record/record.asmx
d2oh4tlt9mrke9.cloudfront.net/Record/js/sessioncam.recorder.js
c.spiegel.de/nm_trck.gif
c.spiegel.de/nm_empty.gif
script.ioam.de
adservice.google.com
stations.cursetech.com
.google-analytics.com
.googletagmanager.com
.twitch.tv
.adnxs.com
.sdad.guru
www.sdad.guru
.i10c.net
metric-agent.i10c.net
x.bidswitch.net
tag.clrstm.com
sync.mathtag.com
.openx.net
eu-u.openx.net
us-u.openx.net
# end
`

const privoxyUserConfig = `
# ********************************************************************
#
#  File        :  $Source: /cvsroot/ijbswa/current/user.filter,v $
#
#  $Id: user.filter,v 1.3 2008/05/21 20:17:03 fabiankeil Exp $
#
#  Purpose     :  Rules to process the content of web pages
#
#  Copyright   :  Written by and Copyright (C) 2006-2008 the
#                 Privoxy team. http://www.privoxy.org/
#
# We value your feedback. However, to provide you with the best support,
# please note:
#
#  * Use the support forum to get help:
#    http://sourceforge.net/tracker/?group_id=11118&atid=211118
#  * Submit bugs only thru our bug forum:
#    http://sourceforge.net/tracker/?group_id=11118&atid=111118
#    Make sure that the bug has not already been submitted. Please try
#    to verify that it is a Privoxy bug, and not a browser or site
#    bug first. If you are using your own custom configuration, please
#    try the stock configs to see if the problem is a configuration
#    related bug. And if not using the latest development snapshot,
#    please try the latest one. Or even better, CVS sources.
#  * Submit feature requests only thru our feature request forum:
#    http://sourceforge.net/tracker/?atid=361118&group_id=11118&func=browse
#
# For any other issues, feel free to use the mailing lists:
# http://sourceforge.net/mail/?group_id=11118
#
# Anyone interested in actively participating in development and related
# discussions can join the appropriate mailing list here:
# http://sourceforge.net/mail/?group_id=11118. Archives are available
# here too.
#
#################################################################################
#
# Syntax:
#
# Generally filters start with a line like "FILTER: name description".
# They are then referrable from the actionsfile with +filter{name}
#
# FILTER marks a filter as content filter, other filter
# types are CLIENT-HEADER-FILTER, CLIENT-HEADER-TAGGER,
# SERVER-HEADER-FILTER and SERVER-HEADER-TAGGER.
#
# Inside the filters, write one Perl-Style substitution (job) per line.
# Jobs that precede the first FILTER: line are ignored.
#
# For Details see the pcrs manpage contained in this distribution.
# (and the perlre, perlop and pcre manpages)
#
# Note that you are free to choose the delimiter as you see fit.
#
# Note2: In addition to the Perl options gimsx, the following nonstandard
# options are supported:
#
# 'U' turns the default to ungreedy matching.  Add ? to quantifiers to
#     switch back to greedy.
#
# 'T' (trivial) prevents parsing for backreferences in the substitute.
#     Use if you want to include text like '$&' in your substitute without
#     quoting.
#
# 'D' (Dynamic) allows the use of variables. Supported variables are:
#     $host, $origin (the IP address the request came from), $path and $url.
#
#     Note that '$' is a bad choice as delimiter for dynamic filters as you
#     might end up with unintended variables if you use a variable name
#     directly after the delimiter. Variables will be resolved without
#     escaping anything, therefore you also have to be careful not to chose
#     delimiters that appear in the replacement text. For example '<' should
#     be save, while '?' will sooner or later cause conflicts with $url.
#
#################################################################################
# end
`

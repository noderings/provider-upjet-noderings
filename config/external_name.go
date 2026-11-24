// SPDX-FileCopyrightText: 2026 NodeRings
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// ExternalNameConfigs maps every Terraform resource to how its external name
// (the crossplane.io/external-name annotation) is derived. Every NodeRings
// resource uses the Terraform resource ID as its external name, which is also
// the ID `terraform import` accepts:
//
//	noderings_dns_record           <zone id>/<name>/<type>
//	noderings_game_server_backup   <game server id>/<backup id>
//	noderings_role_assignment      <principal kind>/<principal id>
//	everything else                the NodeRings ID
//
// Only resources listed here are generated.
var ExternalNameConfigs = map[string]config.ExternalName{
	// compute
	tfVirtualMachine:                          config.IdentifierFromProvider,
	"noderings_virtual_machine_backup":        config.IdentifierFromProvider,
	"noderings_virtual_machine_snapshot":      config.IdentifierFromProvider,
	"noderings_virtual_machine_additional_ip": config.IdentifierFromProvider,

	// dns
	tfDNSZone:                   config.IdentifierFromProvider,
	"noderings_dns_record":      config.IdentifierFromProvider,
	"noderings_dns_geo_record":  config.IdentifierFromProvider,
	"noderings_dns_zone_dnssec": config.IdentifierFromProvider,

	// firewall
	tfRuleset:                               config.IdentifierFromProvider,
	"noderings_firewall_ruleset_rules":      config.IdentifierFromProvider,
	"noderings_firewall_ruleset_attachment": config.IdentifierFromProvider,
	"noderings_firewall_address_group":      config.IdentifierFromProvider,

	// gameserver
	tfGameServer:                       config.IdentifierFromProvider,
	"noderings_game_server_allocation": config.IdentifierFromProvider,
	"noderings_game_server_backup":     config.IdentifierFromProvider,
	"noderings_game_server_database":   config.IdentifierFromProvider,
	"noderings_game_server_schedule":   config.IdentifierFromProvider,

	// iam
	tfSSHKey:                         config.IdentifierFromProvider,
	tfRole:                           config.IdentifierFromProvider,
	"noderings_role_assignment":      config.IdentifierFromProvider,
	tfGroup:                          config.IdentifierFromProvider,
	"noderings_group_membership":     config.IdentifierFromProvider,
	"noderings_secret":               config.IdentifierFromProvider,
	"noderings_environment_variable": config.IdentifierFromProvider,
}

// ExternalNameConfigurations applies all external name configurations listed
// in ExternalNameConfigs.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually, as anchored regular expressions.
func ExternalNameConfigured() []string {
	l := make([]string, 0, len(ExternalNameConfigs))
	for name := range ExternalNameConfigs {
		l = append(l, name+"$")
	}
	return l
}

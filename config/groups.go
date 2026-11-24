// SPDX-FileCopyrightText: 2026 NodeRings
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Terraform resources that other resources reference.
const (
	tfVirtualMachine = "noderings_virtual_machine"
	tfDNSZone        = "noderings_dns_zone"
	tfGameServer     = "noderings_game_server"
	tfSSHKey         = "noderings_ssh_key"
	tfRuleset        = "noderings_firewall_ruleset"
	tfRole           = "noderings_role"
	tfGroup          = "noderings_group"
)

// API groups below the root group.
const (
	groupCompute    = "compute"
	groupDNS        = "dns"
	groupFirewall   = "firewall"
	groupGameServer = "gameserver"
	groupIAM        = "iam"
)

// groupKind is the Kubernetes API group (below the root group) and kind of a
// Terraform resource.
type groupKind struct {
	Group string
	Kind  string
}

// GroupKinds names every resource explicitly. Upjet's default would split
// noderings_virtual_machine into group "virtual" and kind "Machine".
var GroupKinds = map[string]groupKind{
	tfVirtualMachine:                          {groupCompute, "VirtualMachine"},
	"noderings_virtual_machine_backup":        {groupCompute, "VirtualMachineBackup"},
	"noderings_virtual_machine_snapshot":      {groupCompute, "VirtualMachineSnapshot"},
	"noderings_virtual_machine_additional_ip": {groupCompute, "VirtualMachineAdditionalIP"},

	tfDNSZone:                   {groupDNS, "Zone"},
	"noderings_dns_record":      {groupDNS, "Record"},
	"noderings_dns_geo_record":  {groupDNS, "GeoRecord"},
	"noderings_dns_zone_dnssec": {groupDNS, "ZoneDNSSEC"},

	tfRuleset:                               {groupFirewall, "Ruleset"},
	"noderings_firewall_ruleset_rules":      {groupFirewall, "RulesetRules"},
	"noderings_firewall_ruleset_attachment": {groupFirewall, "RulesetAttachment"},
	"noderings_firewall_address_group":      {groupFirewall, "AddressGroup"},

	tfGameServer:                       {groupGameServer, "GameServer"},
	"noderings_game_server_allocation": {groupGameServer, "Allocation"},
	"noderings_game_server_backup":     {groupGameServer, "Backup"},
	"noderings_game_server_database":   {groupGameServer, "Database"},
	"noderings_game_server_schedule":   {groupGameServer, "Schedule"},

	tfSSHKey:                         {groupIAM, "SSHKey"},
	tfRole:                           {groupIAM, "Role"},
	"noderings_role_assignment":      {groupIAM, "RoleAssignment"},
	tfGroup:                          {groupIAM, "Group"},
	"noderings_group_membership":     {groupIAM, "GroupMembership"},
	"noderings_secret":               {groupIAM, "Secret"},
	"noderings_environment_variable": {groupIAM, "EnvironmentVariable"},
}

// references lets a field point at another managed resource by name
// (fooRef / fooSelector) instead of a hard-coded ID.
var references = map[string]map[string]string{
	tfVirtualMachine:                          {"ssh_key_ids": tfSSHKey},
	"noderings_virtual_machine_backup":        {"virtual_machine_id": tfVirtualMachine},
	"noderings_virtual_machine_snapshot":      {"virtual_machine_id": tfVirtualMachine},
	"noderings_virtual_machine_additional_ip": {"virtual_machine_id": tfVirtualMachine},

	"noderings_dns_record":      {"zone_id": tfDNSZone},
	"noderings_dns_geo_record":  {"zone_id": tfDNSZone},
	"noderings_dns_zone_dnssec": {"zone_id": tfDNSZone},

	"noderings_firewall_ruleset_rules": {"ruleset_id": tfRuleset},
	"noderings_firewall_ruleset_attachment": {
		"ruleset_id":         tfRuleset,
		"virtual_machine_id": tfVirtualMachine,
	},

	"noderings_game_server_allocation": {"game_server_id": tfGameServer},
	"noderings_game_server_backup":     {"game_server_id": tfGameServer},
	"noderings_game_server_database":   {"game_server_id": tfGameServer},
	"noderings_game_server_schedule":   {"game_server_id": tfGameServer},

	"noderings_role_assignment": {
		"role_id":  tfRole,
		"group_id": tfGroup,
	},
	"noderings_group_membership": {"group_id": tfGroup},
}

// Configure applies the group, kind and references of every resource. It is
// shared by the cluster-scoped and the namespaced provider.
func Configure(p *ujconfig.Provider) {
	for name, gk := range GroupKinds {
		refs := references[name]
		p.AddResourceConfigurator(name, func(r *ujconfig.Resource) {
			r.ShortGroup = gk.Group
			r.Kind = gk.Kind
			for field, target := range refs {
				r.References[field] = ujconfig.Reference{TerraformName: target}
			}
		})
	}
}

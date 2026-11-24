// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	virtualmachine "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/compute/virtualmachine"
	virtualmachineadditionalip "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/compute/virtualmachineadditionalip"
	virtualmachinebackup "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/compute/virtualmachinebackup"
	virtualmachinesnapshot "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/compute/virtualmachinesnapshot"
	georecord "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/dns/georecord"
	record "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/dns/record"
	zone "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/dns/zone"
	zonednssec "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/dns/zonednssec"
	addressgroup "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/firewall/addressgroup"
	ruleset "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/firewall/ruleset"
	rulesetattachment "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/firewall/rulesetattachment"
	rulesetrules "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/firewall/rulesetrules"
	allocation "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/gameserver/allocation"
	backup "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/gameserver/backup"
	database "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/gameserver/database"
	gameserver "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/gameserver/gameserver"
	schedule "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/gameserver/schedule"
	environmentvariable "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/iam/environmentvariable"
	group "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/iam/group"
	groupmembership "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/iam/groupmembership"
	role "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/iam/role"
	roleassignment "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/iam/roleassignment"
	secret "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/iam/secret"
	sshkey "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/iam/sshkey"
	providerconfig "github.com/noderings/provider-upjet-noderings/internal/controller/namespaced/providerconfig"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		virtualmachine.Setup,
		virtualmachineadditionalip.Setup,
		virtualmachinebackup.Setup,
		virtualmachinesnapshot.Setup,
		georecord.Setup,
		record.Setup,
		zone.Setup,
		zonednssec.Setup,
		addressgroup.Setup,
		ruleset.Setup,
		rulesetattachment.Setup,
		rulesetrules.Setup,
		allocation.Setup,
		backup.Setup,
		database.Setup,
		gameserver.Setup,
		schedule.Setup,
		environmentvariable.Setup,
		group.Setup,
		groupmembership.Setup,
		role.Setup,
		roleassignment.Setup,
		secret.Setup,
		sshkey.Setup,
		providerconfig.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		virtualmachine.SetupGated,
		virtualmachineadditionalip.SetupGated,
		virtualmachinebackup.SetupGated,
		virtualmachinesnapshot.SetupGated,
		georecord.SetupGated,
		record.SetupGated,
		zone.SetupGated,
		zonednssec.SetupGated,
		addressgroup.SetupGated,
		ruleset.SetupGated,
		rulesetattachment.SetupGated,
		rulesetrules.SetupGated,
		allocation.SetupGated,
		backup.SetupGated,
		database.SetupGated,
		gameserver.SetupGated,
		schedule.SetupGated,
		environmentvariable.SetupGated,
		group.SetupGated,
		groupmembership.SetupGated,
		role.SetupGated,
		roleassignment.SetupGated,
		secret.SetupGated,
		sshkey.SetupGated,
		providerconfig.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		virtualmachine.SetupWebhookWithManager,
		virtualmachineadditionalip.SetupWebhookWithManager,
		virtualmachinebackup.SetupWebhookWithManager,
		virtualmachinesnapshot.SetupWebhookWithManager,
		georecord.SetupWebhookWithManager,
		record.SetupWebhookWithManager,
		zone.SetupWebhookWithManager,
		zonednssec.SetupWebhookWithManager,
		addressgroup.SetupWebhookWithManager,
		ruleset.SetupWebhookWithManager,
		rulesetattachment.SetupWebhookWithManager,
		rulesetrules.SetupWebhookWithManager,
		allocation.SetupWebhookWithManager,
		backup.SetupWebhookWithManager,
		database.SetupWebhookWithManager,
		gameserver.SetupWebhookWithManager,
		schedule.SetupWebhookWithManager,
		environmentvariable.SetupWebhookWithManager,
		group.SetupWebhookWithManager,
		groupmembership.SetupWebhookWithManager,
		role.SetupWebhookWithManager,
		roleassignment.SetupWebhookWithManager,
		secret.SetupWebhookWithManager,
		sshkey.SetupWebhookWithManager,
		providerconfig.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}

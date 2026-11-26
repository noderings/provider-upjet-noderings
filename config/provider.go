// SPDX-FileCopyrightText: 2026 NodeRings
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	// Embed the provider schema and metadata.
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/crossplane/upjet/v2/pkg/registry/reference"
	"github.com/noderings/terraform-provider-noderings/shim"

	"github.com/noderings/provider-upjet-noderings/internal/version"
)

const (
	resourcePrefix = "noderings"
	modulePath     = "github.com/noderings/provider-upjet-noderings"

	// RootGroup is the API group root of cluster-scoped resources, for
	// example dns.noderings.com.
	RootGroup = "noderings.com"
	// RootGroupNamespaced is the API group root of namespaced resources, for
	// example dns.m.noderings.com.
	RootGroupNamespaced = "m.noderings.com"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns the configuration of the cluster-scoped resources.
func GetProvider() *ujconfig.Provider {
	return newProvider(RootGroup)
}

// GetProviderNamespaced returns the configuration of the namespaced resources.
func GetProviderNamespaced() *ujconfig.Provider {
	return newProvider(RootGroupNamespaced,
		ujconfig.WithExampleManifestConfiguration(ujconfig.ExampleManifestConfiguration{
			ManagedResourceNamespace: "default",
		}))
}

func newProvider(rootGroup string, opts ...ujconfig.ProviderOption) *ujconfig.Provider {
	opts = append([]ujconfig.ProviderOption{
		ujconfig.WithRootGroup(rootGroup),
		ujconfig.WithShortName("noderings"),
		// Every resource is served by the Terraform provider compiled into
		// this binary (terraform-plugin-framework); none runs through the
		// Terraform CLI.
		ujconfig.WithIncludeList(nil),
		ujconfig.WithTerraformPluginFrameworkIncludeList(ExternalNameConfigured()),
		ujconfig.WithTerraformPluginFrameworkProvider(shim.NewProvider(version.Version)),
		ujconfig.WithReferenceInjectors([]ujconfig.ReferenceInjector{reference.NewInjector(modulePath)}),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithDefaultResourceOptions(ExternalNameConfigurations()),
	}, opts...)

	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata), opts...)
	Configure(pc)
	pc.ConfigureResources()
	return pc
}

// SPDX-FileCopyrightText: 2026 NodeRings
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"strings"
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// upstreamResources returns every resource in the Terraform provider schema.
func upstreamResources(t *testing.T) []string {
	t.Helper()
	var s struct {
		ProviderSchemas map[string]struct {
			ResourceSchemas map[string]json.RawMessage `json:"resource_schemas"`
		} `json:"provider_schemas"`
	}
	if err := json.Unmarshal([]byte(providerSchema), &s); err != nil {
		t.Fatalf("cannot parse schema.json: %v", err)
	}
	var names []string
	for _, p := range s.ProviderSchemas {
		for name := range p.ResourceSchemas {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		t.Fatal("schema.json has no resources")
	}
	return names
}

// A new resource in the Terraform provider must be configured here, or it
// would be silently missing from the Crossplane provider.
func TestEveryUpstreamResourceIsConfigured(t *testing.T) {
	for _, name := range upstreamResources(t) {
		if _, ok := ExternalNameConfigs[name]; !ok {
			t.Errorf("%s has no external name in ExternalNameConfigs", name)
		}
		if _, ok := GroupKinds[name]; !ok {
			t.Errorf("%s has no group and kind in GroupKinds", name)
		}
	}
	for name := range GroupKinds {
		if _, ok := ExternalNameConfigs[name]; !ok {
			t.Errorf("%s is in GroupKinds but not in ExternalNameConfigs", name)
		}
	}
}

func TestProvidersGenerateEveryResource(t *testing.T) {
	want := len(upstreamResources(t))
	for scope, p := range map[string]*ujconfig.Provider{
		RootGroup:           GetProvider(),
		RootGroupNamespaced: GetProviderNamespaced(),
	} {
		if got := len(p.Resources); got != want {
			t.Errorf("%s: %d resources, want %d", scope, got, want)
		}
		if p.RootGroup != scope {
			t.Errorf("root group %q, want %q", p.RootGroup, scope)
		}
		for name, r := range p.Resources {
			gk := GroupKinds[name]
			if r.ShortGroup != gk.Group || r.Kind != gk.Kind {
				t.Errorf("%s: %s/%s, want %s/%s", name, r.ShortGroup, r.Kind, gk.Group, gk.Kind)
			}
			if r.TerraformPluginFrameworkResource == nil {
				t.Errorf("%s is not served by the in-process Terraform provider", name)
			}
		}
	}
}

func TestReferencesPointAtConfiguredResources(t *testing.T) {
	p := GetProviderNamespaced()
	for name, refs := range references {
		r, ok := p.Resources[name]
		if !ok {
			t.Errorf("references for unknown resource %s", name)
			continue
		}
		for field, target := range refs {
			if _, ok := p.Resources[target]; !ok {
				t.Errorf("%s.%s references unknown resource %s", name, field, target)
			}
			if _, ok := r.References[field]; !ok {
				t.Errorf("%s.%s reference was not applied", name, field)
			}
			if !strings.HasSuffix(field, "_id") && !strings.HasSuffix(field, "_ids") {
				t.Errorf("%s.%s: references should be on ID fields", name, field)
			}
		}
	}
}

// SPDX-FileCopyrightText: 2026 NodeRings
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"reflect"
	"strings"
	"testing"
)

func TestProviderConfiguration(t *testing.T) {
	cases := map[string]struct {
		in      string
		want    map[string]any
		wantErr string
	}{
		"token only": {
			in:   `{"token":"t"}`,
			want: map[string]any{"token": "t", "skip_organization_check": true},
		},
		"all keys": {
			in: `{"token":"t","organization_id":"o","endpoint":"https://api.example","max_retries":3,"requests_per_second":1.5,"http_timeout":"30s"}`,
			want: map[string]any{
				"token": "t", "organization_id": "o", "endpoint": "https://api.example",
				"max_retries": int64(3), "requests_per_second": 1.5, "http_timeout": "30s",
				"skip_organization_check": true,
			},
		},
		"unknown keys are ignored": {
			in:   `{"token":"t","color":"blue"}`,
			want: map[string]any{"token": "t", "skip_organization_check": true},
		},
		"missing token":   {in: `{"organization_id":"o"}`, wantErr: "no \"token\""},
		"empty token":     {in: `{"token":""}`, wantErr: "no \"token\""},
		"not json":        {in: `token=abc`, wantErr: "cannot unmarshal"},
		"wrong type":      {in: `{"token":"t","max_retries":"3"}`, wantErr: "must be a number"},
		"string as other": {in: `{"token":"t","endpoint":5}`, wantErr: "must be a string"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ProviderConfiguration([]byte(tc.in))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error %v, want one containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestUserAgent(t *testing.T) {
	if ua := UserAgent(); !strings.HasPrefix(ua, "crossplane-provider-noderings/") {
		t.Errorf("user agent %q", ua)
	}
}

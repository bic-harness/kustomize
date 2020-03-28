// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package annotations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	filtertest_test "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

var annosFs = builtinconfig.MakeDefaultConfig().CommonAnnotations

func TestAnnotations_Filter(t *testing.T) {
	testCases := map[string]struct {
		input          string
		expectedOutput string
		filter         Filter
		fsslice        types.FsSlice
	}{
		"add": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
    auto: ford
    bean: cannellini
    clown: emmett kelley
    dragon: smaug
`,
			filter: Filter{Annotations: annoMap{
				"clown":  "emmett kelley",
				"auto":   "ford",
				"dragon": "smaug",
				"bean":   "cannellini",
			}},
		},
		"update": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: batman
    fiend: riddler
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    hero: superman
    fiend: luthor
    bean: cannellini
    clown: emmett kelley
`,
			filter: Filter{Annotations: annoMap{
				"clown": "emmett kelley",
				"hero":  "superman",
				"fiend": "luthor",
				"bean":  "cannellini",
			}},
		},
		"data-fieldspecs": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
  annotations:
    sleater: kinney
a:
  b:
    sleater: kinney
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
  annotations:
    sleater: kinney
a:
  b:
    sleater: kinney
`,
			filter: Filter{Annotations: annoMap{
				"sleater": "kinney",
			}},
			fsslice: []types.FieldSpec{
				{
					Path:               "a/b",
					CreateIfNotPresent: true,
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			filter := tc.filter
			filter.FsSlice = append(annosFs, tc.fsslice...)
			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput),
				strings.TrimSpace(filtertest_test.RunFilter(t, tc.input, filter))) {
				t.FailNow()
			}
		})
	}
}

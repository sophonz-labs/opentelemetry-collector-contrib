// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

// OCB has no stanza operator section. Register the Sophonz parser alongside
// its generated components so regeneration preserves the registration.
import _ "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonzgrok"

// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"time"
	"strings"
)

// Configuration of application
type Configuration struct {
	// Tenant represent tenant of given vault
	Tenant string
	// LakeHostname represent hostname of openbank lake service
	LakeHostname string
	// RootStorage gives where to store journals
	RootStorage string
	// LogLevel ignorecase log level
	LogLevel string
	// MetricsRefreshRate represents interval in which in memory metrics should be
	// persisted to disk
	MetricsRefreshRate time.Duration
	// MetricsOutput represents output file for metrics persistence
	MetricsOutput string
	// SnapshotSaturationTreshold represents number of events needed in account to
	// consider account snapshot in given version to be saturated
	SnapshotSaturationTreshold int
}

// GetConfig loads application configuration
func GetConfig() Configuration {
	return Configuration{
		Tenant:                     envString("VAULT_TENANT", ""),
		LakeHostname:               envString("VAULT_LAKE_HOSTNAME", "lake"),
		RootStorage:                envString("VAULT_STORAGE", "/data") + "/" + "t_" + envString("VAULT_TENANT", ""),
		LogLevel:                   strings.ToUpper(envString("VAULT_LOG_LEVEL", "DEBUG")),
		MetricsRefreshRate:         envDuration("VAULT_METRICS_REFRESHRATE", time.Second),
		MetricsOutput:              envFilename("VAULT_METRICS_OUTPUT", "/tmp/vault-unit-metrics"),
		SnapshotSaturationTreshold: envInteger("VAULT_SNAPSHOT_SATURATION_TRESHOLD", 100),
	}
}

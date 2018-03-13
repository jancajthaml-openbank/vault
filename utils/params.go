// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package utils

import "time"

// RunParams is a structure of application parameters
type RunParams struct {
	// Tenant represent tenant of given vault
	Tenant string
	// LakeHostname represent hostname of openbank lake service
	LakeHostname string
	// RootStorage gives where to store journals
	RootStorage string
	// Log represents log output
	Log string
	// LogLevel ignorecase log level
	LogLevel string
	// JournalSaturation represents number of events needed in account to consider
	// account snapshot in given version to be saturated
	JournalSaturation int
	// SnapshotScanInterval represents backoff between scan for saturated
	// snapshots
	SnapshotScanInterval time.Duration
	// MetricsRefreshRate represents interval in which in memory metrics should
	// be persisted to disk
	MetricsRefreshRate time.Duration
	// MetricsOutput represents output file for metrics persistence
	MetricsOutput string
}

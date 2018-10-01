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

import (
	"strconv"
)

// AccountsPath returns filepath for accounts
func AccountsPath(root string) string {
	return root + "/account"
}

// EventsPath returns filepath for events
func EventsPath(root, name string) string {
	return root + "/account/" + name + "/events"
}

// SnapshotsPath returns filepath for snapshots
func SnapshotsPath(root, name string) string {
	return root + "/account/" + name + "/snapshot"
}

// EventPath returns filepath for given event
func EventPath(root, name string, version int) string {
	value := strconv.Itoa(version)
	return root + "/account/" + name + "/events/" + "0000000000"[0:10-len(value)] + value
}

// SnapshotPath returns filepath for given snapshot
func SnapshotPath(root, name string, version int) string {
	value := strconv.Itoa(version)
	return root + "/account/" + name + "/snapshot/" + "0000000000"[0:10-len(value)] + value
}

// MetadataPath returns filepath for given metadata
func MetadataPath(root, name string) string {
	return root + "/account/" + name + "/meta"
}

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
	"math"
	"strings"

	"github.com/jancajthaml-openbank/vault/model"

	money "gopkg.in/inf.v0"
)

func snapshotExists(params RunParams, name string) bool {
	return Exists(SnapshotsPath(params, name))
}

func LoadMetadata(params RunParams, name string) *model.Account {
	metaPath := MetadataPath(params, name)

	ok, data := ReadFileFully(metaPath)

	if !ok {
		return nil
	}

	result := new(model.Account)
	result.DeserializeFromStorage(data)

	return result
}

func LoadSnapshot(params RunParams, name string) *model.Snapshot {
	allPath := SnapshotsPath(params, name)
	snapshots := ListDirectory(allPath, false)

	if len(snapshots) == 0 {
		return nil
	}

	ok, data := ReadFileFully(allPath + "/" + snapshots[0])

	if !ok {
		return nil
	}

	result := new(model.Snapshot)
	result.DeserializeFromStorage(data)

	events := ListDirectory(EventPath(params, name, result.Version), false)

	for _, event := range events {

		s := strings.Split(event, "_")
		kind, amountString, transaction := s[0], s[1], s[2]

		switch kind {

		case model.EventPromise:
			amount, _ := new(money.Dec).SetString(amountString)

			result.PromiseBuffer.Add(transaction)
			result.Promised = new(money.Dec).Add(result.Promised, amount)

		case model.EventCommit:
			amount, _ := new(money.Dec).SetString(amountString)

			result.PromiseBuffer.Remove(transaction)
			result.Promised = new(money.Dec).Sub(result.Promised, amount)
			result.Balance = new(money.Dec).Add(result.Balance, amount)

		case model.EventRollback:
			amount, _ := new(money.Dec).SetString(amountString)

			result.PromiseBuffer.Remove(transaction)
			result.Promised = new(money.Dec).Sub(result.Promised, amount)
		}
	}

	return result
}

func CreateMetadata(params RunParams, name string, currency string, isBalanceCheck bool) *model.Account {
	return StoreMetadata(params, &model.Account{
		AccountName:    name,
		Currency:       currency,
		IsBalanceCheck: isBalanceCheck,
	})
}

// FIXME benchmark this function
func CreateSnapshot(params RunParams, name string) *model.Snapshot {
	return StoreSnapshot(params, name, &model.Snapshot{
		Balance:       new(money.Dec),
		Promised:      new(money.Dec),
		PromiseBuffer: model.NewTransactionSet(),
		Version:       0,
	})
}

func UpdateSnapshot(params RunParams, name string, entity *model.Snapshot) *model.Snapshot {
	if entity.Version == math.MaxInt32 {
		return entity
	}

	return StoreSnapshot(params, name, &model.Snapshot{
		Balance:       entity.Balance,
		Promised:      entity.Promised,
		PromiseBuffer: entity.PromiseBuffer,
		Version:       entity.Version + 1,
	})
}

func StoreSnapshot(params RunParams, name string, entity *model.Snapshot) *model.Snapshot {
	data := entity.SerializeForStorage()
	path := SnapshotPath(params, name, entity.Version)

	if !WriteFile(path, data) {
		return nil
	}

	return entity
}

func StoreMetadata(params RunParams, entity *model.Account) *model.Account {
	data := entity.SerializeForStorage()
	path := MetadataPath(params, entity.AccountName)

	if !WriteFile(path, data) {
		return nil
	}

	return entity
}

func PersistPromise(params RunParams, name string, version int, amount *money.Dec, transaction string) bool {
	event := model.EventPromise + "_" + amount.String() + "_" + transaction
	fullPath := EventPath(params, name, version) + "/" + event
	return TouchFile(fullPath)
}

func PersistCommit(params RunParams, name string, version int, amount *money.Dec, transaction string) bool {
	event := model.EventCommit + "_" + amount.String() + "_" + transaction
	fullPath := EventPath(params, name, version) + "/" + event
	return TouchFile(fullPath)
}

func PersistRollback(params RunParams, name string, version int, amount *money.Dec, transaction string) bool {
	event := model.EventRollback + "_" + amount.String() + "_" + transaction
	fullPath := EventPath(params, name, version) + "/" + event
	return TouchFile(fullPath)
}

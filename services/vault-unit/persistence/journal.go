// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package persistence

import (
	"math"
	"strconv"
	"strings"

	"github.com/jancajthaml-openbank/vault-unit/model"
	"github.com/jancajthaml-openbank/vault-unit/utils"

	localfs "github.com/jancajthaml-openbank/local-fs"
	money "gopkg.in/inf.v0"
)

// LoadAccount rehydrates account entity state from storage
func LoadAccount(storage *localfs.Storage, name string) *model.Account {
	allPath := utils.SnapshotsPath(name)

	snapshots, err := storage.ListDirectory(allPath, false)
	if err != nil || len(snapshots) == 0 {
		return nil
	}

	data, err := storage.ReadFileFully(allPath + "/" + snapshots[0])
	if err != nil {
		return nil
	}

	result := new(model.Account)

	version, err := strconv.Atoi(snapshots[0])
	if err != nil {
		return nil
	}

	result.Version = version
	result.AccountName = name
	result.Deserialise(data)

	events, err := storage.ListDirectory(utils.EventPath(name, result.Version), false)
	if err == nil {
		for _, event := range events {
			s := strings.SplitN(event, "_", 3)
			kind, amountString, transaction := s[0], s[1], s[2]

			amount, _ := new(money.Dec).SetString(amountString)

			switch kind {

			case model.EventPromise:
				result.PromiseBuffer.Add(transaction)
				result.Promised = new(money.Dec).Add(result.Promised, amount)

			case model.EventCommit:
				result.PromiseBuffer.Remove(transaction)
				result.Promised = new(money.Dec).Sub(result.Promised, amount)
				result.Balance = new(money.Dec).Add(result.Balance, amount)

			case model.EventRollback:
				result.PromiseBuffer.Remove(transaction)
				result.Promised = new(money.Dec).Sub(result.Promised, amount)

			}
		}
	}

	return result
}

// CreateAccount persist account entity state to storage
func CreateAccount(storage *localfs.Storage, name, currency string, isBalanceCheck bool) *model.Account {
	return PersistAccount(storage, name, &model.Account{
		Balance:        new(money.Dec),
		Promised:       new(money.Dec),
		PromiseBuffer:  model.NewTransactionSet(),
		Version:        0,
		AccountName:    name,
		Currency:       currency,
		IsBalanceCheck: isBalanceCheck,
	})
}

// UpdateAccount persist account entity state with incremented version
func UpdateAccount(storage *localfs.Storage, name string, entity *model.Account) *model.Account {
	if entity.Version == math.MaxInt32 {
		return entity
	}

	return PersistAccount(storage, name, &model.Account{
		Balance:        entity.Balance,
		Promised:       entity.Promised,
		PromiseBuffer:  entity.PromiseBuffer,
		Version:        entity.Version + 1,
		Currency:       entity.Currency,
		AccountName:    entity.AccountName,
		IsBalanceCheck: entity.IsBalanceCheck,
	})
}

// PersistAccount persist account entity state to storage
func PersistAccount(storage *localfs.Storage, name string, entity *model.Account) *model.Account {
	data := entity.Serialise()
	path := utils.SnapshotPath(name, entity.Version)

	if storage.WriteFile(path, data) != nil {
		return nil
	}

	return entity
}

// PersistPromise persists promise event
func PersistPromise(storage *localfs.Storage, name string, version int, amount *money.Dec, transaction string) bool {
	event := model.EventPromise + "_" + amount.String() + "_" + transaction
	fullPath := utils.EventPath(name, version) + "/" + event
	return storage.TouchFile(fullPath) == nil
}

// PersistCommit persists commit event
func PersistCommit(storage *localfs.Storage, name string, version int, amount *money.Dec, transaction string) bool {
	event := model.EventCommit + "_" + amount.String() + "_" + transaction
	fullPath := utils.EventPath(name, version) + "/" + event
	return storage.TouchFile(fullPath) == nil
}

// PersistRollback persists rollback event
func PersistRollback(storage *localfs.Storage, name string, version int, amount *money.Dec, transaction string) bool {
	event := model.EventRollback + "_" + amount.String() + "_" + transaction
	fullPath := utils.EventPath(name, version) + "/" + event
	return storage.TouchFile(fullPath) == nil
}

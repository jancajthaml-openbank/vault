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
func LoadAccount(storage *localfs.PlaintextStorage, name string) (*model.Account, error) {
	allPath := utils.SnapshotsPath(name)

	snapshots, err := storage.ListDirectory(allPath, false)
	if err != nil || len(snapshots) == 0 {
		return nil, err
	}

	data, err := storage.ReadFileFully(allPath + "/" + snapshots[0])
	if err != nil {
		return nil, err
	}

	result := new(model.Account)

	version, err := strconv.ParseInt(snapshots[0], 10, 64)
	if err != nil {
		return nil, err
	}

	result.SnapshotVersion = version
	result.Name = name
	result.Deserialise(data)

	events, err := storage.ListDirectory(utils.EventPath(name, result.SnapshotVersion), false)
	if err == nil {
		for _, event := range events {
			s := strings.SplitN(event, "_", 3)
			kind, amountString, transaction := s[0], s[1], s[2]

			amount, _ := new(money.Dec).SetString(amountString)

			switch kind {

			case EventPromise:
				result.Promises.Add(transaction)
				result.Promised = new(money.Dec).Add(result.Promised, amount)

			case EventCommit:
				result.Promises.Remove(transaction)
				result.Promised = new(money.Dec).Sub(result.Promised, amount)
				result.Balance = new(money.Dec).Add(result.Balance, amount)

			case EventRollback:
				result.Promises.Remove(transaction)
				result.Promised = new(money.Dec).Sub(result.Promised, amount)

			}
		}
	}

	if len(events) > 0 {
		eventData, err := storage.ReadFileFully(utils.EventPath(name, result.SnapshotVersion) + "/" + events[len(events)-1])
		if err != nil {
			return nil, err
		}

		eventCounter, err := strconv.ParseInt(string(eventData), 10, 64)
		if err != nil {
			return nil, err
		}

		result.EventCounter = eventCounter
	} else {
		result.EventCounter = 0
	}

	return result, nil
}

// CreateAccount persist account entity state to storage
func CreateAccount(storage *localfs.PlaintextStorage, name string, format string, currency string, isBalanceCheck bool) (*model.Account, error) {
	entity := model.NewAccount(name)
	entity.Format = strings.ToUpper(format)
	entity.Currency = strings.ToUpper(currency)
	entity.IsBalanceCheck = isBalanceCheck
	data := entity.Serialise()
	path := utils.SnapshotPath(name, entity.SnapshotVersion)
	err := storage.WriteFileExclusive(path, data)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// UpdateAccount persist account entity state with incremented version
func UpdateAccount(storage *localfs.PlaintextStorage, name string, original *model.Account) (*model.Account, error) {
	if original.SnapshotVersion == math.MaxInt32 {
		return original, nil
	}
	entity := &model.Account{
		Balance:         original.Balance,
		Promised:        original.Promised,
		Promises:        original.Promises,
		SnapshotVersion: original.SnapshotVersion + 1,
		EventCounter:    0,
		Currency:        original.Currency,
		Name:            original.Name,
		Format:          original.Format,
		IsBalanceCheck:  original.IsBalanceCheck,
	}
	data := entity.Serialise()
	path := utils.SnapshotPath(name, entity.SnapshotVersion)
	err := storage.WriteFileExclusive(path, data)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// PersistPromise persists promise event
func PersistPromise(storage *localfs.PlaintextStorage, account model.Account, amount *money.Dec, transaction string) error {
	event := EventPromise + "_" + amount.String() + "_" + transaction
	fullPath := utils.EventPath(account.Name, account.SnapshotVersion) + "/" + event
	data := []byte(strconv.FormatInt(account.EventCounter, 10))
	return storage.WriteFileExclusive(fullPath, data)
}

// PersistCommit persists commit event
func PersistCommit(storage *localfs.PlaintextStorage, account model.Account, amount *money.Dec, transaction string) error {
	event := EventCommit + "_" + amount.String() + "_" + transaction
	fullPath := utils.EventPath(account.Name, account.SnapshotVersion) + "/" + event
	data := []byte(strconv.FormatInt(account.EventCounter, 10))
	return storage.WriteFileExclusive(fullPath, data)
}

// PersistRollback persists rollback event
func PersistRollback(storage *localfs.PlaintextStorage, account model.Account, amount *money.Dec, transaction string) error {
	event := EventRollback + "_" + amount.String() + "_" + transaction
	fullPath := utils.EventPath(account.Name, account.SnapshotVersion) + "/" + event
	data := []byte(strconv.FormatInt(account.EventCounter, 10))
	return storage.WriteFileExclusive(fullPath, data)
}

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
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jancajthaml-openbank/vault-unit/model"

	localfs "github.com/jancajthaml-openbank/local-fs"
	money "gopkg.in/inf.v0"
)

// LoadAccount rehydrates account entity state from storage
func LoadAccount(storage localfs.Storage, name string) (*model.Account, error) {
	allPath := SnapshotsPath(name)

	snapshots, err := storage.ListDirectory(allPath, false)
	if err != nil || len(snapshots) == 0 {
		return nil, err
	}

	data, err := storage.ReadFileFully(allPath + "/" + snapshots[0])
	if err != nil {
		return nil, err
	}

	version, err := strconv.ParseInt(snapshots[0], 10, 64)
	if err != nil {
		return nil, err
	}

	result := new(model.Account)
	result.SnapshotVersion = version
	result.Name = name
	result.Deserialize(data)
	result.EventCounter = 0

	events, err := storage.ListDirectory(EventPath(name, result.SnapshotVersion), false)
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

			eventData, err := storage.ReadFileFully(EventPath(name, result.SnapshotVersion) + "/" + event)
			if err != nil {
				return nil, err
			}

			eventCounter, err := strconv.ParseInt(string(eventData), 10, 64)
			if err != nil {
				return nil, err
			}

			if eventCounter > result.EventCounter {
				result.EventCounter = eventCounter
			}
		}
	}

	return result, nil
}

// CreateAccount persist account entity state to storage
func CreateAccount(storage localfs.Storage, name string, format string, currency string, isBalanceCheck bool) (*model.Account, error) {
	entity := model.NewAccount(name)
	entity.Format = strings.ToUpper(format)
	entity.Currency = strings.ToUpper(currency)
	entity.IsBalanceCheck = isBalanceCheck
	data := entity.Serialize()
	path := SnapshotPath(name, entity.SnapshotVersion)
	err := storage.WriteFileExclusive(path, data)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// UpdateAccount persist account entity state with incremented version
func UpdateAccount(storage localfs.Storage, name string, original *model.Account) error {
	if original.SnapshotVersion == math.MaxInt32 {
		return fmt.Errorf("reached maximum snapshot version")
	}
	updated := new(model.Account)
	updated.Balance = original.Balance
	updated.Promised = original.Promised
	updated.Promises = original.Promises
	updated.SnapshotVersion = original.SnapshotVersion + 1
	updated.EventCounter = 0
	updated.Currency = original.Currency
	updated.Name = original.Name
	updated.Format = original.Format
	updated.IsBalanceCheck = original.IsBalanceCheck
	data := updated.Serialize()
	path := SnapshotPath(name, updated.SnapshotVersion)
	err := storage.WriteFileExclusive(path, data)
	if err != nil {
		return err
	}
	return nil
}

// PersistPromise persists promise event
func PersistPromise(storage localfs.Storage, account model.Account, amount *money.Dec, transaction string) error {
	event := EventPromise + "_" + amount.String() + "_" + transaction
	fullPath := EventPath(account.Name, account.SnapshotVersion) + "/" + event
	data := []byte(strconv.FormatInt(account.EventCounter, 10))
	return storage.WriteFileExclusive(fullPath, data)
}

// PersistCommit persists commit event
func PersistCommit(storage localfs.Storage, account model.Account, amount *money.Dec, transaction string) error {
	event := EventCommit + "_" + amount.String() + "_" + transaction
	fullPath := EventPath(account.Name, account.SnapshotVersion) + "/" + event
	data := []byte(strconv.FormatInt(account.EventCounter, 10))
	return storage.WriteFileExclusive(fullPath, data)
}

// PersistRollback persists rollback event
func PersistRollback(storage localfs.Storage, account model.Account, amount *money.Dec, transaction string) error {
	event := EventRollback + "_" + amount.String() + "_" + transaction
	fullPath := EventPath(account.Name, account.SnapshotVersion) + "/" + event
	data := []byte(strconv.FormatInt(account.EventCounter, 10))
	return storage.WriteFileExclusive(fullPath, data)
}

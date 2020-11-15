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

package model

import (
	"bytes"

	money "gopkg.in/inf.v0"
)

// Account represents metadata of account entity
type Account struct {
	Name            string `json:"name"`
	Format          string `json:"format"`
	Currency        string `json:"currency"`
	IsBalanceCheck  bool   `json:"isBalanceCheck"`
	Balance         *money.Dec
	Promised        *money.Dec
	Promises        Promises
	SnapshotVersion int64
	EventCounter    int64
}

// Copy returns copy of Account
func (entity Account) Copy() Account {
	return Account{
		Name:            entity.Name,
		Format:          entity.Format,
		Currency:        entity.Currency,
		IsBalanceCheck:  entity.IsBalanceCheck,
		Balance:         new(money.Dec).Set(entity.Balance),
		Promised:        new(money.Dec).Set(entity.Promised),
		Promises:        entity.Promises.Copy(),
		SnapshotVersion: entity.SnapshotVersion,
		EventCounter:    entity.EventCounter,
	}
}

// NewAccount returns new Account
func NewAccount(name string) Account {
	return Account{
		Name:            name,
		Format:          "???",
		Currency:        "???",
		IsBalanceCheck:  true,
		Balance:         new(money.Dec),
		Promised:        new(money.Dec),
		SnapshotVersion: 0,
		EventCounter:    0,
		Promises:        NewPromises(),
	}
}

// Serialize Account entity to persistable data
func (entity Account) Serialize() []byte {
	var buffer bytes.Buffer

	// [CURRENCY FORMAT_IS-CHECK]
	// [BALANCE]
	// [PROMISED]
	// [...PROMISE-BUFFER]

	switch len(entity.Currency) {
	case 0:
		buffer.WriteString("???")
	case 1:
		buffer.WriteString("??")
		buffer.WriteString(entity.Currency)
	case 2:
		buffer.WriteString("?")
		buffer.WriteString(entity.Currency)
	default:
		buffer.WriteString(entity.Currency[0:3])
	}

	buffer.WriteString(" ")

	buffer.WriteString(entity.Format)
	if entity.IsBalanceCheck {
		buffer.WriteString("_T")
	} else {
		buffer.WriteString("_F")
	}

	buffer.WriteString("\n")

	if entity.Balance == nil {
		buffer.WriteString("0.0")
	} else {
		buffer.WriteString(entity.Balance.String())
	}

	buffer.WriteString("\n")

	if entity.Promised == nil {
		buffer.WriteString("0.0")
	} else {
		buffer.WriteString(entity.Promised.String())
	}

	for _, v := range entity.Promises.Values() {
		buffer.WriteString("\n")
		buffer.WriteString(v)
	}

	return buffer.Bytes()
}

// Deserialize Account entity from persistable data
func (entity *Account) Deserialize(data []byte) {
	if entity == nil {
		return
	}

	entity.Promises = NewPromises()
	entity.Currency = string(data[0:3])

	var (
		i = 4
		j = bytes.IndexByte(data[4:], '\n') + 4
	)

	format := string(data[i:j])

	entity.IsBalanceCheck = (format[len(format)-1:] != "F")
	entity.Format = format[:len(format)-2]

	i = j + 1

	j = bytes.IndexByte(data[i:], '\n')
	if j < 0 {
		if len(data) > 0 {
			entity.Balance, _ = new(money.Dec).SetString(string(data[i]))
		}
		return
	}
	j += i
	entity.Balance, _ = new(money.Dec).SetString(string(data[i:j]))
	i = j + 1

	j = bytes.IndexByte(data[i:], '\n')
	if j < 0 {
		if len(data) > 0 {
			entity.Promised, _ = new(money.Dec).SetString(string(data[i]))
		}
		return
	}
	j += i
	entity.Promised, _ = new(money.Dec).SetString(string(data[i:j]))
	i = j + 1

	for {
		j = bytes.IndexByte(data[i:], '\n')
		if j < 0 {
			if len(data) > 0 {
				entity.Promises.Add(string(data[i:]))
			}
			return
		}
		j += i
		entity.Promises.Add(string(data[i:j]))
		i = j + 1
	}
}

// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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

	"github.com/jancajthaml-openbank/vault-unit/support/cast"
)

// Account represents metadata of account entity
type Account struct {
	Name            string `json:"name"`
	Format          string `json:"format"`
	Currency        string `json:"currency"`
	IsBalanceCheck  bool   `json:"isBalanceCheck"`
	Balance         Dec
	Promised        Dec
	Promises        Promises
	SnapshotVersion int64
	EventCounter    int64
}

// NewAccount returns new Account
func NewAccount(name string) Account {
	return Account{
		Name:            name,
		Format:          "???",
		Currency:        "???",
		IsBalanceCheck:  true,
		Balance:         Dec{},
		Promised:        Dec{},
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
		buffer.WriteString(entity.Currency)
		buffer.WriteString("??")
	case 2:
		buffer.WriteString(entity.Currency)
		buffer.WriteString("?")
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

	buffer.WriteString(entity.Balance.String())

	buffer.WriteString("\n")

	buffer.WriteString(entity.Promised.String())

	for i := 0; i < len(entity.Promises.keys); i++ {
		buffer.WriteString("\n")
		buffer.WriteString(entity.Promises.values[entity.Promises.keys[i]])
	}

	return buffer.Bytes()
}

// Deserialize Account entity from persistable data
func (entity *Account) Deserialize(data []byte) {
	var (
		i = 0
		j = 4
		l = len(data)
	)

	if entity == nil {
		return
	}

	entity.Promises = NewPromises()

	if l < 3 {
		return
	}

	entity.Currency = cast.BytesToString(data[0 : 3])

	i = j
	for ;j < l && data[j] != '\n' ; j++ {}

	entity.IsBalanceCheck = (data[j-1] != byte('F'))
	entity.Format = cast.BytesToString(data[i : j-2])

	j++
	if j >= l {
		return
	}

	i = j
	for ;j < l && data[j] != '\n' ; j++ {}
	if j < l {
		j++
	}

	_ = entity.Balance.SetString(cast.BytesToString(data[i:j]))
	if j >= l {
		return
	}
	
	i = j
	for ;j < l && data[j] != '\n' ; j++ {}
	if j < l {
		j++
	}

	_ = entity.Promised.SetString(cast.BytesToString(data[i:j]))
	if j >= l {
		return
	}
	
	i = j
	for ;j < l; j++ {
		if data[j] == '\n' {
			entity.Promises.Add(cast.BytesToString(data[i:j]))
			i = j + 1
		}
	}

	if j >= l && i < l {
		entity.Promises.Add(cast.BytesToString(data[i:]))
	}
}

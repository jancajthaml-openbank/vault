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

	//money "gopkg.in/inf.v0"
)

// Account represents metadata of account entity
type Account struct {
	Name            string `json:"name"`
	Format          string `json:"format"`
	Currency        string `json:"currency"`
	IsBalanceCheck  bool   `json:"isBalanceCheck"`
	Balance         *Dec
	Promised        *Dec
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
		Balance:         new(Dec),
		Promised:        new(Dec),
		SnapshotVersion: 0,
		EventCounter:    0,
		Promises:        NewPromises(),
	}
}

// Serialize Account entity to persistable data
func (entity Account) Serialize() []byte {
	var buffer bytes.Buffer // alloc

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

	if entity.Balance == nil {
		buffer.WriteString("0.0")
	} else {
		buffer.WriteString(entity.Balance.String())	// allos
	}

	buffer.WriteString("\n")

	if entity.Promised == nil {
		buffer.WriteString("0.0")
	} else {
		buffer.WriteString(entity.Promised.String())	// alloc
	}

	for i := 0; i < len(entity.Promises.keys); i++ {
		buffer.WriteString("\n")
		buffer.WriteString(entity.Promises.values[entity.Promises.keys[i]])
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
		j = bytes.IndexByte(data, '\n')
		l = len(data)
	)

	if j < i || j > l {
		j = l
	}

	entity.IsBalanceCheck = (data[j-1] != byte('F'))
	entity.Format = string(data[i : j-2])

	if j >= l {
		return
	}

	i = j + 1
	j = bytes.IndexByte(data[i:], '\n')

	if j < 0 {
		if i < l {
			entity.Balance, _ = new(Dec).SetString(string(data[i:]))
		}
		return
	}
	j += i

	entity.Balance, _ = new(Dec).SetString(string(data[i:j]))
	i = j + 1

	j = bytes.IndexByte(data[i:], '\n')
	if j < 0 {
		if i < l {
			entity.Promised, _ = new(Dec).SetString(string(data[i:]))
		}
		return
	}
	j += i
	entity.Promised, _ = new(Dec).SetString(string(data[i:j]))
	i = j + 1

	for {
		j = bytes.IndexByte(data[i:], '\n')
		if j < 0 {
			if i < l {
				entity.Promises.Add(string(data[i:]))
			}
			return
		}
		j += i
		entity.Promises.Add(string(data[i:j]))
		i = j + 1
	}
}

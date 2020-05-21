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
	Name           string `json:"name"`
	Format         string `json:"format"`
	Currency       string `json:"currency"`
	IsBalanceCheck bool   `json:"isBalanceCheck"`
	Balance        *money.Dec
	Promised       *money.Dec
	PromiseBuffer  TransactionSet
	Version        int64
}

// Copy returns copy of Account
func (entity Account) Copy() Account {
	return Account{
		Name:           entity.Name,
		Format:         entity.Format,
		Currency:       entity.Currency,
		IsBalanceCheck: entity.IsBalanceCheck,
		Balance:        new(money.Dec).Set(entity.Balance),
		Promised:       new(money.Dec).Set(entity.Promised),
		PromiseBuffer:  entity.PromiseBuffer, //.Copy(), // FIXME implement
		Version:        entity.Version,
	}
}

// CreateAccount is inbound request for creation of new account
type CreateAccount struct {
	Name           string
	Format         string
	Currency       string
	IsBalanceCheck bool
}

// GetAccountState is inbound request for balance of account
type GetAccountState struct {
}

// Update is inbound request to update snapshot
type Update struct {
	Version int64
}

// Promise is inbound request for transaction promise
type Promise struct {
	Transaction string
	Amount      *money.Dec
	Currency    string
}

// FIXME commit and rollback online TransactionID save amount in PromiseSet

// Commit is inbound request for transaction commit
type Commit struct {
	Transaction string
	Amount      *money.Dec
	Currency    string
}

// FIXME commit and rollback online TransactionID save amount in PromiseSet

// Rollback is inbound request for transaction rollback
type Rollback struct {
	Transaction string
	Amount      *money.Dec
	Currency    string
}

// Committed is reply message that transaction is committed
type Committed struct {
	IDTransaction string
}

// Rollbacked is reply message that transaction is rollbacked
type Rollbacked struct {
	IDTransaction string
}

// NewAccount returns new Account
func NewAccount(name string) Account {
	return Account{
		Name:           name,
		Format:         "???",
		Currency:       "???",
		IsBalanceCheck: true,
		Balance:        new(money.Dec),
		Promised:       new(money.Dec),
		Version:        0,
		PromiseBuffer:  NewTransactionSet(),
	}
}

// Serialise Account entity to persistable data
func (entity Account) Serialise() []byte {
	var buffer bytes.Buffer

	// [CURRENCY FORMAT_IS-CHECK]
	// [BALANCE]
	// [PROMISED]
	// [...PROMISE-BUFFER]

	buffer.WriteString("???"[0 : 3-len(entity.Currency)])
	buffer.WriteString(entity.Currency)
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

	for _, v := range entity.PromiseBuffer.Values() {
		buffer.WriteString("\n")
		buffer.WriteString(v)
	}

	return buffer.Bytes()
}

// Deserialise Account entity from persistable data
func (entity *Account) Deserialise(data []byte) {
	if entity == nil {
		return
	}

	entity.PromiseBuffer = NewTransactionSet()
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

scan:
	j = bytes.IndexByte(data[i:], '\n')
	if j < 0 {
		if len(data) > 0 {
			entity.PromiseBuffer.Add(string(data[i:]))
		}
		return
	}
	j += i
	entity.PromiseBuffer.Add(string(data[i:j]))
	i = j + 1
	goto scan
}

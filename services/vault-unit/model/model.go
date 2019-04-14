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

package model

import (
	"bytes"

	money "gopkg.in/inf.v0"
)

// Account represents metadata of account entity
type Account struct {
	AccountName    string `json:"name"`
	Currency       string `json:"currency"`
	IsBalanceCheck bool   `json:"isBalanceCheck"`
	Balance        *money.Dec
	Promised       *money.Dec
	PromiseBuffer  TransactionSet
	Version        int
}

// Copy returns copy of Account
func (entity Account) Copy() Account {
	return Account{
		AccountName:    entity.AccountName,
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
	AccountName    string
	Currency       string
	IsBalanceCheck bool
}

// GetAccountState is inbound request for balance of account
type GetAccountState struct {
}

// Update is inbound request to update snapshot
type Update struct {
	Version int
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
		AccountName:    name,
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

	if entity.IsBalanceCheck {
		buffer.WriteString("T")
	} else {
		buffer.WriteString("F")
	}

	buffer.WriteString("???"[0 : 3-len(entity.Currency)])
	buffer.WriteString(entity.Currency)
	buffer.WriteString("\n")

	if entity.Balance == nil {
		buffer.WriteString("0.0\n")
	} else {
		buffer.WriteString(entity.Balance.String())
		buffer.WriteString("\n")
	}

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
	entity.IsBalanceCheck = string(data[0]) != "F"
	entity.Currency = string(data[1:4])

	var (
		i = 5
		j = bytes.IndexByte(data[5:], '\n') + 5
	)

	entity.Balance, _ = new(money.Dec).SetString(string(data[i:j]))

	i = j + 1
	j = bytes.IndexByte(data[i:], '\n')
	if j < 0 {
		if len(data) > 0 {
			entity.Promised, _ = new(money.Dec).SetString(string(data[i:]))
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

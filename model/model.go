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

package model

import (
	"bytes"
	"strconv"

	money "gopkg.in/inf.v0"
)

// Account represents metadata of account entity
type Account struct {
	AccountName    string `json:"accountNumber"`
	Currency       string `json:"currency"`
	IsBalanceCheck bool   `json:"isBalanceCheck"`
	Balance        *money.Dec
	Promised       *money.Dec
	PromiseBuffer  TransactionSet
	Version        int
}

// Copy returns copy of Account
func (s Account) Copy() Account {
	return Account{
		AccountName:    s.AccountName,
		Currency:       s.Currency,
		IsBalanceCheck: s.IsBalanceCheck,
		Balance:        new(money.Dec).Set(s.Balance),
		Promised:       new(money.Dec).Set(s.Promised),
		PromiseBuffer:  s.PromiseBuffer, //.Copy(), // FIXME implement
		Version:        s.Version,
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

// Persist serializes Account entity to persistable data
func (entity *Account) Persist() []byte {
	var buffer bytes.Buffer

	if entity.IsBalanceCheck {
		buffer.WriteString("T")
	} else {
		buffer.WriteString("F")
	}

	buffer.WriteString(entity.Currency)
	buffer.WriteString(entity.AccountName)
	buffer.WriteString("\n")
	buffer.WriteString(strconv.Itoa(entity.Version))
	buffer.WriteString("\n")
	buffer.WriteString(entity.Balance.String())
	buffer.WriteString("\n")
	buffer.WriteString(entity.Promised.String())

	for v := range entity.PromiseBuffer.Items {
		buffer.WriteString("\n")
		buffer.WriteString(v)
	}

	return buffer.Bytes()
}

// Hydrate deserializes Account entity from persistent data
func (entity *Account) Hydrate(data []byte) {
	var (
		j = 0
		i = 4
	)

	entity.PromiseBuffer = NewTransactionSet()
	entity.IsBalanceCheck = string(data[0]) != "F"
	entity.Currency = string(data[1:4])

	j = bytes.IndexByte(data[4:], '\n') + 4
	entity.AccountName = string(data[4:j])
	i = j + 1

	j = bytes.IndexByte(data[i:], '\n')
	j += i
	entity.Version, _ = strconv.Atoi(string(data[i:j]))
	i = j + 1

	j = bytes.IndexByte(data[i:], '\n')
	j += i
	entity.Balance, _ = new(money.Dec).SetString(string(data[i:j]))
	i = j + 1

	j = bytes.IndexByte(data[i:], '\n')
	if j < 0 {
		entity.Promised, _ = new(money.Dec).SetString(string(data[i:]))
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

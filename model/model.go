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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	money "gopkg.in/inf.v0"
)

// Account represents metadata of account entity
type Account struct {
	AccountName    string `json:"accountNumber"`
	Currency       string `json:"currency"`
	IsBalanceCheck bool   `json:"isBalanceCheck"`
}

// Snapshot represents current state of account entity
type Snapshot struct {
	Balance       *money.Dec
	Promised      *money.Dec
	PromiseBuffer TransactionSet
	Version       int
}

// CreateAccount is inbound request for creation of new account
type CreateAccount struct {
	AccountName    string
	Currency       string
	IsBalanceCheck bool
}

// GetAccountBalance is inbound request for balance of account
type GetAccountBalance struct {
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

// FIXME commit and rollback online TransactionId save amount in PromiseSet
// Commit is inbound request for transaction commit
type Commit struct {
	Transaction string
	Amount      *money.Dec
	Currency    string
}

// FIXME commit and rollback online TransactionId save amount in PromiseSet
// Rollback is inbound request for transaction rollback
type Rollback struct {
	Transaction string
	Amount      *money.Dec
	Currency    string
}

// Committed is reply message that transaction is commited
type Committed struct {
	IDTransaction string
}

// Rollbacked is reply message that transaction is rollbacked
type Rollbacked struct {
	IDTransaction string
}

// NewSnapshot returns new Snapshot
func NewSnapshot() Snapshot {
	snapshot := Snapshot{}
	snapshot.Balance = new(money.Dec)
	snapshot.Promised = new(money.Dec)
	snapshot.Version = 0

	return snapshot
}

// NewAccount returns new Account
func NewAccount(name string) Account {
	account := Account{}
	account.AccountName = name
	account.Currency = "???"
	account.IsBalanceCheck = true

	return account
}

// Copy returns value copy of Snapshot
func (entity *Snapshot) Copy() *Snapshot {
	clone := new(Snapshot)
	*clone = *entity
	return clone
}

// SerializeForStorage returns data for Snapshot persistence
func (entity *Snapshot) SerializeForStorage() []byte {
	var buffer bytes.Buffer

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(entity.Version))
	buffer.Write(buf)

	buffer.WriteString(entity.Balance.String())
	buffer.WriteString("\n")
	buffer.WriteString(entity.Promised.String())

	entity.PromiseBuffer.WriteTo(&buffer)

	return buffer.Bytes()
}

// DeserializeFromStorage Snapshot instance from persisted data
func (entity *Snapshot) DeserializeFromStorage(data []byte) {
	// FIXME try to find way to save without newline
	version := int(binary.BigEndian.Uint32(data[:4]))
	entity.Version = version

	lines := strings.Split(string(data[4:]), "\n")

	balance, _ := new(money.Dec).SetString(lines[0])
	promised, _ := new(money.Dec).SetString(lines[1])

	entity.Balance = balance
	entity.Promised = promised
	entity.PromiseBuffer = NewTransactionSet()

	entity.PromiseBuffer.AddAll(lines[2:])

	return
}

// SerializeForStorage returns data for Account persistence
func (entity *Account) SerializeForStorage() []byte {
	if entity.IsBalanceCheck {
		return []byte("t" + strings.ToUpper(entity.Currency) + entity.AccountName)
	}

	return []byte("f" + strings.ToUpper(entity.Currency) + entity.AccountName)
}

// DeserializeFromStorage return Account instance from persisted data
func (entity *Account) DeserializeFromStorage(data []byte) {
	entity.IsBalanceCheck = string(data[0]) != "f"
	entity.Currency = string(data[1:4])
	entity.AccountName = string(data[4:])

	return
}

// UnmarshalJSON is json Account unmarhalling companion
func (entity *Account) UnmarshalJSON(data []byte) (err error) {
	all := struct {
		AccountNumber  *string `json:"accountNumber"`
		Currency       *string `json:"currency"`
		IsBalanceCheck *bool   `json:"isBalanceCheck"`
	}{}

	err = json.Unmarshal(data, &all)
	if err != nil {
		return
	}

	if all.AccountNumber == nil {
		err = fmt.Errorf("Required field for accountNumber missing")
		return
	}

	if all.Currency == nil {
		err = fmt.Errorf("Required field for Currency missing")
		return
	}

	if len(*all.Currency) != 3 {
		err = fmt.Errorf("Invalid currency")
		return
	}

	if all.IsBalanceCheck == nil {
		err = fmt.Errorf("Required field for isBalanceCheck missing")
		return
	}

	entity.AccountName = *all.AccountNumber
	entity.Currency = strings.ToUpper(*all.Currency)
	entity.IsBalanceCheck = *all.IsBalanceCheck

	return
}

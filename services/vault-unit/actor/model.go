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

package actor

import (
	"github.com/jancajthaml-openbank/vault-unit/model"
)

// CreateAccount is inbound request for creation of new account
type CreateAccount struct {
	Format         string
	Currency       string
	IsBalanceCheck bool
}

// GetAccountState is inbound request for balance of account
type GetAccountState struct {
}

// Promise is inbound request for transaction promise
type Promise struct {
	Transaction string
	Amount      *model.Dec
	Currency    string
}

// Commit is inbound request for transaction commit
type Commit struct {
	Transaction string
	Amount      *model.Dec
	Currency    string
}

// Rollback is inbound request for transaction rollback
type Rollback struct {
	Transaction string
	Amount      *model.Dec
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

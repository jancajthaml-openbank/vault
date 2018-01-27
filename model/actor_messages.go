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

const (
	ReqAccountBalance  string = "get_balance"
	RespAccountBalance string = "account_balance"

	ReqCreateAccount  string = "create_account"
	RespCreateAccount string = "account_created"

	PromiseOrder  string = EventPromise + "_order"
	CommitOrder   string = EventCommit + "_order"
	RollbackOrder string = EventRollback + "_order"

	PromiseAccepted  string = EventPromise + "_accepted"
	CommitAccepted   string = EventCommit + "_accepted"
	RollbackAccepted string = EventRollback + "_accepted"

	UpdateSnapshot string = "update_snapshot"
)

// FatalErrorMessage is reply message carrying failure
func FatalErrorMessage(self, sender string) string {
	return "error " + self + " " + sender
}

// AccountCreatedMessage is reply message informing that account was created
func AccountCreatedMessage(self, sender string) string {
	return RespCreateAccount + " " + self + " " + sender
}

// AccountCreatedMessage is reply message informing that transaction promise was
// accepted
func PromiseAcceptedMessage(self, sender string) string {
	return PromiseAccepted + " " + self + " " + sender
}

// AccountCreatedMessage is reply message informing that transaction commit was
// accepted
func CommitAcceptedMessage(self, sender string) string {
	return CommitAccepted + " " + self + " " + sender
}

// AccountCreatedMessage is reply message informing that transaction rollback
// was accepted
func RollbackAcceptedMessage(self, sender string) string {
	return RollbackAccepted + " " + self + " " + sender
}

// AccountCreatedMessage is reply message carrying account balance
func AccountBalanceMessage(self, sender, currency, balance string) string {
	return RespAccountBalance + " " + self + " " + sender + " " + currency + " " + balance
}

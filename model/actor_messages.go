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
	// EventPromise represents promise prefix
	EventPromise = "0"
	// EventCommit represents commit prefix
	EventCommit = "1"
	// EventRollback represents rollback prefix
	EventRollback = "2"
)

const (
	// ReqAccountState vault message request code for "Get Snapshot"
	ReqAccountState = "GS"
	// RespAccountState vault message response code for "Get Snapshot"
	RespAccountState = "SG"
	// ReqCreateAccount vault message request code for "New Account"
	ReqCreateAccount = "NA"
	// RespCreateAccount vault message response code for "New Account"
	RespCreateAccount = "AN"
	// PromiseOrder vault message request code for "Promise"
	PromiseOrder = EventPromise + "X"
	// CommitOrder vault message request code for "Commit"
	CommitOrder = EventCommit + "X"
	// RollbackOrder vault message request code for "Rollback"
	RollbackOrder = EventRollback + "X"
	// PromiseAccepted vault message response code for "Promise"
	PromiseAccepted = "X" + EventPromise
	// CommitAccepted vault message response code for "Commit"
	CommitAccepted = "X" + EventCommit
	// RollbackAccepted vault message response code for "Rollback"
	RollbackAccepted = "X" + EventRollback
	// FatalError vault message response code for "Error"
	FatalError = "EE"
	// UpdateSnapshot vault message request code for "Update Snapshot"
	UpdateSnapshot = "US"
)

// FatalErrorMessage is reply message carrying failure
func FatalErrorMessage(self, sender string) string {
	return self + " " + sender + " " + FatalError
}

// AccountCreatedMessage is reply message informing that account was created
func AccountCreatedMessage(self, sender string) string {
	return self + " " + sender + " " + RespCreateAccount
}

// PromiseAcceptedMessage is reply message informing that transaction promise was
// accepted
func PromiseAcceptedMessage(self, sender string) string {
	return self + " " + sender + " " + PromiseAccepted
}

// CommitAcceptedMessage is reply message informing that transaction commit was
// accepted
func CommitAcceptedMessage(self, sender string) string {
	return self + " " + sender + " " + CommitAccepted
}

// RollbackAcceptedMessage is reply message informing that transaction rollback
// was accepted
func RollbackAcceptedMessage(self, sender string) string {
	return self + " " + sender + " " + RollbackAccepted
}

// AccountStateMessage is reply message carrying account state
func AccountStateMessage(self, sender, currency, balance, promised string, isBalanceCheck bool) string {
	if isBalanceCheck {
		return self + " " + sender + " " + RespAccountState + " " + currency + " t " + balance + " " + promised
	}

	return self + " " + sender + " " + RespAccountState + " " + currency + " f " + balance + " " + promised
}

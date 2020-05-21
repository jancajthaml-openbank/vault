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
	"strconv"
	"github.com/jancajthaml-openbank/vault-unit/model"

	//system "github.com/jancajthaml-openbank/actor-system"
)

const (
	// ReqAccountState vault message request code for "Get Snapshot"
	ReqAccountState = "GS"
	// RespAccountState vault message response code for "Get Snapshot"
	RespAccountState = "S0"
	// RespAccountMissing vault message response code for "Get Snapshot"
	RespAccountMissing = "S1"
	// ReqCreateAccount vault message request code for "New Account"
	ReqCreateAccount = "NA"
	// RespCreateAccount vault message response code for "New Account"
	RespCreateAccount = "AN"

	// PromiseOrder vault message request code for "Promise"
	PromiseOrder = "NP"
	// PromiseAccepted vault message response code for "Promise" accepted
	PromiseAccepted = "P1"
	// PromiseRejected vault message response code for "Promise" rejected
	PromiseRejected = "P2"

	// CommitOrder vault message request code for "Commit"
	CommitOrder = "NC"
	// CommitAccepted vault message response code for "Commit" accepted
	CommitAccepted = "C1"
	// CommitRejected vault message response code for "Commit" rejected
	CommitRejected = "C2"

	// RollbackOrder vault message request code for "Rollback"
	RollbackOrder = "NR"
	// RollbackAccepted vault message response code for "Rollback" accepted
	RollbackAccepted = "R1"
	// RollbackRejected vault message response code for "Rollback" rejected
	RollbackRejected = "R2"

	// FatalError vault message response code for "Error"
	FatalError = "EE"
	// UpdateSnapshot vault message request code for "Update Snapshot"
	UpdateSnapshot = "US"
)

// FatalErrorMessage is reply message carrying failure
func UpdateSnapshotMessage(version int64) string {
	return UpdateSnapshot + " " + strconv.FormatInt(version, 10)
}

// FatalErrorMessage is reply message carrying failure
func FatalErrorMessage() string {
	return FatalError
}

// AccountCreatedMessage is reply message informing that account was created
func AccountCreatedMessage() string {
	return RespCreateAccount
}

// PromiseAcceptedMessage is reply message informing that transaction promise was
// accepted
func PromiseAcceptedMessage() string {
	return PromiseAccepted
}

// PromiseRejectedMessage is reply message informing that transaction promise was
// rejected
func PromiseRejectedMessage(reason string) string {
	return PromiseRejected + " " + reason
}

// CommitAcceptedMessage is reply message informing that transaction commit was
// accepted
func CommitAcceptedMessage() string {
	return CommitAccepted
}

// CommitRejectedMessage is reply message informing that transaction commit was
// rejected
func CommitRejectedMessage(reason string) string {
	return CommitRejected + " " + reason
}

// RollbackAcceptedMessage is reply message informing that transaction rollback
// was accepted
func RollbackAcceptedMessage() string {
	return RollbackAccepted
}

// RollbackRejectedMessage is reply message informing that transaction rollback
// was rejected
func RollbackRejectedMessage(reason string) string {
	return RollbackRejected + " " + reason
}

// AccountStateMessage is reply message carrying account state
func AccountStateMessage(state model.Account) string {
	if state.IsBalanceCheck {
		return RespAccountState + " " + state.Format + " " + state.Currency + " t " + state.Balance.String() + " " + state.Promised.String()
	}
	return RespAccountState + " " + state.Format + " " + state.Currency + " f " + state.Balance.String() + " " + state.Promised.String()
}

// AccountMissingMessage is reply message informing that account does not exist
func AccountMissingMessage() string {
	return RespAccountMissing
}

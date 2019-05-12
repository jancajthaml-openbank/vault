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

package actor

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
	// FatalError vault message response code for "Error"
	FatalError = "EE"
)

// CreateAccountMessage is message for creation of new account
func CreateAccountMessage(tenant string, sender string, name string, currency string, isBalanceCheck bool) string {
	if isBalanceCheck {
		return "VaultUnit/" + tenant + " VaultRest " + name + " " + sender + " " + ReqCreateAccount + " " + currency + " t"
	}
	return "VaultUnit/" + tenant + " VaultRest " + name + " " + sender + " " + ReqCreateAccount + " " + currency + " f"
}

// GetAccountMessage is message for getting balance of account
func GetAccountMessage(tenant string, sender string, name string) string {
	return "VaultUnit/" + tenant + " VaultRest " + name + " " + sender + " " + ReqAccountState
}

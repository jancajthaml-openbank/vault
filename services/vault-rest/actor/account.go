// Copyright (c) 2016-2023, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"time"

	"github.com/jancajthaml-openbank/vault-rest/model"
	"github.com/rs/xid"

	system "github.com/jancajthaml-openbank/actor-system"
)

func receive(sys *System, channel chan<- interface{}) system.ReceiverFunction {
	return func(context system.Context) system.ReceiverFunction {
		channel <- context.Data
		return receive(sys, channel)
	}
}

// CreateAccount creates new account for target tenant vault
func CreateAccount(sys *System, tenant string, account model.Account) interface{} {
	ch := make(chan interface{})

	envelope := system.NewActor("relay/"+xid.New().String(), receive(sys, ch))

	sys.RegisterActor(envelope)
	defer sys.UnregisterActor(envelope.Name)

	sys.SendMessage(
		CreateAccountMessage(account.Format, account.Currency, account.IsBalanceCheck),
		system.Coordinates{
			Region: "VaultUnit/" + tenant,
			Name:   account.Name,
		},
		system.Coordinates{
			Region: "VaultRest",
			Name:   envelope.Name,
		},
	)

	select {
	case result := <-ch:
		return result
	case <-time.After(5 * time.Second):
		return new(ReplyTimeout)
	}
}

// GetAccount retrives account state from target tenant vault
func GetAccount(sys *System, tenant string, name string) interface{} {
	ch := make(chan interface{})

	envelope := system.NewActor("relay/"+xid.New().String(), receive(sys, ch))

	sys.RegisterActor(envelope)
	defer sys.UnregisterActor(envelope.Name)

	sys.SendMessage(
		ReqAccountState,
		system.Coordinates{
			Region: "VaultUnit/" + tenant,
			Name:   name,
		},
		system.Coordinates{
			Region: "VaultRest",
			Name:   envelope.Name,
		},
	)

	select {
	case result := <-ch:
		return result
	case <-time.After(5 * time.Second):
		return new(ReplyTimeout)
	}
}

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
	"time"

	"github.com/jancajthaml-openbank/vault-rest/model"
	"github.com/rs/xid"

	system "github.com/jancajthaml-openbank/actor-system"
)

// CreateAccount creates new account for target tenant vault
func CreateAccount(sys *ActorSystem, tenant string, account model.Account) (result interface{}) {
	sys.Metrics.TimeCreateAccount(func() {
		ch := make(chan interface{})
		defer close(ch)

		envelope := system.NewEnvelope("relay/"+xid.New().String(), nil)
		defer sys.UnregisterActor(envelope.Name)

		sys.RegisterActor(envelope, func(state interface{}, context system.Context) {
			ch <- context.Data
		})

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

		case result = <-ch:
			return

		case <-time.After(10 * time.Second):
			result = new(ReplyTimeout)
			return
		}
	})
	return
}

// GetAccount retrives account state from target tenant vault
func GetAccount(sys *ActorSystem, tenant string, name string) (result interface{}) {
	sys.Metrics.TimeGetAccount(func() {
		ch := make(chan interface{})
		defer close(ch)

		envelope := system.NewEnvelope("relay/"+xid.New().String(), nil)
		defer sys.UnregisterActor(envelope.Name)

		sys.RegisterActor(envelope, func(state interface{}, context system.Context) {
			ch <- context.Data
		})

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

		case result = <-ch:
			return

		case <-time.After(10 * time.Second):
			result = new(ReplyTimeout)
			return
		}
	})
	return
}

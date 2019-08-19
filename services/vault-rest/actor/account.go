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

import (
	"time"

	"github.com/rs/xid"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
)

// CreateAccount creates new account for target tenant vault
func CreateAccount(sys *ActorSystem, tenant string, account Account) (result interface{}) {
	sys.Metrics.TimeCreateAccount(func() {
		// FIXME properly determine fail states
		// input validation -> input error
		// system in invalid state (and panics) -> fatal error
		// timeout -> timeout
		// account answer -> expected vs unexpected

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("CreateAccount recovered in %v", r)
				result = nil
			}
		}()

		ch := make(chan interface{})
		defer close(ch)

		envelope := system.NewEnvelope("relay/"+xid.New().String(), nil)
		defer sys.UnregisterActor(envelope.Name)

		sys.RegisterActor(envelope, func(state interface{}, context system.Context) {
			ch <- context.Data
		})

		sys.SendRemote(CreateAccountMessage(tenant, envelope.Name, account.Name, account.Format, account.Currency, account.IsBalanceCheck))

		select {

		case result = <-ch:
			return

		case <-time.After(time.Second):
			result = new(ReplyTimeout)
			return
		}
	})
	return
}

// GetAccount retrives account state from target tenant vault
func GetAccount(sys *ActorSystem, tenant string, name string) (result interface{}) {
	sys.Metrics.TimeGetAccount(func() {
		// FIXME properly determine fail states
		// input validation -> input error
		// system in invalid state (and panics) -> fatal error
		// timeout -> timeout
		// account answer -> expected vs unexpected

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("GetAccount recovered in %v", r)
				result = nil
			}
		}()

		ch := make(chan interface{})
		defer close(ch)

		envelope := system.NewEnvelope("relay/"+xid.New().String(), nil)
		defer sys.UnregisterActor(envelope.Name)

		sys.RegisterActor(envelope, func(state interface{}, context system.Context) {
			ch <- context.Data
		})

		sys.SendRemote(GetAccountMessage(tenant, envelope.Name, name))

		select {

		case result = <-ch:
			return

		case <-time.After(time.Second):
			result = new(ReplyTimeout)
			return
		}
	})
	return
}

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

package actor

import (
	"fmt"
	"time"

	"github.com/jancajthaml-openbank/vault-rest/daemon"
	"github.com/jancajthaml-openbank/vault-rest/model"

	"github.com/rs/xid"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
)

// CreateAccount creates new account for target tenant vault
func CreateAccount(s *daemon.ActorSystem, tenant string, account model.Account) (result interface{}) {
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
	defer s.UnregisterActor(envelope.Name)

	s.RegisterActor(envelope, func(state interface{}, context system.Context) {
		switch msg := context.Data.(type) {
		case model.AccountCreated:
			if account.IsBalanceCheck {
				s.BroadcastRemote(tenant + " A_NEW " + account.Name + " " + account.Currency + " T") // FIXME ACTIVE
			} else {
				s.BroadcastRemote(tenant + " A_NEW " + account.Name + " " + account.Currency + " F") // FIXME PASIVE
			}
			ch <- &msg
		default:
			ch <- nil
		}
	})

	s.SendRemote("VaultUnit/"+tenant, CreateAccountMessage(envelope.Name, account.Name, account.Currency, account.IsBalanceCheck))

	select {

	case result = <-ch:
		return

	case <-time.After(time.Second):
		result = new(model.ReplyTimeout)
		return
	}
	return
}

// GetAccount retrives account state from target tenant vault
func GetAccount(s *daemon.ActorSystem, tenant string, id string) (result interface{}) {
	// FIXME properly determine fail states
	// input validation -> input error
	// system in invalid state (and panics) -> fatal error
	// timeout -> timeout
	// account answer -> expected vs unexpected

	fmt.Printf("actor get account tenant: %+v, id: %+v\n", tenant, id)

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("GetAccount recovered in %v", r)
			result = nil
		}
	}()

	ch := make(chan interface{})
	defer close(ch)

	envelope := system.NewEnvelope("relay/"+xid.New().String(), nil)
	defer s.UnregisterActor(envelope.Name)

	s.RegisterActor(envelope, func(state interface{}, context system.Context) {
		switch msg := context.Data.(type) {
		case model.Account:
			ch <- &msg
		default:
			ch <- nil
		}
	})

	s.SendRemote("VaultUnit/"+tenant, GetAccountMessage(envelope.Name, id))

	select {

	case result = <-ch:
		return

	case <-time.After(time.Second):
		result = new(model.ReplyTimeout)
		return
	}

	return
}

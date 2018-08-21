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
	"github.com/jancajthaml-openbank/vault/pkg/metrics"
	"github.com/jancajthaml-openbank/vault/pkg/model"
	"github.com/jancajthaml-openbank/vault/pkg/persistence"
	"github.com/jancajthaml-openbank/vault/pkg/utils"

	"strings"

	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

func nilAccount(params utils.RunParams, m *metrics.Metrics, system *ActorSystem) func(model.Account, Context) {
	return func(state model.Account, context Context) {
		snapshotHydration := persistence.LoadAccount(params, state.AccountName)

		if snapshotHydration == nil {
			context.Receiver.Become(state, nonExistAccount(params, m, system))
			log.Debugf("%s ~ Nil -> NonExist", state.AccountName)
		} else {
			context.Receiver.Become(*snapshotHydration, existAccount(params, m, system))
			log.Debugf("%s ~ Nil -> Exist", state.AccountName)
		}

		context.Receiver.Tell(context.Data, context.Sender)
	}
}

func nonExistAccount(params utils.RunParams, m *metrics.Metrics, system *ActorSystem) func(model.Account, Context) {
	return func(state model.Account, context Context) {
		switch msg := context.Data.(type) {

		case model.CreateAccount:
			currency := strings.ToUpper(msg.Currency)
			isBalanceCheck := msg.IsBalanceCheck

			snaphostResult := persistence.CreateAccount(params, state.AccountName, currency, isBalanceCheck)

			if snaphostResult == nil {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (NonExist CreateAccount) Error", state.AccountName)
				return
			}

			context.Receiver.Become(*snaphostResult, existAccount(params, m, system))
			system.SendRemote(context.Sender.Region, model.AccountCreatedMessage(context.Receiver.Name, context.Sender.Name))
			log.Infof("New Account %s Created", state.AccountName)
			log.Debugf("%s ~ (NonExist CreateAccount) OK", state.AccountName)
			m.AccountCreated()

		case model.Rollback:
			system.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist Rollback) OK", state.AccountName)

		default:
			system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist Unknown Message) Error", state.AccountName)

		}

		return
	}
}

func existAccount(params utils.RunParams, m *metrics.Metrics, system *ActorSystem) func(model.Account, Context) {
	return func(state model.Account, context Context) {
		switch msg := context.Data.(type) {

		case model.GetAccountState:
			system.SendRemote(context.Sender.Region, model.AccountStateMessage(context.Receiver.Name, context.Sender.Name, state.Currency, state.Balance.String(), state.Promised.String(), state.IsBalanceCheck))
			log.Debugf("%s ~ (Exist GetAccountState) OK", state.AccountName)

		case model.CreateAccount:
			system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist CreateAccount) Error", state.AccountName)

		case model.Promise:
			if state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.PromiseAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) OK Already Accepted", state.AccountName)
				return
			}

			if state.Currency != msg.Currency {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Promise) Error Currency Mismatch", state.AccountName)
				return
			}

			nextPromised := new(money.Dec).Add(state.Promised, msg.Amount)

			if !state.IsBalanceCheck || new(money.Dec).Add(state.Balance, nextPromised).Sign() >= 0 {
				if !persistence.PersistPromise(params, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
					system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
					log.Warnf("%s ~ (Exist Promise) Error Could not Persist", state.AccountName)
					return
				}

				next := state.Copy()
				next.Promised = nextPromised
				next.PromiseBuffer.Add(msg.Transaction)

				context.Receiver.Become(*next, existAccount(params, m, system))

				system.SendRemote(context.Sender.Region, model.PromiseAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) OK", state.AccountName)
				m.PromiseAccepted()
				return
			}

			if new(money.Dec).Sub(state.Balance, msg.Amount).Sign() < 0 {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) Error Insufficient Funds", state.AccountName)
				return
			}

			// FIXME boucing not handled
			system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Promise) Error ... (Bounce?)", state.AccountName)

		case model.Commit:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.CommitAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Commit) OK Already Accepted", state.AccountName)
				return
			}

			if !persistence.PersistCommit(params, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Commit) Error Could not Persist", state.AccountName)
				return
			}

			next := state.Copy()
			next.Balance = new(money.Dec).Add(state.Balance, msg.Amount)
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(*next, existAccount(params, m, system))

			system.SendRemote(context.Sender.Region, model.CommitAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist Commit) OK", state.AccountName)
			m.CommitAccepted()

		case model.Rollback:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Rollback) OK Already Accepted", state.AccountName)
				return
			}

			if !persistence.PersistRollback(params, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Rollback) Error Could not Persist", state.AccountName)
				return
			}

			next := state.Copy()
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(*next, existAccount(params, m, system))

			system.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist Rollback) OK", state.AccountName)
			m.RollbackAccepted()

		case model.Update:
			if msg.Version != state.Version {
				log.Debugf("%s ~ (Exist Update) Error Already Updated", state.AccountName)
				return
			}

			result := persistence.LoadAccount(params, state.AccountName)
			if result == nil {
				log.Warnf("%s ~ (Exist Update) Error no existing snapshot", state.AccountName)
				return
			}

			next := persistence.UpdateAccount(params, state.AccountName, result)
			if next == nil {
				log.Warnf("%s ~ (Exist Update) Error unable to update", state.AccountName)
				return
			}

			context.Receiver.Become(*next, existAccount(params, m, system))
			log.Debugf("%s ~ (Exist Update) OK", state.AccountName)

		default:
			system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Unknown Message) Error", state.AccountName)

		}

		return
	}
}

// FIXME split to multiple functions
// SpawnAccountActor returns new account actor instance registered into actor
// system
func (system *ActorSystem) SpawnAccountActor(params utils.RunParams, m *metrics.Metrics, path string) string {
	if system == nil {
		log.Warnf("%s ~ Spawning Actor Error no Actor System", path)
		return ""
	}

	envelope := NewActor(path)
	err := system.RegisterActor(envelope, nilAccount(params, m, system))
	if err != nil {
		log.Warnf("%s ~ Spawning Actor Error unable to register", path)
		return ""
	}

	log.Debugf("%s ~ Actor Spawned", path)
	return envelope.Name
}

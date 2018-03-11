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
	"github.com/jancajthaml-openbank/vault/metrics"
	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/utils"

	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

func nilAccount(params utils.RunParams, m *metrics.Metrics, system *ActorSystem) func(model.Snapshot, model.Account, Context) {
	return func(state model.Snapshot, meta model.Account, context Context) {
		snapshotHydration := utils.LoadSnapshot(params, meta.AccountName)
		metaHydration := utils.LoadMetadata(params, meta.AccountName)

		if snapshotHydration == nil || metaHydration == nil {
			context.Receiver.Become(state, meta, nonExistAccount(params, m, system))
			log.Debugf("%s ~ Nil -> NonExist", meta.AccountName)
		} else {
			context.Receiver.Become(*snapshotHydration, *metaHydration, existAccount(params, m, system))
			log.Debugf("%s ~ Nil -> Exist", meta.AccountName)
		}

		context.Receiver.Tell(context.Data, context.Sender)
	}
}

func nonExistAccount(params utils.RunParams, m *metrics.Metrics, system *ActorSystem) func(model.Snapshot, model.Account, Context) {
	return func(state model.Snapshot, meta model.Account, context Context) {
		switch msg := context.Data.(type) {

		case model.CreateAccount:
			currency := msg.Currency
			isBalanceCheck := msg.IsBalanceCheck

			// FIXME not ideal there could be case where snapshot was created but meta data not
			// and vice versa... should investigate what to do about that
			snaphostResult := utils.CreateSnapshot(params, meta.AccountName)
			metaResult := utils.CreateMetadata(params, meta.AccountName, currency, isBalanceCheck)

			if snaphostResult == nil || metaResult == nil {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (NonExist CreateAccount) Error", meta.AccountName)
				return
			}

			context.Receiver.Become(*snaphostResult, *metaResult, existAccount(params, m, system))
			system.SendRemote(context.Sender.Region, model.AccountCreatedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist CreateAccount) OK", meta.AccountName)
			m.AccountCreated()

		case model.Rollback:
			system.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist Rollback) OK", meta.AccountName)

		default:
			system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist Unknown Message) Error", meta.AccountName)

		}

		return
	}
}

func existAccount(params utils.RunParams, m *metrics.Metrics, system *ActorSystem) func(model.Snapshot, model.Account, Context) {
	return func(state model.Snapshot, meta model.Account, context Context) {
		switch msg := context.Data.(type) {

		case model.GetAccountState:
			system.SendRemote(context.Sender.Region, model.AccountStateMessage(context.Receiver.Name, context.Sender.Name, meta.Currency, state.Balance.String(), state.Promised.String(), meta.IsBalanceCheck))
			log.Debugf("%s ~ (Exist GetAccountState) OK", meta.AccountName)

		case model.CreateAccount:
			system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist CreateAccount) Error", meta.AccountName)

		case model.Promise:
			if state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.PromiseAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) OK Already Accepted", meta.AccountName)
				return
			}

			if meta.Currency != msg.Currency {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Promise) Error Currency Mismatch", meta.AccountName)
				return
			}

			nextPromised := new(money.Dec).Add(state.Promised, msg.Amount)

			if !meta.IsBalanceCheck || new(money.Dec).Add(state.Balance, nextPromised).Sign() >= 0 {
				if !utils.PersistPromise(params, meta.AccountName, state.Version, msg.Amount, msg.Transaction) {
					system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
					log.Warnf("%s ~ (Exist Promise) Error Could not Persist", meta.AccountName)
					return
				}

				next := state.Copy()
				next.Promised = nextPromised
				next.PromiseBuffer.Add(msg.Transaction)

				context.Receiver.Become(*next, meta, existAccount(params, m, system))

				system.SendRemote(context.Sender.Region, model.PromiseAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) OK", meta.AccountName)
				m.PromiseAccepted()
				return
			}

			if new(money.Dec).Sub(state.Balance, msg.Amount).Sign() < 0 {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) Error Insufficient Funds", meta.AccountName)
				return
			}

			// FIXME boucing not handled
			system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Promise) Error ... (Bounce?)", meta.AccountName)

		case model.Commit:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.CommitAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Commit) OK Already Accepted", meta.AccountName)
				return
			}

			if !utils.PersistCommit(params, meta.AccountName, state.Version, msg.Amount, msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Commit) Error Could not Persist", meta.AccountName)
				return
			}

			next := state.Copy()
			next.Balance = new(money.Dec).Add(state.Balance, msg.Amount)
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(*next, meta, existAccount(params, m, system))

			system.SendRemote(context.Sender.Region, model.CommitAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist Commit) OK", meta.AccountName)
			m.CommitAccepted()

		case model.Rollback:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Rollback) OK Already Accepted", meta.AccountName)
				return
			}

			if !utils.PersistRollback(params, meta.AccountName, state.Version, msg.Amount, msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Rollback) Error Could not Persist", meta.AccountName)
				return
			}

			next := state.Copy()
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(*next, meta, existAccount(params, m, system))

			system.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist Rollback) OK", meta.AccountName)
			m.RollbackAccepted()

		case model.Update:
			if msg.Version != state.Version {
				log.Debugf("%s ~ (Exist Update) Error Already Updated", meta.AccountName)
				return
			}

			result := utils.LoadSnapshot(params, meta.AccountName)
			if result == nil {
				log.Warnf("%s ~ (Exist Update) Error no existing snapshot", meta.AccountName)
				return
			}

			next := utils.UpdateSnapshot(params, meta.AccountName, result)
			if next == nil {
				log.Warnf("%s ~ (Exist Update) Error unable to update", meta.AccountName)
				return
			}

			context.Receiver.Become(*next, meta, existAccount(params, m, system))
			log.Debugf("%s ~ (Exist Update) OK", meta.AccountName)

		default:
			system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Unknown Message) Error", meta.AccountName)

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

	envelope := NewAccountEnvelope(path)
	err := system.RegisterActor(envelope, nilAccount(params, m, system))
	if err != nil {
		log.Warnf("%s ~ Spawning Actor Error unable to register", path)
		return ""
	}

	log.Debugf("%s ~ Actor Spawned", path)
	return envelope.Name
}

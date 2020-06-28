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
	"github.com/jancajthaml-openbank/vault-unit/model"
	"github.com/jancajthaml-openbank/vault-unit/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
	money "gopkg.in/inf.v0"
)

// NilAccount represents account that is neither existing neither non existing
func NilAccount(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		entity, err := persistence.LoadAccount(s.Storage, state.Name)
		if err != nil {
			context.Self.Become(state, NonExistAccount(s))
			log.WithField("account", state.Name).Debug("Nil -> NonExist")
		} else {
			context.Self.Become(*entity, ExistAccount(s))
			log.WithField("account", state.Name).Debug("Nil -> Exist")
		}

		context.Self.Receive(context)
	}
}

// NonExistAccount represents account that does not exist
func NonExistAccount(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		switch msg := context.Data.(type) {

		case CreateAccount:
			entity, err := persistence.CreateAccount(s.Storage, state.Name, msg.Format, msg.Currency, msg.IsBalanceCheck)
			if err != nil {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.WithField("account", state.Name).Warnf("(NonExist CreateAccount) Error %+v", err)
				return
			}

			s.SendMessage(RespCreateAccount, context.Sender, context.Receiver)

			s.Metrics.AccountCreated()

			log.WithField("account", state.Name).Info("Created")
			log.WithField("account", state.Name).Debug("(NonExist CreateAccount) OK")

			context.Self.Become(*entity, ExistAccount(s))

		case Rollback:
			s.SendMessage(RollbackAccepted, context.Sender, context.Receiver)
			log.WithField("account", state.Name).Debug("(NonExist Rollback) OK")

		case GetAccountState:
			s.SendMessage(RespAccountMissing, context.Sender, context.Receiver)
			log.WithField("account", state.Name).Debug("(NonExist GetAccountState) Error")

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("account", state.Name).Debug("(NonExist Unknown Message) Error")
		}

		return
	}
}

// ExistAccount represents account that does exist
func ExistAccount(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		switch msg := context.Data.(type) {

		case GetAccountState:
			s.SendMessage(AccountStateMessage(state), context.Sender, context.Receiver)
			log.WithField("account", state.Name).Debug("(Exist GetAccountState) OK")

		case CreateAccount:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("account", state.Name).Debug("(Exist CreateAccount) Error")

		case Promise:
			if state.Promises.Contains(msg.Transaction) {
				s.SendMessage(PromiseAccepted, context.Sender, context.Receiver)
				log.WithField("account", state.Name).Debug("(Exist Promise) OK Already Accepted")
				return
			}

			if state.Currency != msg.Currency {
				s.SendMessage(
					PromiseRejected+" CURRENCY_MISMATCH",
					context.Sender,
					context.Receiver,
				)
				log.WithField("account", state.Name).Debug("(Exist Promise) Error Currency Mismatch")
				return
			}

			next := state.Copy()
			next.Promised = new(money.Dec).Add(state.Promised, msg.Amount)
			next.Promises.Add(msg.Transaction)
			next.EventCounter = state.EventCounter + 1

			if !state.IsBalanceCheck || new(money.Dec).Add(state.Balance, next.Promised).Sign() >= 0 {
				err := persistence.PersistPromise(s.Storage, next, msg.Amount, msg.Transaction)
				if err != nil {
					s.SendMessage(
						PromiseRejected+" STORAGE_ERROR",
						context.Sender,
						context.Receiver,
					)
					log.WithField("account", state.Name).Warnf("Exist Promise) Error Could not Persist %+v", err)
					return
				}

				s.Metrics.PromiseAccepted()

				s.SendMessage(PromiseAccepted, context.Sender, context.Receiver)

				if state.EventCounter >= s.MaxEventsInSnapshot {
				updated, err := persistence.UpdateAccount(s.Storage, state.Name, &next)
				if err != nil {
					log.WithField("account", state.Name).Warnf("(Exist Promise) Error unable to update snapshot %+v", err)
				} else {
					next = *updated
					log.WithField("account", state.Name).Infof("(Exist Promise) Updated Snapshot to %d", next.SnapshotVersion)
				}
			}

				context.Self.Become(next, ExistAccount(s))
				log.WithField("account", state.Name).Infof("Promised %s %s", msg.Amount.String(), state.Currency)
				log.WithField("account", state.Name).Debug("(Exist Promise) OK")
				return
			}

			if new(money.Dec).Sub(state.Balance, msg.Amount).Sign() < 0 {
				s.SendMessage(
					PromiseRejected+" INSUFFICIESNT_FUNDS",
					context.Sender,
					context.Receiver,
				)
				log.WithField("account", state.Name).Debug("(Exist Promise) Error Insufficient Funds")
				return
			}

			// FIXME boucing not handled
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("account", state.Name).Warn("(Exist Promise) Error possible bounce")
			return

		case Commit:

			if !state.Promises.Contains(msg.Transaction) {
				s.SendMessage(CommitAccepted, context.Sender, context.Receiver)
				log.WithField("account", state.Name).Debug("(Exist Commit) OK Already Accepted")
				return
			}

			next := state.Copy()
			next.Balance = new(money.Dec).Add(state.Balance, msg.Amount)
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.Promises.Remove(msg.Transaction)
			next.EventCounter = state.EventCounter + 1

			err := persistence.PersistCommit(s.Storage, next, msg.Amount, msg.Transaction)
			if err != nil {
				s.SendMessage(
					CommitRejected+" STORAGE_ERROR",
					context.Sender,
					context.Receiver,
				)
				log.WithField("account", state.Name).Warnf("(Exist Commit) Error Could not Persist %+v", err)
				return
			}

			s.Metrics.CommitAccepted()

			s.SendMessage(CommitAccepted, context.Sender, context.Receiver)

			if state.EventCounter >= s.MaxEventsInSnapshot {
				updated, err := persistence.UpdateAccount(s.Storage, state.Name, &next)
				if err != nil {
					log.WithField("account", state.Name).Warnf("(Exist Commit) Error unable to update snapshot %+v", err)
				} else {
					next = *updated
					log.WithField("account", state.Name).Infof("(Exist Commit) Updated Snapshot to %d", next.SnapshotVersion)
				}
			}

			context.Self.Become(next, ExistAccount(s))
			log.WithField("account", state.Name).Debug("(Exist Commit) OK")
			return

		case Rollback:
			if !state.Promises.Contains(msg.Transaction) {
				s.SendMessage(RollbackAccepted, context.Sender, context.Receiver)
				log.WithField("account", state.Name).Debug("(Exist Rollback) OK Already Accepted")
				return
			}

			next := state.Copy()
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.Promises.Remove(msg.Transaction)
			next.EventCounter = state.EventCounter + 1

			err := persistence.PersistRollback(s.Storage, next, msg.Amount, msg.Transaction)
			if err != nil {
				s.SendMessage(
					RollbackRejected+" STORAGE_ERROR",
					context.Sender,
					context.Receiver,
				)
				log.WithField("account", state.Name).Warnf("(Exist Rollback) Error Could not Persist %+v", err)
				return
			}

			s.Metrics.RollbackAccepted()

			s.SendMessage(RollbackAccepted, context.Sender, context.Receiver)

			if state.EventCounter >= s.MaxEventsInSnapshot {
				updated, err := persistence.UpdateAccount(s.Storage, state.Name, &next)
				if err != nil {
					log.WithField("account", state.Name).Warnf("(Exist Rollback) Error unable to update snapshot %+v", err)
				} else {
					next = *updated
					log.WithField("account", state.Name).Infof("(Exist Rollback) Updated Snapshot to %d", next.SnapshotVersion)
				}
			}

			context.Self.Become(next, ExistAccount(s))
			log.WithField("account", state.Name).Infof("Rejected %s %s", msg.Amount.String(), state.Currency)
			log.WithField("account", state.Name).Debug("(Exist Rollback) OK")
			return

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("account", state.Name).Warn("(Exist Unknown Message) Error")

		}

		return
	}
}

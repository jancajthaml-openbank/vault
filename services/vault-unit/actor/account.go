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
func NilAccount(s *System) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		entity, err := persistence.LoadAccount(s.Storage, state.Name)
		if err != nil {
			context.Self.Become(state, NonExistAccount(s))
			log.Debug().Msgf("%s/Nil -> %s/NonExist", state.Name, state.Name)
		} else {
			context.Self.Become(*entity, ExistAccount(s))
			log.Debug().Msgf("%s/Nil -> %s/Exist", state.Name, state.Name)
		}

		context.Self.Receive(context)
	}
}

// NonExistAccount represents account that does not exist
func NonExistAccount(s *System) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		switch msg := context.Data.(type) {

		case CreateAccount:
			entity, err := persistence.CreateAccount(s.Storage, state.Name, msg.Format, msg.Currency, msg.IsBalanceCheck)
			if err != nil {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.Warn().Msgf("%s/NonExist/CreateAccount Error %+v", state.Name, err)
				return
			}

			s.SendMessage(RespCreateAccount, context.Sender, context.Receiver)

			s.Metrics.AccountCreated()

			log.Info().Msgf("Account %s created", state.Name)
			log.Debug().Msgf("%s/NonExist/CreateAccount OK", state.Name)

			context.Self.Become(*entity, ExistAccount(s))

		case Rollback:
			s.SendMessage(RollbackAccepted, context.Sender, context.Receiver)
			log.Debug().Msgf("%s/NonExist/Rollback OK", state.Name)

		case GetAccountState:
			s.SendMessage(RespAccountMissing, context.Sender, context.Receiver)
			log.Debug().Msgf("%s/NonExist/GetAccountState Error", state.Name)

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("%s/NonExist/Unknown Error", state.Name)
		}

		return
	}
}

// ExistAccount represents account that does exist
func ExistAccount(s *System) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		switch msg := context.Data.(type) {

		case GetAccountState:
			s.SendMessage(AccountStateMessage(state), context.Sender, context.Receiver)
			log.Debug().Msgf("%s/Exist/GetAccountState OK", state.Name)

		case CreateAccount:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("%s/Exist/CreateAccount Error", state.Name)

		case Promise:
			if state.Promises.Contains(msg.Transaction) {
				s.SendMessage(PromiseAccepted, context.Sender, context.Receiver)
				log.Debug().Msgf("%s/Exist/Promise OK Already Accepted", state.Name)
				return
			}

			if state.Currency != msg.Currency {
				s.SendMessage(
					PromiseRejected+" CURRENCY_MISMATCH",
					context.Sender,
					context.Receiver,
				)
				log.Debug().Msgf("%s/Exist/Promise Error Currency Mismatch", state.Name)
				return
			}

			//next := state.Copy()
			state.Promised = new(money.Dec).Add(state.Promised, msg.Amount)
			state.Promises.Add(msg.Transaction)
			state.EventCounter = state.EventCounter + 1

			if !state.IsBalanceCheck || new(money.Dec).Add(state.Balance, state.Promised).Sign() >= 0 {
				err := persistence.PersistPromise(s.Storage, state, msg.Amount, msg.Transaction)
				if err != nil {
					s.SendMessage(
						PromiseRejected+" STORAGE_ERROR",
						context.Sender,
						context.Receiver,
					)
					log.Warn().Msgf("%s/Exist/Promise Error could not persist %+v", state.Name, err)
					state.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
					state.Promises.Remove(msg.Transaction)
					state.EventCounter = state.EventCounter - 1
					return
				}

				s.Metrics.PromiseAccepted()

				s.SendMessage(PromiseAccepted, context.Sender, context.Receiver)

				if state.EventCounter >= s.EventCounterTreshold {
					err := persistence.UpdateAccount(s.Storage, state.Name, &state)
					if err != nil {
						log.Warn().Msgf("%s/Exist/Promise Error unable to update snapshot %+v", state.Name, err)
					} else {
						state.SnapshotVersion += 1
						state.EventCounter = 0
						log.Info().Msgf("%s/Exist/Promise Updated snapshot to version %d", state.Name, state.SnapshotVersion)
					}
				}

				context.Self.Become(state, ExistAccount(s))
				log.Info().Msgf("Account %s promised %s %s", state.Name, msg.Amount.String(), state.Currency)
				log.Debug().Msgf("%s/Exist/Promise OK", state.Name)
				return
			}

			if new(money.Dec).Sub(state.Balance, msg.Amount).Sign() < 0 {
				s.SendMessage(
					PromiseRejected+" INSUFFICIESNT_FUNDS",
					context.Sender,
					context.Receiver,
				)
				log.Debug().Msgf("%s/Exist/Promise Error insufficient funds", state.Name)
				return
			}

			// FIXME boucing not handled
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Warn().Msgf("%s/Exist/Promise Error possible bounce", state.Name)
			return

		case Commit:

			if !state.Promises.Contains(msg.Transaction) {
				s.SendMessage(CommitAccepted, context.Sender, context.Receiver)
				log.Debug().Msgf("%s/Exist/Commit OK already accepted", state.Name)
				return
			}

			state.Balance = new(money.Dec).Add(state.Balance, msg.Amount)
			state.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			state.Promises.Remove(msg.Transaction)
			state.EventCounter = state.EventCounter + 1

			err := persistence.PersistCommit(s.Storage, state, msg.Amount, msg.Transaction)
			if err != nil {
				s.SendMessage(
					CommitRejected+" STORAGE_ERROR",
					context.Sender,
					context.Receiver,
				)
				log.Warn().Msgf("%s/Exist/Commit Error could not persist %+v", state.Name, err)
				state.Balance = new(money.Dec).Sub(state.Balance, msg.Amount)
				state.Promised = new(money.Dec).Add(state.Promised, msg.Amount)
				state.Promises.Add(msg.Transaction)
				state.EventCounter = state.EventCounter - 1
				return
			}

			s.Metrics.CommitAccepted()

			s.SendMessage(CommitAccepted, context.Sender, context.Receiver)

			if state.EventCounter >= s.EventCounterTreshold {
				err := persistence.UpdateAccount(s.Storage, state.Name, &state)
				if err != nil {
					log.Warn().Msgf("%s/Exist/Commit Error unable to update snapshot %+v", state.Name, err)
				} else {
					state.SnapshotVersion += 1
					state.EventCounter = 0
					log.Info().Msgf("%s/Exist/Commit Updated snapshot to version %d", state.Name, state.SnapshotVersion)
				}
			}

			context.Self.Become(state, ExistAccount(s))
			log.Debug().Msgf("%s/Exist/Commit OK", state.Name)
			return

		case Rollback:
			if !state.Promises.Contains(msg.Transaction) {
				s.SendMessage(RollbackAccepted, context.Sender, context.Receiver)
				log.Debug().Msgf("%s/Exist/Rollback OK Already Accepted", state.Name)
				return
			}

			state.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			state.Promises.Remove(msg.Transaction)
			state.EventCounter = state.EventCounter + 1

			err := persistence.PersistRollback(s.Storage, state, msg.Amount, msg.Transaction)
			if err != nil {
				s.SendMessage(
					RollbackRejected+" STORAGE_ERROR",
					context.Sender,
					context.Receiver,
				)
				log.Warn().Msgf("%s/Exist/Rollback Error could not persist %+v", state.Name, err)
				state.Promised = new(money.Dec).Add(state.Promised, msg.Amount)
				state.Promises.Add(msg.Transaction)
				state.EventCounter = state.EventCounter - 1
				return
			}

			s.Metrics.RollbackAccepted()

			s.SendMessage(RollbackAccepted, context.Sender, context.Receiver)

			if state.EventCounter >= s.EventCounterTreshold {
				err := persistence.UpdateAccount(s.Storage, state.Name, &state)
				if err != nil {
					log.Warn().Msgf("%s/Exist/Rollback Error unable to update snapshot %+v", state.Name, err)
				} else {
					state.SnapshotVersion += 1
						state.EventCounter = 0
					log.Info().Msgf("%s/Exist/Rollback Updated snapshot to version %d", state.Name, state.SnapshotVersion)
				}
			}

			context.Self.Become(state, ExistAccount(s))
			log.Info().Msgf("Account %s rejected %s %s", state.Name, msg.Amount.String(), state.Currency)
			log.Debug().Msgf("%s/Exist/Rollback OK", state.Name)
			return

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Warn().Msgf("%s/Exist/Unknown Error", state.Name)

		}

		return
	}
}

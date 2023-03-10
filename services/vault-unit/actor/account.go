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
	"github.com/jancajthaml-openbank/vault-unit/model"
	"github.com/jancajthaml-openbank/vault-unit/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
)

// NilAccount represents account that is neither existing neither non existing
func NilAccount(s *System, state model.Account) system.ReceiverFunction {
	return func(context system.Context) system.ReceiverFunction {
		context.Self.Tell(context.Data, context.Receiver, context.Sender)
		entity, err := persistence.LoadAccount(s.Storage, state.Name)
		if err != nil {
			log.Debug().Msgf("%s/Nil -> %s/NonExist", state.Name, state.Name)
			return NonExistAccount(s, state)
		} else {
			log.Debug().Msgf("%s/Nil -> %s/Exist", state.Name, state.Name)
			return ExistAccount(s, *entity)
		}
	}
}

// NonExistAccount represents account that does not exist
func NonExistAccount(s *System, state model.Account) system.ReceiverFunction {
	return func(context system.Context) system.ReceiverFunction {

		switch msg := context.Data.(type) {

		case CreateAccount:
			entity, err := persistence.CreateAccount(s.Storage, state.Name, msg.Format, msg.Currency, msg.IsBalanceCheck)
			if err != nil {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.Warn().Err(err).Msgf("%s/NonExist/CreateAccount Error", state.Name)
				return NonExistAccount(s, state)
			}
			s.SendMessage(RespCreateAccount, context.Sender, context.Receiver)
			s.Metrics.AccountCreated()
			log.Info().Msgf("Account %s created", state.Name)
			log.Debug().Msgf("%s/NonExist/CreateAccount OK", state.Name)
			return ExistAccount(s, *entity)

		case Rollback:
			s.SendMessage(RollbackAccepted, context.Sender, context.Receiver)
			log.Debug().Msgf("%s/NonExist/Rollback OK", state.Name)
			return NonExistAccount(s, state)

		case GetAccountState:
			s.SendMessage(RespAccountMissing, context.Sender, context.Receiver)
			log.Debug().Msgf("%s/NonExist/GetAccountState Error", state.Name)
			return NonExistAccount(s, state)

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("%s/NonExist/Unknown Error", state.Name)
			return NonExistAccount(s, state)
		}

	}
}

// ExistAccount represents account that does exist
func ExistAccount(s *System, state model.Account) system.ReceiverFunction {
	return func(context system.Context) system.ReceiverFunction {

		switch msg := context.Data.(type) {

		case GetAccountState:
			s.SendMessage(AccountStateMessage(state), context.Sender, context.Receiver)
			log.Debug().Msgf("%s/Exist/GetAccountState OK", state.Name)
			return ExistAccount(s, state)

		case CreateAccount:
			if (msg.Format == state.Format && msg.Currency == state.Currency && msg.IsBalanceCheck == state.IsBalanceCheck) {
				s.SendMessage(RespCreateAccount, context.Sender, context.Receiver)
				log.Debug().Msgf("%s/Exist/CreateAccount Already Exist", state.Name)
				return ExistAccount(s, state)
			}
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("%s/Exist/CreateAccount Error", state.Name)
			return ExistAccount(s, state)

		case Promise:
			promiseHash := msg.Transaction + "_" + msg.Currency + "_" + msg.Amount.String()

			if state.Promises.Contains(promiseHash) {
				s.SendMessage(PromiseAccepted, context.Sender, context.Receiver)
				log.Debug().Msgf("%s/Exist/Promise OK Already Accepted", state.Name)
				return ExistAccount(s, state)
			}

			if state.Currency != msg.Currency {
				s.SendMessage(
					PromiseRejected+" CURRENCY_MISMATCH",
					context.Sender,
					context.Receiver,
				)
				log.Debug().Msgf("%s/Exist/Promise Error Currency Mismatch", state.Name)
				return ExistAccount(s, state)
			}

			nextBalance := new(model.Dec)
			nextBalance.Add(&state.Balance)
			nextBalance.Add(&state.Promised)
			nextBalance.Add(msg.Amount)

			state.Promised.Add(msg.Amount)
			state.Promises.Add(promiseHash)
			state.EventCounter++

			if !state.IsBalanceCheck || nextBalance.Sign() >= 0 {

				err := persistence.PersistPromise(s.Storage, state, msg.Amount, msg.Transaction)
				if err != nil {
					s.SendMessage(
						PromiseRejected+" STORAGE_ERROR",
						context.Sender,
						context.Receiver,
					)
					log.Warn().Err(err).Msgf("%s/Exist/Promise Error could not persist", state.Name)
					state.Promised.Sub(msg.Amount)
					state.Promises.Remove(promiseHash)
					state.EventCounter--
					return ExistAccount(s, state)
				}

				s.SendMessage(PromiseAccepted, context.Sender, context.Receiver)
				s.Metrics.PromiseAccepted()

				if state.EventCounter >= s.EventCounterTreshold {
					err := persistence.UpdateAccount(s.Storage, state.Name, &state)
					if err != nil {
						log.Warn().Err(err).Msgf("%s/Exist/Promise Error unable to update snapshot", state.Name)
					} else {
						state.SnapshotVersion++
						state.EventCounter = 0
						log.Info().Msgf("%s/Exist/Promise Updated snapshot to version %d", state.Name, state.SnapshotVersion)
					}
				}

				//context.Self.Become(state, ExistAccount(s))
				log.Info().Msgf("Account %s promised %s %s", state.Name, msg.Amount.String(), state.Currency)
				log.Debug().Msgf("%s/Exist/Promise OK", state.Name)
				return ExistAccount(s, state)
			}

			nextBalance = new(model.Dec)
			nextBalance.Add(&state.Balance)
			nextBalance.Add(msg.Amount)

			if nextBalance.Sign() < 0 {
				s.SendMessage(
					PromiseRejected+" INSUFFICIENT_FUNDS",
					context.Sender,
					context.Receiver,
				)
				log.Debug().Msgf("%s/Exist/Promise Error insufficient funds", state.Name)
				return ExistAccount(s, state)
			}

			s.SendMessage(
				PromiseBounced,
				context.Sender,
				context.Receiver,
			)

			log.Warn().Msgf("%s/Exist/Promise bounced", state.Name)
			return ExistAccount(s, state)

		case Commit:

			promiseHash := msg.Transaction + "_" + msg.Currency + "_" + msg.Amount.String()

			if !state.Promises.Contains(promiseHash) {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.Debug().Msgf("%s/Exist/Commit Error unknown promise to commit", state.Name)
				return ExistAccount(s, state)
			}

			state.Balance.Add(msg.Amount)
			state.Promised.Sub(msg.Amount)
			state.Promises.Remove(promiseHash)
			state.EventCounter++

			err := persistence.PersistCommit(s.Storage, state, msg.Amount, msg.Transaction)
			if err != nil {
				s.SendMessage(
					CommitRejected+" STORAGE_ERROR",
					context.Sender,
					context.Receiver,
				)
				log.Warn().Err(err).Msgf("%s/Exist/Commit Error could not persist", state.Name)
				state.Balance.Sub(msg.Amount)
				state.Promised.Add(msg.Amount)
				state.Promises.Add(promiseHash)
				state.EventCounter--
				return ExistAccount(s, state)
			}

			s.SendMessage(CommitAccepted, context.Sender, context.Receiver)
			s.Metrics.CommitAccepted()

			if state.EventCounter >= s.EventCounterTreshold {
				err := persistence.UpdateAccount(s.Storage, state.Name, &state)
				if err != nil {
					log.Warn().Err(err).Msgf("%s/Exist/Commit Error unable to update snapshot", state.Name)
				} else {
					state.SnapshotVersion++
					state.EventCounter = 0
					log.Info().Msgf("%s/Exist/Commit Updated snapshot to version %d", state.Name, state.SnapshotVersion)
				}
			}

			log.Debug().Msgf("%s/Exist/Commit OK", state.Name)
			return ExistAccount(s, state)

		case Rollback:

			promiseHash := msg.Transaction + "_" + msg.Currency + "_" + msg.Amount.String()

			if !state.Promises.Contains(promiseHash) {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.Debug().Msgf("%s/Exist/Rollback Error unknown promise to rollback", state.Name)
				return ExistAccount(s, state)
			}

			state.Promised.Sub(msg.Amount)
			state.Promises.Remove(promiseHash)
			state.EventCounter++

			err := persistence.PersistRollback(s.Storage, state, msg.Amount, msg.Transaction)
			if err != nil {
				s.SendMessage(
					RollbackRejected+" STORAGE_ERROR",
					context.Sender,
					context.Receiver,
				)
				log.Warn().Err(err).Msgf("%s/Exist/Rollback Error could not persist", state.Name)
				state.Promised.Add(msg.Amount)
				state.Promises.Add(promiseHash)
				state.EventCounter--
				return ExistAccount(s, state)
			}

			s.SendMessage(RollbackAccepted, context.Sender, context.Receiver)
			s.Metrics.RollbackAccepted()

			if state.EventCounter >= s.EventCounterTreshold {
				err := persistence.UpdateAccount(s.Storage, state.Name, &state)
				if err != nil {
					log.Warn().Err(err).Msgf("%s/Exist/Rollback Error unable to update snapshot", state.Name)
				} else {
					state.SnapshotVersion++
					state.EventCounter = 0
					log.Info().Msgf("%s/Exist/Rollback Updated snapshot to version %d", state.Name, state.SnapshotVersion)
				}
			}

			log.Info().Msgf("Account %s rejected %s %s", state.Name, msg.Amount.String(), state.Currency)
			log.Debug().Msgf("%s/Exist/Rollback OK", state.Name)
			return ExistAccount(s, state)

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Warn().Msgf("%s/Exist/Unknown Error", state.Name)
			return ExistAccount(s, state)

		}

	}
}

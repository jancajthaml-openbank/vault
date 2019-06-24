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
	"strings"

	"github.com/jancajthaml-openbank/vault-unit/model"
	"github.com/jancajthaml-openbank/vault-unit/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

// NilAccount represents account that is neither existing neither non existing
func NilAccount(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		snapshotHydration := persistence.LoadAccount(s.Storage, state.Name)

		if snapshotHydration == nil {
			context.Self.Become(state, NonExistAccount(s))
			log.Debugf("%s ~ Nil -> NonExist", state.Name)
		} else {
			context.Self.Become(*snapshotHydration, ExistAccount(s))
			log.Debugf("%s ~ Nil -> Exist", state.Name)
		}

		context.Self.Receive(context)
	}
}

// NonExistAccount represents account that does not exist
func NonExistAccount(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		switch msg := context.Data.(type) {

		case model.CreateAccount:
			currency := strings.ToUpper(msg.Currency)
			isBalanceCheck := msg.IsBalanceCheck
			format := strings.ToUpper(msg.Format)

			snaphostResult := persistence.CreateAccount(s.Storage, state.Name, format, currency, isBalanceCheck)

			if snaphostResult == nil {
				s.SendRemote(FatalErrorMessage(context))
				log.Debugf("%s ~ (NonExist CreateAccount) Error", state.Name)
				return
			}

			s.Metrics.AccountCreated()

			s.SendRemote(AccountCreatedMessage(context))

			context.Self.Become(*snaphostResult, ExistAccount(s))
			log.Infof("New Account %s Created", state.Name)
			log.Debugf("%s ~ (NonExist CreateAccount) OK", state.Name)

		case model.Rollback:
			s.SendRemote(RollbackAcceptedMessage(context))
			log.Debugf("%s ~ (NonExist Rollback) OK", state.Name)

		case model.GetAccountState:
			s.SendRemote(AccountMissingMessage(context))
			log.Debugf("%s ~ (NonExist GetAccountState) Error", state.Name)

		default:
			s.SendRemote(FatalErrorMessage(context))
			log.Debugf("%s ~ (NonExist Unknown Message) Error", state.Name)
		}

		return
	}
}

// ExistAccount represents account that does exist
func ExistAccount(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		switch msg := context.Data.(type) {

		case model.GetAccountState:
			s.SendRemote(AccountStateMessage(context, state))
			log.Debugf("%s ~ (Exist GetAccountState) OK", state.Name)

		case model.CreateAccount:
			s.SendRemote(FatalErrorMessage(context))
			log.Debugf("%s ~ (Exist CreateAccount) Error", state.Name)

		case model.Promise:
			if state.PromiseBuffer.Contains(msg.Transaction) {
				s.SendRemote(PromiseAcceptedMessage(context))
				log.Debugf("%s ~ (Exist Promise) OK Already Accepted", state.Name)
				return
			}

			if state.Currency != msg.Currency {
				s.SendRemote(PromiseRejectedMessage(context, "CURRENCY_MISMATCH"))
				log.Warnf("%s ~ (Exist Promise) Error Currency Mismatch", state.Name)
				return
			}

			nextPromised := new(money.Dec).Add(state.Promised, msg.Amount)

			if !state.IsBalanceCheck || new(money.Dec).Add(state.Balance, nextPromised).Sign() >= 0 {
				if err := persistence.PersistPromise(s.Storage, state.Name, state.Version, msg.Amount, msg.Transaction); err != nil {
					s.SendRemote(PromiseRejectedMessage(context, "STORAGE_ERROR"))
					log.Warnf("%s ~ (Exist Promise) Error Could not Persist %+v", state.Name, err)
					return
				}

				next := state.Copy()
				next.Promised = nextPromised
				next.PromiseBuffer.Add(msg.Transaction)

				s.Metrics.PromiseAccepted()

				s.SendRemote(PromiseAcceptedMessage(context))

				context.Self.Become(next, ExistAccount(s))
				log.Infof("Account %s Promised %s %s", state.Name, msg.Amount.String(), state.Currency)
				log.Debugf("%s ~ (Exist Promise) OK", state.Name)
				return
			}

			if new(money.Dec).Sub(state.Balance, msg.Amount).Sign() < 0 {
				s.SendRemote(PromiseRejectedMessage(context, "INSUFFICIESNT_FUNDS"))
				log.Debugf("%s ~ (Exist Promise) Error Insufficient Funds", state.Name)
				return
			}

			// FIXME boucing not handled
			s.SendRemote(FatalErrorMessage(context))
			log.Warnf("%s ~ (Exist Promise) Error ... (Bounce?)", state.Name)
			return

		case model.Commit:

			if !state.PromiseBuffer.Contains(msg.Transaction) {
				s.SendRemote(CommitAcceptedMessage(context))
				log.Debugf("%s ~ (Exist Commit) OK Already Accepted", state.Name)
				return
			}

			if err := persistence.PersistCommit(s.Storage, state.Name, state.Version, msg.Amount, msg.Transaction); err != nil {
				s.SendRemote(CommitRejectedMessage(context, "STORAGE_ERROR"))
				log.Warnf("%s ~ (Exist Commit) Error Could not Persist %+v", state.Name, err)
				return
			}

			next := state.Copy()
			next.Balance = new(money.Dec).Add(state.Balance, msg.Amount)
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			s.Metrics.CommitAccepted()

			s.SendRemote(CommitAcceptedMessage(context))

			context.Self.Become(next, ExistAccount(s))
			log.Debugf("%s ~ (Exist Commit) OK", state.Name)
			return

		case model.Rollback:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				s.SendRemote(RollbackAcceptedMessage(context))
				log.Debugf("%s ~ (Exist Rollback) OK Already Accepted", state.Name)
				return
			}

			if err := persistence.PersistRollback(s.Storage, state.Name, state.Version, msg.Amount, msg.Transaction); err != nil {
				s.SendRemote(RollbackRejectedMessage(context, "STORAGE_ERROR"))
				log.Warnf("%s ~ (Exist Rollback) Error Could not Persist %+v", state.Name, err)
				return
			}

			next := state.Copy()
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			s.Metrics.RollbackAccepted()

			s.SendRemote(RollbackAcceptedMessage(context))

			context.Self.Become(next, ExistAccount(s))
			log.Infof("Account %s Rejected %s %s", state.Name, msg.Amount.String(), state.Currency)
			log.Debugf("%s ~ (Exist Rollback) OK", state.Name)
			return

		case model.Update:
			if msg.Version != state.Version {
				log.Debugf("%s ~ (Exist Update) Error Already Updated", state.Name)
				return
			}

			result := persistence.LoadAccount(s.Storage, state.Name)
			if result == nil {
				log.Warnf("%s ~ (Exist Update) Error no existing snapshot", state.Name)
				return
			}

			next := persistence.UpdateAccount(s.Storage, state.Name, result)
			if next == nil {
				log.Warnf("%s ~ (Exist Update) Error unable to update", state.Name)
				return
			}

			context.Self.Become(*next, ExistAccount(s))
			log.Infof("Account %s Updated Snapshot to %d", state.Name, next.Version)
			log.Debugf("%s ~ (Exist Update) OK", state.Name)

		default:
			s.SendRemote(FatalErrorMessage(context))
			log.Warnf("%s ~ (Exist Unknown Message) Error", state.Name)

		}

		return
	}
}

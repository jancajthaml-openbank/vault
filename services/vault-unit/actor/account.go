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
	"strings"

	"github.com/jancajthaml-openbank/vault-unit/daemon"
	"github.com/jancajthaml-openbank/vault-unit/model"
	"github.com/jancajthaml-openbank/vault-unit/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

// NilAccount represents account that is neither existing neither non existing
func NilAccount(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		snapshotHydration := persistence.LoadAccount(s.Storage, state.AccountName)

		if snapshotHydration == nil {
			context.Receiver.Become(state, NonExistAccount(s))
			log.Debugf("%s ~ Nil -> NonExist", state.AccountName)
		} else {
			context.Receiver.Become(*snapshotHydration, ExistAccount(s))
			log.Debugf("%s ~ Nil -> Exist", state.AccountName)
		}

		context.Receiver.Tell(context.Data, context.Sender)
	}
}

// NonExistAccount represents account that does not exist
func NonExistAccount(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		switch msg := context.Data.(type) {

		case model.CreateAccount:
			currency := strings.ToUpper(msg.Currency)
			isBalanceCheck := msg.IsBalanceCheck

			snaphostResult := persistence.CreateAccount(s.Storage, state.AccountName, currency, isBalanceCheck)

			if snaphostResult == nil {
				s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (NonExist CreateAccount) Error", state.AccountName)
				return
			}

			context.Receiver.Become(*snaphostResult, ExistAccount(s))
			s.SendRemote(context.Sender.Region, AccountCreatedMessage(context.Receiver.Name, context.Sender.Name))
			log.Infof("New Account %s Created", state.AccountName)
			log.Debugf("%s ~ (NonExist CreateAccount) OK", state.AccountName)
			s.Metrics.AccountCreated()

		case model.Rollback:
			s.SendRemote(context.Sender.Region, RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist Rollback) OK", state.AccountName)

		default:
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist Unknown Message) Error", state.AccountName)
		}

		return
	}
}

// ExistAccount represents account that does exist
func ExistAccount(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		switch msg := context.Data.(type) {

		case model.GetAccountState:
			s.SendRemote(context.Sender.Region, AccountStateMessage(context.Receiver.Name, context.Sender.Name, state.Currency, state.Balance.String(), state.Promised.String(), state.IsBalanceCheck))
			log.Debugf("%s ~ (Exist GetAccountState) OK", state.AccountName)

		case model.CreateAccount:
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist CreateAccount) Error", state.AccountName)

		case model.Promise:
			if state.PromiseBuffer.Contains(msg.Transaction) {
				s.SendRemote(context.Sender.Region, PromiseAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) OK Already Accepted", state.AccountName)
				return
			}

			if state.Currency != msg.Currency {
				s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Promise) Error Currency Mismatch", state.AccountName)
				return
			}

			nextPromised := new(money.Dec).Add(state.Promised, msg.Amount)

			if !state.IsBalanceCheck || new(money.Dec).Add(state.Balance, nextPromised).Sign() >= 0 {
				if !persistence.PersistPromise(s.Storage, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
					s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
					log.Warnf("%s ~ (Exist Promise) Error Could not Persist", state.AccountName)
					return
				}

				next := state.Copy()
				next.Promised = nextPromised
				next.PromiseBuffer.Add(msg.Transaction)

				context.Receiver.Become(next, ExistAccount(s))

				s.SendRemote(context.Sender.Region, PromiseAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Infof("Account %s Promised %s %s", state.AccountName, msg.Amount.String(), state.Currency)
				log.Debugf("%s ~ (Exist Promise) OK", state.AccountName)
				s.Metrics.PromiseAccepted()
				return
			}

			if new(money.Dec).Sub(state.Balance, msg.Amount).Sign() < 0 {
				s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) Error Insufficient Funds", state.AccountName)
				return
			}

			// FIXME boucing not handled
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Promise) Error ... (Bounce?)", state.AccountName)

		case model.Commit:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				s.SendRemote(context.Sender.Region, CommitAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Commit) OK Already Accepted", state.AccountName)
				return
			}

			if !persistence.PersistCommit(s.Storage, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
				s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Commit) Error Could not Persist", state.AccountName)
				return
			}

			next := state.Copy()
			next.Balance = new(money.Dec).Add(state.Balance, msg.Amount)
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(next, ExistAccount(s))

			s.SendRemote(context.Sender.Region, CommitAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist Commit) OK", state.AccountName)
			s.Metrics.CommitAccepted()

		case model.Rollback:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				s.SendRemote(context.Sender.Region, RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Rollback) OK Already Accepted", state.AccountName)
				return
			}

			if !persistence.PersistRollback(s.Storage, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
				s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Rollback) Error Could not Persist", state.AccountName)
				return
			}

			next := state.Copy()
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(next, ExistAccount(s))

			s.SendRemote(context.Sender.Region, RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Infof("Account %s Rejected %s %s", state.AccountName, msg.Amount.String(), state.Currency)
			log.Debugf("%s ~ (Exist Rollback) OK", state.AccountName)
			s.Metrics.RollbackAccepted()

		case model.Update:
			if msg.Version != state.Version {
				log.Debugf("%s ~ (Exist Update) Error Already Updated", state.AccountName)
				return
			}

			result := persistence.LoadAccount(s.Storage, state.AccountName)
			if result == nil {
				log.Warnf("%s ~ (Exist Update) Error no existing snapshot", state.AccountName)
				return
			}

			next := persistence.UpdateAccount(s.Storage, state.AccountName, result)
			if next == nil {
				log.Warnf("%s ~ (Exist Update) Error unable to update", state.AccountName)
				return
			}

			context.Receiver.Become(*next, ExistAccount(s))
			log.Infof("Account %s Updated Snapshot to %d", state.AccountName, next.Version)
			log.Debugf("%s ~ (Exist Update) OK", state.AccountName)

		default:
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Unknown Message) Error", state.AccountName)

		}

		return
	}
}

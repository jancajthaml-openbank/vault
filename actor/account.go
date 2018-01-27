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
	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/utils"

	money "gopkg.in/inf.v0"
)

func nilAccount(params utils.RunParams, system *ActorSystem) func(model.Snapshot, model.Account, Context) {
	return func(state model.Snapshot, meta model.Account, context Context) {
		snapshotHydration := utils.LoadSnapshot(params, meta.AccountName)
		metaHydration := utils.LoadMetadata(params, meta.AccountName)

		if snapshotHydration == nil || metaHydration == nil {
			context.Reciever.Become(state, meta, nonExistAccount(params, system))
		} else {
			context.Reciever.Become(*snapshotHydration, *metaHydration, existAccount(params, system))
		}

		context.Reciever.Tell(context.Data, context.Sender)
	}
}

func nonExistAccount(params utils.RunParams, system *ActorSystem) func(model.Snapshot, model.Account, Context) {
	return func(state model.Snapshot, meta model.Account, context Context) {
		switch msg := context.Data.(type) {

		case model.CreateAccount:
			currency := msg.Currency
			isBalanceCheck := msg.IsBalanceCheck

			snaphostResult := utils.CreateSnapshot(params, meta.AccountName)
			metaResult := utils.CreateMetadata(params, meta.AccountName, currency, isBalanceCheck)

			if snaphostResult == nil || metaResult == nil {
				system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))
			} else {
				context.Reciever.Become(*snaphostResult, *metaResult, existAccount(params, system))
				system.SendRemote("Server", model.AccountCreatedMessage(context.Reciever.Name, context.Sender))
			}

		default:
			system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))

		}

		return
	}
}

func existAccount(params utils.RunParams, system *ActorSystem) func(model.Snapshot, model.Account, Context) {
	return func(state model.Snapshot, meta model.Account, context Context) {
		switch msg := context.Data.(type) {

		case model.GetAccountBalance:
			system.SendRemote("Server", model.AccountBalanceMessage(context.Reciever.Name, context.Sender, meta.Currency, state.Balance.String()))

		case model.CreateAccount:
			system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))

		case model.Promise:
			if state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote("Server", model.PromiseAcceptedMessage(context.Reciever.Name, context.Sender))
				return
			}

			if meta.Currency != msg.Currency {
				system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))
				return
			}

			nextPromised := new(money.Dec).Add(state.Promised, msg.Amount)

			if !meta.IsBalanceCheck || new(money.Dec).Add(state.Balance, nextPromised).Sign() >= 0 {
				if !utils.PersistPromise(params, meta.AccountName, state.Version, msg.Amount, msg.Transaction) {
					system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))
					return
				}

				next := state.Copy()
				next.Promised = nextPromised
				next.PromiseBuffer.Add(msg.Transaction)

				context.Reciever.Become(*next, meta, existAccount(params, system))

				system.SendRemote("Server", model.PromiseAcceptedMessage(context.Reciever.Name, context.Sender))
				return
			}

			if new(money.Dec).Sub(state.Balance, msg.Amount).Sign() < 0 {
				system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))
				return
			}

			// FIXME boucing not handled
			system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))

		case model.Commit:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote("Server", model.CommitAcceptedMessage(context.Reciever.Name, context.Sender))
				return
			}

			if !utils.PersistCommit(params, meta.AccountName, state.Version, msg.Amount, msg.Transaction) {
				system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))
				return
			}

			next := state.Copy()
			next.Balance = new(money.Dec).Add(state.Balance, msg.Amount)
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Reciever.Become(*next, meta, existAccount(params, system))

			system.SendRemote("Server", model.CommitAcceptedMessage(context.Reciever.Name, context.Sender))

		case model.Rollback:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote("Server", model.RollbackAcceptedMessage(context.Reciever.Name, context.Sender))
				return
			}

			if !utils.PersistRollback(params, meta.AccountName, state.Version, msg.Amount, msg.Transaction) {
				system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))
				return
			}

			next := state.Copy()
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Reciever.Become(*next, meta, existAccount(params, system))

			system.SendRemote("Server", model.RollbackAcceptedMessage(context.Reciever.Name, context.Sender))

		case model.Update:
			if msg.Version != state.Version {
				return
			}

			result := utils.LoadSnapshot(params, meta.AccountName)
			if result == nil {
				return
			}

			next := utils.UpdateSnapshot(params, meta.AccountName, result)
			if next == nil {
				return
			}

			context.Reciever.Become(*next, meta, existAccount(params, system))

		default:
			system.SendRemote("Server", model.FatalErrorMessage(context.Reciever.Name, context.Sender))
		}

		return
	}
}

// FIXME split to multiple functions
// SpawnAccountActor returns new account actor instance registered into actor
// system
func (system *ActorSystem) SpawnAccountActor(params utils.RunParams, path string) string {
	if system == nil {
		// FIXME check for len(x) == 0
		return ""
	}

	accountEnvelope := NewAccountEnvelope(path)
	system.RegisterActor(accountEnvelope, nilAccount(params, system))

	return accountEnvelope.Name
}

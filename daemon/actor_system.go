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

package daemon

import (
	"context"
	"strings"

	"github.com/jancajthaml-openbank/vault/config"
	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

// ActorSystem represents actor system subroutine
type ActorSystem struct {
	system.Support
	tenant  string
	storage string
	metrics *Metrics
}

// NewActorSystem returns actor system fascade
func NewActorSystem(ctx context.Context, cfg config.Configuration, metrics *Metrics) ActorSystem {
	actorSystem := ActorSystem{
		Support: system.NewSupport(ctx, "Vault/"+cfg.Tenant, cfg.LakeHostname),
		storage: cfg.RootStorage,
		tenant:  cfg.Tenant,
		metrics: metrics,
	}
	actorSystem.Support.RegisterOnLocalMessage(actorSystem.ProcessLocalMessage)
	actorSystem.Support.RegisterOnRemoteMessage(actorSystem.ProcessRemoteMessage)
	return actorSystem
}

// ProcessLocalMessage processing of local message to this vault
func (s *ActorSystem) ProcessLocalMessage(msg interface{}, receiver string, sender system.Coordinates) {
	ref, err := s.ActorOf(receiver)
	if err != nil {
		ref, err = s.ActorOf(s.SpawnAccountActor(receiver))
	}

	if err != nil {
		log.Warnf("Actor not found [%s local]", receiver)
		return
	}
	ref.Tell(msg, sender)
}

// ProcessRemoteMessage processing of remote message to this vault
func (s *ActorSystem) ProcessRemoteMessage(parts []string) {
	if len(parts) < 4 {
		log.Warnf("invalid message received %+v", parts)
		return
	}

	region, receiver, sender, payload := parts[0], parts[1], parts[2], parts[3]

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("procesRemoteMessage recovered in [%s %s/%s] : %v", r, receiver, region, sender)
			s.SendRemote(region, model.FatalErrorMessage(receiver, sender))
		}
	}()

	ref, err := s.ActorOf(receiver)
	if err != nil {
		ref, err = s.ActorOf(s.SpawnAccountActor(receiver))
	}

	if err != nil {
		log.Warnf("Actor not found [%s %s/%s]", receiver, region, sender)
		s.SendRemote(region, model.FatalErrorMessage(receiver, sender))
		return
	}

	var message interface{}

	switch payload {

	case model.ReqAccountState:
		message = model.GetAccountState{}

	case model.ReqCreateAccount:
		message = model.CreateAccount{
			AccountName:    receiver,
			Currency:       parts[4],
			IsBalanceCheck: parts[5] != "f",
		}

	case model.PromiseOrder:
		if amount, ok := new(money.Dec).SetString(parts[5]); ok {
			message = model.Promise{
				Transaction: parts[4],
				Amount:      amount,
				Currency:    parts[6],
			}
		}

	case model.CommitOrder:
		if amount, ok := new(money.Dec).SetString(parts[5]); ok {
			message = model.Commit{
				Transaction: parts[4],
				Amount:      amount,
				Currency:    parts[6],
			}
		}

	case model.RollbackOrder:
		if amount, ok := new(money.Dec).SetString(parts[5]); ok {
			message = model.Rollback{
				Transaction: parts[4],
				Amount:      amount,
				Currency:    parts[6],
			}
		}

	}

	if message == nil {
		log.Warnf("Deserialization of unsuported message [%s %s/%s] : %v", ref.Name, region, sender, parts)
		s.SendRemote(region, model.FatalErrorMessage(receiver, sender))
		return
	}

	ref.Tell(message, system.Coordinates{
		Name:   sender,
		Region: region,
	})
}

// NilAccount represents account that is neither existing neither non existing
func NilAccount(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		log.Info("Nil account reacting")
		state := t_state.(model.Account)

		snapshotHydration := persistence.LoadAccount(s.storage, state.AccountName)

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
func NonExistAccount(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Account)

		switch msg := context.Data.(type) {

		case model.CreateAccount:
			currency := strings.ToUpper(msg.Currency)
			isBalanceCheck := msg.IsBalanceCheck

			snaphostResult := persistence.CreateAccount(s.storage, state.AccountName, currency, isBalanceCheck)

			if snaphostResult == nil {
				s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (NonExist CreateAccount) Error", state.AccountName)
				return
			}

			context.Receiver.Become(*snaphostResult, ExistAccount(s))
			s.SendRemote(context.Sender.Region, model.AccountCreatedMessage(context.Receiver.Name, context.Sender.Name))
			log.Infof("New Account %s Created", state.AccountName)
			log.Debugf("%s ~ (NonExist CreateAccount) OK", state.AccountName)
			s.metrics.AccountCreated()

		case model.Rollback:
			s.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist Rollback) OK", state.AccountName)

		default:
			s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist Unknown Message) Error", state.AccountName)
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
			s.SendRemote(context.Sender.Region, model.AccountStateMessage(context.Receiver.Name, context.Sender.Name, state.Currency, state.Balance.String(), state.Promised.String(), state.IsBalanceCheck))
			log.Debugf("%s ~ (Exist GetAccountState) OK", state.AccountName)

		case model.CreateAccount:
			s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist CreateAccount) Error", state.AccountName)

		case model.Promise:
			if state.PromiseBuffer.Contains(msg.Transaction) {
				s.SendRemote(context.Sender.Region, model.PromiseAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) OK Already Accepted", state.AccountName)
				return
			}

			if state.Currency != msg.Currency {
				s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Promise) Error Currency Mismatch", state.AccountName)
				return
			}

			nextPromised := new(money.Dec).Add(state.Promised, msg.Amount)

			if !state.IsBalanceCheck || new(money.Dec).Add(state.Balance, nextPromised).Sign() >= 0 {
				if !persistence.PersistPromise(s.storage, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
					s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
					log.Warnf("%s ~ (Exist Promise) Error Could not Persist", state.AccountName)
					return
				}

				next := state.Copy()
				next.Promised = nextPromised
				next.PromiseBuffer.Add(msg.Transaction)

				context.Receiver.Become(next, ExistAccount(s))

				s.SendRemote(context.Sender.Region, model.PromiseAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Infof("Account %s Promised %s %s", state.AccountName, msg.Amount.String(), state.Currency)
				log.Debugf("%s ~ (Exist Promise) OK", state.AccountName)
				s.metrics.PromiseAccepted()
				return
			}

			if new(money.Dec).Sub(state.Balance, msg.Amount).Sign() < 0 {
				s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Promise) Error Insufficient Funds", state.AccountName)
				return
			}

			// FIXME boucing not handled
			s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Promise) Error ... (Bounce?)", state.AccountName)

		case model.Commit:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				s.SendRemote(context.Sender.Region, model.CommitAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Commit) OK Already Accepted", state.AccountName)
				return
			}

			if !persistence.PersistCommit(s.storage, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
				s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Commit) Error Could not Persist", state.AccountName)
				return
			}

			next := state.Copy()
			next.Balance = new(money.Dec).Add(state.Balance, msg.Amount)
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(next, ExistAccount(s))

			s.SendRemote(context.Sender.Region, model.CommitAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist Commit) OK", state.AccountName)
			s.metrics.CommitAccepted()

		case model.Rollback:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				s.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Rollback) OK Already Accepted", state.AccountName)
				return
			}

			if !persistence.PersistRollback(s.storage, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
				s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Rollback) Error Could not Persist", state.AccountName)
				return
			}

			next := state.Copy()
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(next, ExistAccount(s))

			s.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Infof("Account %s Rejected %s %s", state.AccountName, msg.Amount.String(), state.Currency)
			log.Debugf("%s ~ (Exist Rollback) OK", state.AccountName)
			s.metrics.RollbackAccepted()

		case model.Update:
			if msg.Version != state.Version {
				log.Debugf("%s ~ (Exist Update) Error Already Updated", state.AccountName)
				return
			}

			result := persistence.LoadAccount(s.storage, state.AccountName)
			if result == nil {
				log.Warnf("%s ~ (Exist Update) Error no existing snapshot", state.AccountName)
				return
			}

			next := persistence.UpdateAccount(s.storage, state.AccountName, result)
			if next == nil {
				log.Warnf("%s ~ (Exist Update) Error unable to update", state.AccountName)
				return
			}

			context.Receiver.Become(*next, ExistAccount(s))
			log.Infof("Account %s Updated Snapshot to %d", state.AccountName, next.Version)
			log.Debugf("%s ~ (Exist Update) OK", state.AccountName)

		default:
			s.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Unknown Message) Error", state.AccountName)

		}

		return
	}
}

// SpawnAccountActor returns new account actor instance registered into actor
// system
func (s *ActorSystem) SpawnAccountActor(name string) string {
	// FIXME split to multiple functions

	envelope := system.NewEnvelope(name, model.NewAccount(name))

	err := s.RegisterActor(envelope, NilAccount(s))
	if err != nil {
		log.Warnf("%s ~ Spawning Actor Error unable to register", name)
		return ""
	}

	log.Debugf("%s ~ Actor Spawned", name)
	return envelope.Name
}

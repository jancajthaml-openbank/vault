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
	"fmt"
	"strings"
	"sync"

	"github.com/jancajthaml-openbank/vault/actor"
	"github.com/jancajthaml-openbank/vault/config"
	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/persistence"

	lake "github.com/jancajthaml-openbank/lake-client/go"
	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

type actorsMap struct {
	All sync.Map
}

// ActorSystem represents actor system subroutine
type ActorSystem struct {
	Support
	tenant  string
	storage string
	metrics *Metrics
	Actors  *actorsMap
	Client  *lake.Client
	Name    string
}

// NewActorSystem returns actor system fascade
func NewActorSystem(ctx context.Context, cfg config.Configuration, metrics *Metrics) ActorSystem {
	lakeClient, _ := lake.NewClient(ctx, "Vault/"+cfg.Tenant, cfg.LakeHostname)

	return ActorSystem{
		Support: NewDaemonSupport(ctx),
		storage: cfg.RootStorage,
		tenant:  cfg.Tenant,
		metrics: metrics,
		Actors:  new(actorsMap),
		Client:  lakeClient,
		Name:    "Vault/" + cfg.Tenant,
	}
}

// ProcessLocalMessage send local message to actor by name
func (system ActorSystem) ProcessLocalMessage(msg interface{}, receiver string, sender actor.Coordinates) {
	ref, err := system.ActorOf(receiver)
	if err != nil {
		ref, err = system.ActorOf(system.SpawnAccountActor(receiver))
	}

	if err != nil {
		log.Warnf("Actor not found [%s local]", receiver)
		return
	}

	ref.Tell(msg, sender)
}

func (system ActorSystem) processRemoteMessage(parts []string) {
	region, receiver, sender, payload := parts[0], parts[1], parts[2], parts[3]

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("procesRemoteMessage recovered in [%s %s/%s] : %v", r, receiver, region, sender)
			system.SendRemote(region, model.FatalErrorMessage(receiver, sender))
		}
	}()

	ref, err := system.ActorOf(receiver)
	if err != nil {
		ref, err = system.ActorOf(system.SpawnAccountActor(receiver))
	}

	if err != nil {
		log.Warnf("Actor not found [%s %s/%s]", receiver, region, sender)
		system.SendRemote(region, model.FatalErrorMessage(receiver, sender))
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
		system.SendRemote(region, model.FatalErrorMessage(receiver, sender))
		return
	}

	ref.Tell(message, actor.Coordinates{
		Name:   sender,
		Region: region,
	})
	return
}

// NilAccount represents account that is neither existing neither non existing
func NilAccount(system ActorSystem) func(model.Account, actor.Context) {
	return func(state model.Account, context actor.Context) {
		snapshotHydration := persistence.LoadAccount(system.storage, state.AccountName)

		if snapshotHydration == nil {
			context.Receiver.Become(state, NonExistAccount(system))
			log.Debugf("%s ~ Nil -> NonExist", state.AccountName)
		} else {
			context.Receiver.Become(*snapshotHydration, ExistAccount(system))
			log.Debugf("%s ~ Nil -> Exist", state.AccountName)
		}

		context.Receiver.Tell(context.Data, context.Sender)
	}
}

// NonExistAccount represents account that does not exist
func NonExistAccount(system ActorSystem) func(model.Account, actor.Context) {
	return func(state model.Account, context actor.Context) {
		switch msg := context.Data.(type) {

		case model.CreateAccount:
			currency := strings.ToUpper(msg.Currency)
			isBalanceCheck := msg.IsBalanceCheck

			snaphostResult := persistence.CreateAccount(system.storage, state.AccountName, currency, isBalanceCheck)

			if snaphostResult == nil {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (NonExist CreateAccount) Error", state.AccountName)
				return
			}

			context.Receiver.Become(*snaphostResult, ExistAccount(system))
			system.SendRemote(context.Sender.Region, model.AccountCreatedMessage(context.Receiver.Name, context.Sender.Name))
			log.Infof("New Account %s Created", state.AccountName)
			log.Debugf("%s ~ (NonExist CreateAccount) OK", state.AccountName)
			system.metrics.AccountCreated()

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

// ExistAccount represents account that does exist
func ExistAccount(system ActorSystem) func(model.Account, actor.Context) {
	return func(state model.Account, context actor.Context) {
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
				if !persistence.PersistPromise(system.storage, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
					system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
					log.Warnf("%s ~ (Exist Promise) Error Could not Persist", state.AccountName)
					return
				}

				next := state.Copy()
				next.Promised = nextPromised
				next.PromiseBuffer.Add(msg.Transaction)

				context.Receiver.Become(*next, ExistAccount(system))

				system.SendRemote(context.Sender.Region, model.PromiseAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Infof("Account %s Promised %s %s", state.AccountName, msg.Amount.String(), state.Currency)
				log.Debugf("%s ~ (Exist Promise) OK", state.AccountName)
				system.metrics.PromiseAccepted()
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

			if !persistence.PersistCommit(system.storage, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Commit) Error Could not Persist", state.AccountName)
				return
			}

			next := state.Copy()
			next.Balance = new(money.Dec).Add(state.Balance, msg.Amount)
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(*next, ExistAccount(system))

			system.SendRemote(context.Sender.Region, model.CommitAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist Commit) OK", state.AccountName)
			system.metrics.CommitAccepted()

		case model.Rollback:
			if !state.PromiseBuffer.Contains(msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist Rollback) OK Already Accepted", state.AccountName)
				return
			}

			if !persistence.PersistRollback(system.storage, state.AccountName, state.Version, msg.Amount, msg.Transaction) {
				system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Warnf("%s ~ (Exist Rollback) Error Could not Persist", state.AccountName)
				return
			}

			next := state.Copy()
			next.Promised = new(money.Dec).Sub(state.Promised, msg.Amount)
			next.PromiseBuffer.Remove(msg.Transaction)

			context.Receiver.Become(*next, ExistAccount(system))

			system.SendRemote(context.Sender.Region, model.RollbackAcceptedMessage(context.Receiver.Name, context.Sender.Name))
			log.Infof("Account %s Rejected %s %s", state.AccountName, msg.Amount.String(), state.Currency)
			log.Debugf("%s ~ (Exist Rollback) OK", state.AccountName)
			system.metrics.RollbackAccepted()

		case model.Update:
			if msg.Version != state.Version {
				log.Debugf("%s ~ (Exist Update) Error Already Updated", state.AccountName)
				return
			}

			result := persistence.LoadAccount(system.storage, state.AccountName)
			if result == nil {
				log.Warnf("%s ~ (Exist Update) Error no existing snapshot", state.AccountName)
				return
			}

			next := persistence.UpdateAccount(system.storage, state.AccountName, result)
			if next == nil {
				log.Warnf("%s ~ (Exist Update) Error unable to update", state.AccountName)
				return
			}

			context.Receiver.Become(*next, ExistAccount(system))
			log.Infof("Account %s Updated Snapshot to %d", state.AccountName, next.Version)
			log.Debugf("%s ~ (Exist Update) OK", state.AccountName)

		default:
			system.SendRemote(context.Sender.Region, model.FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Unknown Message) Error", state.AccountName)

		}

		return
	}
}

// RegisterActor register new actor into actor system
func (system ActorSystem) RegisterActor(ref *actor.Envelope, initialState func(model.Account, actor.Context)) (err error) {
	_, exists := system.Actors.All.Load(ref.Name)
	if exists {
		return
	}

	ref.React(initialState)
	system.Actors.All.Store(ref.Name, ref)

	go func() {
		defer func() {
			if e := recover(); e != nil {
				switch x := e.(type) {
				case string:
					err = fmt.Errorf(x)
				case error:
					err = x
				default:
					err = fmt.Errorf("Unknown panic")
				}
			}
		}()

		for {
			select {
			case <-system.Done():
				return
			case p := <-ref.Backlog:
				ref.Receive(p)
			}
		}
	}()

	return
}

// SendRemote send message to remote region
func (system ActorSystem) SendRemote(destinationSystem, data string) {
	system.Client.Publish <- []string{destinationSystem, data}
}

// SpawnAccountActor returns new account actor instance registered into actor
// system
func (system ActorSystem) SpawnAccountActor(path string) string {
	// FIXME split to multiple functions

	envelope := actor.NewEnvelope(path)
	err := system.RegisterActor(envelope, NilAccount(system))
	if err != nil {
		log.Warnf("%s ~ Spawning Actor Error unable to register", path)
		return ""
	}

	log.Debugf("%s ~ Actor Spawned", path)
	return envelope.Name
}

// ActorOf return actor reference by name
func (system ActorSystem) ActorOf(name string) (*actor.Envelope, error) {
	ref, exists := system.Actors.All.Load(name)
	if !exists {
		return nil, fmt.Errorf("actor %v not registered", name)
	}

	return ref.(*actor.Envelope), nil
}

// Start handles everything needed to start metrics daemon
func (system ActorSystem) Start() {
	defer system.MarkDone()

	log.Info("Start actor system daemon")

	system.Client.Start()

	system.MarkReady()

	for {
		select {
		case message := <-system.Client.Receive:
			if len(message) < 4 {
				log.Warn("invalid message received")
				continue
			}
			system.processRemoteMessage(message)
		case <-system.Done():
			log.Info("Stopping actor system daemon")
			system.Client.Stop()
			log.Info("Stop actor system daemon")
			return
		}
	}
}

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
	"github.com/jancajthaml-openbank/vault/cron"
	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/utils"

	queue "github.com/jancajthaml-openbank/lake/utils"

	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

// Start starts `Vault/:tenant_id` actor system
func (system *ActorSystem) Start(params utils.RunParams, metrics *cron.Metrics) {
	if len(system.Name) != 0 {
		log.Warn("ActorSystem Already started")
		return
	}

	name := "Vault/" + params.Tenant

	log.Infof("ActorSystem Starting - %v", name)

	system.Name = name
	system.Client = queue.NewZMQClient(name, params.LakeHostname)

	go system.sourceRemoteMessages(params, metrics)

	log.Infof("ActorSystem Started - %v", name)

	return
}

func (system *ActorSystem) sourceRemoteMessages(params utils.RunParams, metrics *cron.Metrics) {
	for {
		if system == nil {
			return
		}
		system.processRemoteMessage(params, metrics)
	}
}

func (system *ActorSystem) ProcessLocalMessage(params utils.RunParams, metrics *cron.Metrics, msg interface{}, receiver, sender string) {
	if system == nil {
		return
	}

	ref, err := system.ActorOf(receiver)
	if err != nil {
		ref, err = system.ActorOf(system.SpawnAccountActor(params, metrics, receiver))
	}

	ref.Tell(msg, sender)
}

func (system *ActorSystem) processRemoteMessage(params utils.RunParams, metrics *cron.Metrics) {
	if system == nil {
		return
	}

	parts := system.Client.Receive()

	if len(parts) < 4 {
		log.Warn("invalid message received")
		return
	}

	region, receiver, sender, message := parts[0], parts[1], parts[2], parts[3]

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("procesRemoteMessage recovered in %v", r)
			system.SendRemote(region, model.FatalErrorMessage(receiver, sender))
		}
	}()

	ref, err := system.ActorOf(receiver)
	if err != nil {
		ref, err = system.ActorOf(system.SpawnAccountActor(params, metrics, receiver))
	}

	if err != nil {
		system.SendRemote(region, model.FatalErrorMessage(receiver, sender))
		return
	}

	switch message {

	case model.ReqAccountState:
		ref.Tell(model.GetAccountState{}, sender)

	case model.ReqCreateAccount:
		ref.Tell(model.CreateAccount{
			AccountName:    receiver,
			Currency:       parts[4],
			IsBalanceCheck: parts[5] != "f",
		}, sender)

	case model.PromiseOrder:
		amount, _ := new(money.Dec).SetString(parts[5])

		ref.Tell(model.Promise{
			Transaction: parts[4],
			Amount:      amount,
			Currency:    parts[6],
		}, sender)

	case model.CommitOrder:
		amount, _ := new(money.Dec).SetString(parts[5])

		ref.Tell(model.Commit{
			Transaction: parts[4],
			Amount:      amount,
			Currency:    parts[6],
		}, sender)

	case model.RollbackOrder:
		amount, _ := new(money.Dec).SetString(parts[5])

		ref.Tell(model.Rollback{
			Transaction: parts[4],
			Amount:      amount,
			Currency:    parts[6],
		}, sender)

	default:
		log.Warnf("Deserialization of unsuported message : %v", parts)
		system.SendRemote(region, model.FatalErrorMessage(receiver, sender))
	}

	return
}

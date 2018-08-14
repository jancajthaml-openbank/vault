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
	"github.com/jancajthaml-openbank/vault/pkg/metrics"
	"github.com/jancajthaml-openbank/vault/pkg/model"
	"github.com/jancajthaml-openbank/vault/pkg/utils"

	lake "github.com/jancajthaml-openbank/lake-client/go"

	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

// Start starts `Vault/:tenant_id` actor system
func (system *ActorSystem) Start(params utils.RunParams, m *metrics.Metrics) {
	if len(system.Name) != 0 {
		log.Warn("ActorSystem Already started")
		return
	}

	name := "Vault/" + params.Setup.Tenant

	log.Infof("ActorSystem Starting - %v", name)

	system.Name = name
	system.Client = lake.NewClient(name, params.Setup.LakeHostname)

	go system.sourceRemoteMessages(params, m)

	log.Infof("ActorSystem Started - %v", name)

	return
}

func (system *ActorSystem) sourceRemoteMessages(params utils.RunParams, m *metrics.Metrics) {
	for {
		if system == nil {
			return
		}
		system.processRemoteMessage(params, m)
	}
}

func (system *ActorSystem) ProcessLocalMessage(params utils.RunParams, m *metrics.Metrics, msg interface{}, receiver string, sender Coordinates) {
	if system == nil {
		return
	}

	ref, err := system.ActorOf(receiver)
	if err != nil {
		ref, err = system.ActorOf(system.SpawnAccountActor(params, m, receiver))
	}

	if err != nil {
		log.Warnf("Actor not found [%s local]", receiver)
		return
	}

	ref.Tell(msg, sender)
}

func (system *ActorSystem) processRemoteMessage(params utils.RunParams, m *metrics.Metrics) {
	if system == nil {
		return
	}

	parts := system.Client.Receive()

	if len(parts) < 4 {
		log.Warn("invalid message received")
		return
	}

	region, receiver, sender, payload := parts[0], parts[1], parts[2], parts[3]

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("procesRemoteMessage recovered in [%s %s/%s] : %v", r, receiver, region, sender)
			system.SendRemote(region, model.FatalErrorMessage(receiver, sender))
		}
	}()

	ref, err := system.ActorOf(receiver)
	if err != nil {
		ref, err = system.ActorOf(system.SpawnAccountActor(params, m, receiver))
	}

	if err != nil {
		log.Warnf("Actor not found [%s %s/%s]", receiver, region, sender)
		system.SendRemote(region, model.FatalErrorMessage(receiver, sender))
		return
	}

	var message interface{} = nil

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

	ref.Tell(message, Coordinates{sender, region})
	return
}

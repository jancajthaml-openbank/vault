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
	"fmt"

	"github.com/jancajthaml-openbank/vault/model"
	"github.com/jancajthaml-openbank/vault/utils"

	queue "github.com/jancajthaml-openbank/lake/utils"

	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

func StartSupervisor(params utils.RunParams) *ActorSystem {
	name := "Vault/" + params.Tenant

	log.Infof("ActorSystem Starting - %v", name)

	system := new(ActorSystem)
	system.Name = name
	system.Client = queue.NewZMQClient(name, "queue") // FIXME hostname to params

	go system.sourceRemoteMessages(params)

	log.Infof("ActorSystem Started - %v", name)

	return system
}

func (system *ActorSystem) sourceRemoteMessages(params utils.RunParams) {
	for {
		if system == nil {
			return
		}
		system.ProcessRemoteMessage(params)
	}
}

func (system *ActorSystem) ProcessLocalMessage(params utils.RunParams, msg interface{}, reciever string, sender string) {
	if system == nil {
		return
	}

	ref, err := system.ActorOf(reciever)
	if err != nil {
		ref, err = system.ActorOf(system.SpawnAccountActor(params, reciever))
	}

	ref.Tell(msg, sender)
}

func (system *ActorSystem) ProcessRemoteMessage(params utils.RunParams) {
	if system == nil {
		return
	}

	parts := system.Client.Receive()

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("procesRemoteMessage recovered in %v", r)
			system.SendRemote(parts[0], fmt.Sprintf("error %s %s", parts[3], parts[2]))
		}
	}()

	ref, err := system.ActorOf(parts[3])
	if err != nil {
		ref, err = system.ActorOf(system.SpawnAccountActor(params, parts[3]))
	}

	if err != nil {
		system.SendRemote(parts[1], fmt.Sprintf("error %s %s", parts[3], parts[2]))
		return
	}

	switch parts[1] {

	case model.ReqAccountBalance:
		ref.Tell(model.GetAccountBalance{}, parts[2])

	case model.ReqCreateAccount:
		ref.Tell(model.CreateAccount{
			AccountName:    parts[3],
			Currency:       parts[4],
			IsBalanceCheck: parts[5] != "f",
		}, parts[2])

	case model.PromiseOrder:
		amount, _ := new(money.Dec).SetString(parts[5])

		ref.Tell(model.Promise{
			Transaction: parts[4],
			Amount:      amount,
			Currency:    parts[6],
		}, parts[2])

	case model.CommitOrder:
		amount, _ := new(money.Dec).SetString(parts[5])

		ref.Tell(model.Commit{
			Transaction: parts[4],
			Amount:      amount,
			Currency:    parts[6],
		}, parts[2])

	case model.RollbackOrder:
		amount, _ := new(money.Dec).SetString(parts[5])

		ref.Tell(model.Rollback{
			Transaction: parts[4],
			Amount:      amount,
			Currency:    parts[6],
		}, parts[2])

	default:
		log.Warnf("Deserialization of unsuported message : %v", parts)
		system.SendRemote(parts[1], fmt.Sprintf("error %s %s", parts[3], parts[2]))
	}

	return
}

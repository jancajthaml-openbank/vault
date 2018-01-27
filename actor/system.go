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
	"sync"

	queue "github.com/jancajthaml-openbank/lake/utils"
	"github.com/jancajthaml-openbank/vault/model"
	log "github.com/sirupsen/logrus"
)

type ActorSystem struct {
	Actors sync.Map
	Client *queue.ZMQClient
	Name   string
}

func (system *ActorSystem) Stop() {
	if system == nil {
		log.Warn("Actor system not started")
		//return fmt.Errorf("actor system not started")
		return
	}
	// FIXME delete these
	//system.Actors
	log.Debugf("Actor system closed")
	return
}

func (system *ActorSystem) RegisterActor(ref *actor, initialState func(model.Snapshot, model.Account, Context)) error {
	if system == nil {
		log.Warn("Actor system not started")
		return fmt.Errorf("actor System not started")
	}

	_, exists := system.Actors.Load(ref.Name)
	if exists {
		return nil
	}

	ref.react(initialState)
	system.Actors.Store(ref.Name, ref)

	go receive(ref)

	return nil
}

func (system *ActorSystem) SendRemote(destinationSystem, data string) {
	if system == nil {
		log.Warn("Actor system not started")
		return
	}

	system.Client.Publish(destinationSystem, data)
}

func (system *ActorSystem) ActorOf(name string) (*actor, error) {
	if system == nil {
		log.Warn("Actor system not started")
		return nil, fmt.Errorf("actor System not started")
	}

	ref, exists := system.Actors.Load(name)
	if !exists {
		return nil, fmt.Errorf("actor %v not registered", name)
	}

	return ref.(*actor), nil
}

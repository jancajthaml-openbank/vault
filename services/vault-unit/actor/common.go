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

	"github.com/jancajthaml-openbank/vault-unit/daemon"
	"github.com/jancajthaml-openbank/vault-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

var nilCoordinates = system.Coordinates{}

// ProcessLocalMessage processing of local message to this vault
func ProcessLocalMessage(s *daemon.ActorSystem) system.ProcessLocalMessage {
	return func(message interface{}, to system.Coordinates, from system.Coordinates) {
		if to.Region != "" && to.Region != s.Name {
			log.Warnf("Invalid region received [local %s -> local %s]", from, to)
			return
		}
		// FIXME check to.Region and if now for this region, relay
		ref, err := s.ActorOf(to.Name)
		if err != nil {
			ref, err = spawnAccountActor(s, to.Name)
		}

		if err != nil {
			log.Warnf("Actor not found [local %s]", to)
			return
		}
		ref.Tell(message, from)
	}
}

func asEnvelopes(s *daemon.ActorSystem, parts []string) (system.Coordinates, system.Coordinates, string, error) {
	if len(parts) < 4 {
		return nilCoordinates, nilCoordinates, "", fmt.Errorf("invalid message received %+v", parts)
	}

	region, receiver, sender, payload := parts[0], parts[1], parts[2], parts[3]

	from := system.Coordinates{
		Name:   sender,
		Region: region,
	}

	to := system.Coordinates{
		Name:   receiver,
		Region: s.Name,
	}

	return from, to, payload, nil
}

func spawnAccountActor(s *daemon.ActorSystem, name string) (*system.Envelope, error) {
	envelope := system.NewEnvelope(name, model.NewAccount(name))

	err := s.RegisterActor(envelope, NilAccount(s))
	if err != nil {
		log.Warnf("%s ~ Spawning Actor Error unable to register", name)
		return nil, err
	}

	log.Debugf("%s ~ Actor Spawned", name)
	return envelope, nil
}

// ProcessRemoteMessage processing of remote message to this vault
func ProcessRemoteMessage(s *daemon.ActorSystem) system.ProcessRemoteMessage {
	return func(parts []string) {
		from, to, payload, err := asEnvelopes(s, parts)
		if err != nil {
			log.Warn(err.Error())
			return
		}

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("procesRemoteMessage recovered in [remote %v -> local %v] : %+v", from, to, r)
				s.SendRemote(from.Region, FatalErrorMessage(to.Name, from.Name))
			}
		}()

		ref, err := s.ActorOf(to.Name)
		if err != nil {
			ref, err = spawnAccountActor(s, to.Name)
		}

		if err != nil {
			log.Warnf("Actor not found [remote %v -> local %v]", from, to)
			s.SendRemote(from.Region, FatalErrorMessage(to.Name, from.Name))
			return
		}

		var message interface{}

		switch payload {

		case ReqAccountState:
			message = model.GetAccountState{}

		case ReqCreateAccount:
			if len(parts) == 6 {
				message = model.CreateAccount{
					AccountName:    to.Name,
					Currency:       parts[4],
					IsBalanceCheck: parts[5] != "f",
				}
			}

		case PromiseOrder:
			if len(parts) == 7 {
				if amount, ok := new(money.Dec).SetString(parts[5]); ok {
					message = model.Promise{
						Transaction: parts[4],
						Amount:      amount,
						Currency:    parts[6],
					}
				}
			}

		case CommitOrder:
			if len(parts) == 7 {
				if amount, ok := new(money.Dec).SetString(parts[5]); ok {
					message = model.Commit{
						Transaction: parts[4],
						Amount:      amount,
						Currency:    parts[6],
					}
				}
			}

		case RollbackOrder:
			if len(parts) == 7 {
				if amount, ok := new(money.Dec).SetString(parts[5]); ok {
					message = model.Rollback{
						Transaction: parts[4],
						Amount:      amount,
						Currency:    parts[6],
					}
				}
			}

		}

		if message == nil {
			log.Warnf("Deserialization of unsuported message [remote %v -> local %v] : %+v", from, to, parts)
			s.SendRemote(from.Region, FatalErrorMessage(to.Name, from.Name))
			return
		}

		ref.Tell(message, from)
	}
}

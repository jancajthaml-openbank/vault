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
	"fmt"
	"strings"

	"github.com/jancajthaml-openbank/vault-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

// ProcessLocalMessage processing of local message to this vault
func ProcessLocalMessage(s *ActorSystem) system.ProcessLocalMessage {
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
		ref.Tell(message, to, from)
	}
}

func asEnvelopes(s *ActorSystem, msg string) (system.Coordinates, system.Coordinates, []string, error) {
	parts := strings.Split(msg, " ")

	if len(parts) < 5 {
		return system.Coordinates{}, system.Coordinates{}, nil, fmt.Errorf("invalid message received %+v", parts)
	}

	recieverRegion, senderRegion, receiverName, senderName := parts[0], parts[1], parts[2], parts[3]

	from := system.Coordinates{
		Name:   senderName,
		Region: senderRegion,
	}

	to := system.Coordinates{
		Name:   receiverName,
		Region: recieverRegion,
	}

	return from, to, parts, nil
}

func spawnAccountActor(s *ActorSystem, name string) (*system.Envelope, error) {
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
func ProcessRemoteMessage(s *ActorSystem) system.ProcessRemoteMessage {
	return func(msg string) {
		from, to, parts, err := asEnvelopes(s, msg)
		if err != nil {
			log.Warn(err.Error())
			return
		}

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("procesRemoteMessage recovered in [remote %v -> local %v] : %+v", from, to, r)
				s.SendRemote(FatalErrorMessage(system.Context{
					Receiver: to,
					Sender:   from,
				}))
			}
		}()

		ref, err := s.ActorOf(to.Name)
		if err != nil {
			ref, err = spawnAccountActor(s, to.Name)
		}

		if err != nil {
			log.Warnf("Actor not found [remote %v -> local %v]", from, to)
			s.SendRemote(FatalErrorMessage(system.Context{
				Receiver: to,
				Sender:   from,
			}))
			return
		}

		var message interface{}

		switch parts[4] {

		case ReqAccountState:
			message = model.GetAccountState{}

		case ReqCreateAccount:
			if len(parts) == 7 {
				message = model.CreateAccount{
					AccountName:    to.Name,
					Currency:       parts[5],
					IsBalanceCheck: parts[6] != "f",
				}
			}

		case PromiseOrder:
			if len(parts) == 8 {
				if amount, ok := new(money.Dec).SetString(parts[6]); ok {
					message = model.Promise{
						Transaction: parts[5],
						Amount:      amount,
						Currency:    parts[7],
					}
				}
			}

		case CommitOrder:
			if len(parts) == 8 {
				if amount, ok := new(money.Dec).SetString(parts[6]); ok {
					message = model.Commit{
						Transaction: parts[5],
						Amount:      amount,
						Currency:    parts[7],
					}
				}
			}

		case RollbackOrder:
			if len(parts) == 8 {
				if amount, ok := new(money.Dec).SetString(parts[6]); ok {
					message = model.Rollback{
						Transaction: parts[5],
						Amount:      amount,
						Currency:    parts[7],
					}
				}
			}

		}

		if message == nil {
			log.Warnf("Deserialization of unsuported message [remote %v -> local %v] : %+v", from, to, parts)
			s.SendRemote(FatalErrorMessage(system.Context{
				Receiver: to,
				Sender:   from,
			}))
			return
		}

		ref.Tell(message, to, from)
	}
}

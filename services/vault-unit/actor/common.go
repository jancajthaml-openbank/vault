// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"strconv"
	"strings"

	"github.com/jancajthaml-openbank/vault-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

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

// ProcessMessage processing of remote message
func ProcessMessage(s *ActorSystem) system.ProcessMessage {
	return func(msg string, to system.Coordinates, from system.Coordinates) {

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("procesRemoteMessage recovered in [remote %v -> local %v] : %+v", from, to, r)
				s.SendMessage(FatalError, from, to)
			}
		}()

		ref, err := s.ActorOf(to.Name)
		if err != nil {
			ref, err = spawnAccountActor(s, to.Name)
		}

		if err != nil {
			log.Warnf("Actor not found [remote %v -> local %v]", from, to)
			s.SendMessage(FatalError, from, to)
			return
		}

		parts := strings.Split(msg, " ")

		var message interface{}

		switch parts[0] {

		case ReqAccountState:
			message = model.GetAccountState{}

		case UpdateSnapshot:
			if len(parts) == 2 {
				if version, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					message = model.Update{
						Version: version,
					}
				}
			}

		case ReqCreateAccount:
			if len(parts) == 4 {
				message = model.CreateAccount{
					Name:           to.Name,
					Format:         parts[1],
					Currency:       parts[2],
					IsBalanceCheck: parts[3] != "f",
				}
			}

		case PromiseOrder:
			if len(parts) == 4 {
				if amount, ok := new(money.Dec).SetString(parts[2]); ok {
					message = model.Promise{
						Transaction: parts[1],
						Amount:      amount,
						Currency:    parts[3],
					}
				}
			}

		case CommitOrder:
			if len(parts) == 4 {
				if amount, ok := new(money.Dec).SetString(parts[2]); ok {
					message = model.Commit{
						Transaction: parts[1],
						Amount:      amount,
						Currency:    parts[3],
					}
				}
			}

		case RollbackOrder:
			if len(parts) == 4 {
				if amount, ok := new(money.Dec).SetString(parts[2]); ok {
					message = model.Rollback{
						Transaction: parts[1],
						Amount:      amount,
						Currency:    parts[3],
					}
				}
			}
		}

		if message == nil {
			log.Warnf("Deserialization of unsuported message [remote %v -> local %v] : %+v", from, to, parts)
			s.SendMessage(FatalError, from, to)
			return
		}

		log.Debugf("Sending mesage to actor %+v message %+v", ref, message)
		ref.Tell(message, to, from)
	}
}

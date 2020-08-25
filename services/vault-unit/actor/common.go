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
	"fmt"

	"github.com/jancajthaml-openbank/vault-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
	money "gopkg.in/inf.v0"
)

func parseMessage(msg string) (interface{}, error) {
	start := 0
	end := len(msg)
	parts := make([]string, 4)
	idx := 0
	i := 0
	for i < end && idx < 4 {
		if msg[i] == 32 {
			if !(start == i && msg[start] == 32) {
				parts[idx] = msg[start:i]
				idx++
			}
			start = i + 1
		}
		i++
	}
	if idx < 4 && msg[start] != 32 && len(msg[start:]) > 0 {
		parts[idx] = msg[start:]
		idx++
	}

	if i != end - 1 {
		return nil, fmt.Errorf("message too large")
	}

	switch parts[0] {

	case ReqAccountState:
		return GetAccountState{}, nil

	case ReqCreateAccount:
		if idx == 4 {
			return CreateAccount{
				Format:         parts[1],
				Currency:       parts[2],
				IsBalanceCheck: parts[3] != "f",
			}, nil
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	case PromiseOrder:
		if idx == 4 {
			if amount, ok := new(money.Dec).SetString(parts[2]); ok {
				return Promise{
					Transaction: parts[1],
					Amount:      amount,
					Currency:    parts[3],
				}, nil
			}
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	case CommitOrder:
		if idx == 4 {
			if amount, ok := new(money.Dec).SetString(parts[2]); ok {
				return Commit{
					Transaction: parts[1],
					Amount:      amount,
					Currency:    parts[3],
				}, nil
			}
			return nil, fmt.Errorf("invalid order amount %s", msg)
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	case RollbackOrder:
		if idx == 4 {
			if amount, ok := new(money.Dec).SetString(parts[2]); ok {
				return Rollback{
					Transaction: parts[1],
					Amount:      amount,
					Currency:    parts[3],
				}, nil
			}
			return nil, fmt.Errorf("invalid order amount %s", msg)
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	default:
		return nil, fmt.Errorf("unknown message %s", msg)
	}
}

// ProcessMessage processing of remote message
func ProcessMessage(s *ActorSystem) system.ProcessMessage {
	return func(msg string, to system.Coordinates, from system.Coordinates) {
		message, err := parseMessage(msg)
		if err != nil {
			log.Warnf("%s [remote %v -> local %v]", err, from, to)
			s.SendMessage(FatalError, from, to)
			return
		}
		ref, err := s.ActorOf(to.Name)
		if err != nil {
			ref, err = NewAccountActor(s, to.Name)
		}
		if err != nil {
			log.Warnf("Actor not found [remote %v -> local %v]", from, to)
			s.SendMessage(FatalError, from, to)
			return
		}
		ref.Tell(message, to, from)
	}
}

// NewAccountActor returns new account actor Envelope
func NewAccountActor(s *ActorSystem, name string) (*system.Envelope, error) {
	envelope := system.NewEnvelope(name, model.NewAccount(name))
	err := s.RegisterActor(envelope, NilAccount(s))
	if err != nil {
		log.Warnf("%s ~ Spawning Actor Error unable to register", name)
		return nil, err
	}
	log.Debugf("%s ~ Actor Spawned", name)
	return envelope, nil
}

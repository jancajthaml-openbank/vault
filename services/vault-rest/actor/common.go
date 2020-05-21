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
	//"fmt"
	"strings"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
)

// ProcessMessage processing of remote message
func ProcessMessage(s *ActorSystem) system.ProcessMessage {
	return func(msg string, to system.Coordinates, from system.Coordinates) {
		parts := strings.Split(msg, " ")

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("procesRemoteMessage recovered in [remote %v -> local %v] : %+v", from, to, r)
			}
		}()

		ref, err := s.ActorOf(to.Name)
		if err != nil {
			// FIXME forward into deadletter receiver and finish whatever has started
			log.Warnf("Deadletter received [remote %v -> local %v] : %+v", from, to, msg)
			return
		}

		var message interface{}

		switch parts[0] {

		case FatalError:
			message = FatalError

		case RespAccountState:
			message = &Account{
				Format:         parts[1],
				Currency:       parts[2],
				IsBalanceCheck: parts[3] != "f",
				Balance:        parts[4],
				Blocking:       parts[5],
			}

		case RespAccountMissing:
			message = new(AccountMissing)

		case RespCreateAccount:
			message = new(AccountCreated)

		default:
			log.Warnf("Deserialization of unsuported message [remote %v -> local %v] : %+v", from, to, msg)
			message = FatalError
		}

		ref.Tell(message, to, from)
		return
	}
}

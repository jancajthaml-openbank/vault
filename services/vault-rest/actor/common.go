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

	"github.com/jancajthaml-openbank/vault-rest/model"

	system "github.com/jancajthaml-openbank/actor-system"
)

func parseMessage(msg string) (interface{}, error) {
	start := 0
	end := len(msg)
	parts := make([]string, 6)
	idx := 0
	i := 0
	for i < end && idx < 6 {
		if msg[i] == 32 {
			if !(start == i && msg[start] == 32) {
				parts[idx] = msg[start:i]
				idx++
			}
			start = i + 1
		}
		i++
	}
	if idx < 6 && msg[start] != 32 && len(msg[start:]) > 0 {
		parts[idx] = msg[start:]
		idx++
	}

	if i != end {
		return nil, fmt.Errorf("message too large")
	}

	switch parts[0] {

	case FatalError:
		return FatalError, nil

	case RespAccountState:
		if idx == 6 {
			return &model.Account{
				Format:         parts[1],
				Currency:       parts[2],
				IsBalanceCheck: parts[3] != "f",
				Balance:        parts[4],
				Blocking:       parts[5],
			}, nil
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	case RespAccountMissing:
		return new(AccountMissing), nil

	case RespCreateAccount:
		return new(AccountCreated), nil

	default:
		return nil, fmt.Errorf("unknown message %s", msg)
	}
}

// ProcessMessage processing of remote message
func ProcessMessage(s *System) system.ProcessMessage {
	return func(msg string, to system.Coordinates, from system.Coordinates) {
		ref, err := s.ActorOf(to.Name)
		if err != nil {
			// FIXME forward into deadletter receiver and finish whatever has started
			return
		}
		message, err := parseMessage(msg)
		if err != nil {
			log.Warn().Msgf("%s [remote %v -> local %v]", err, from, to)
		}
		ref.Tell(message, to, from)
		return
	}
}

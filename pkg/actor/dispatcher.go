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

import log "github.com/sirupsen/logrus"

// Context represents actor message envelope
type Context struct {
	Data     interface{}
	Receiver *Actor
	Sender   Coordinates
}

func dispatch(channel chan Context, data interface{}, receiver *Actor, sender Coordinates) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Dispatch recovered in %v", r)
		}
	}()

	channel <- Context{data, receiver, sender}
}

func receive(ref *Actor) {
	defer func() {
		if r := recover(); r != nil {
			// FIXME reply error to sender
			log.Errorf("Receive recovered in %v", r)
		}
	}()

	for {
		select {
		case p := <-ref.dataChan:
			ref.Receive(p)
		}
	}
}

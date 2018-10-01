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
)

// Envelope represents single actor
type Envelope struct {
	Name    string
	receive func(model.Account, Context)
	Backlog chan Context
	State   model.Account
}

// NewEnvelope returns new actor instance
func NewEnvelope(name string) *Envelope {
	return &Envelope{
		Name:    name,
		Backlog: make(chan Context, 64),
		State:   model.NewAccount(name),
	}
}

// Tell queues message to actor
func (ref *Envelope) Tell(data interface{}, sender Coordinates) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("Unknown panic")
			}
		}
	}()

	if ref == nil {
		err = fmt.Errorf("actor reference %v not found", ref)
		return
	}

	ref.Backlog <- Context{
		Data:     data,
		Receiver: ref,
		Sender:   sender,
	}
	return
}

// Become transforms actor behaviour for next message
func (ref *Envelope) Become(state model.Account, f func(model.Account, Context)) {
	if ref == nil {
		return
	}
	ref.State = state
	ref.React(f)
	return
}

func (ref *Envelope) String() string {
	if ref == nil {
		return "Deadletter"
	}
	return ref.Name
}

func (ref *Envelope) React(f func(model.Account, Context)) {
	if ref == nil {
		return
	}
	ref.receive = f
	return
}

// Receive dequeues message to actor
func (ref *Envelope) Receive(msg Context) {
	if ref.receive == nil {
		return
	}
	ref.receive(ref.State, msg)
}

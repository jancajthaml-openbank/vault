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

	log "github.com/sirupsen/logrus"
)

type actor struct {
	Name     string
	receive  func(model.Snapshot, model.Account, Context)
	dataChan chan Context
	Meta     model.Account
	State    model.Snapshot
}

type Coordinates struct {
	Name   string
	Region string
}

func NewAccountEnvelope(name string) *actor {
	ref := new(actor)
	ref.Name = name
	ref.dataChan = make(chan Context, 64) // FIXME make buffer from params
	ref.State = model.NewSnapshot()
	ref.Meta = model.NewAccount(name)
	return ref
}

func (ref *actor) Tell(data interface{}, sender Coordinates) error {
	if ref == nil {
		log.Warnf("actor reference %v not found", ref)
		return fmt.Errorf("actor reference %v not found", ref)
	}

	dispatch(ref.dataChan, data, ref, sender)
	return nil
}

func (ref *actor) Become(state model.Snapshot, meta model.Account, f func(model.Snapshot, model.Account, Context)) error {
	if ref == nil {
		log.Warnf("actor reference %v not found", ref)
		return fmt.Errorf("actor reference %v not found", ref)
	}
	ref.State = state
	ref.Meta = meta
	ref.react(f)
	return nil
}

func (ref *actor) String() string {
	return ref.Name
}

func (ref *actor) react(f func(model.Snapshot, model.Account, Context)) *actor {
	ref.receive = f
	return ref
}

func (ref *actor) Receive(message Context) {
	if ref.receive == nil {
		return
	}

	ref.receive(ref.State, ref.Meta, message)
}

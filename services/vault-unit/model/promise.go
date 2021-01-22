// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package model

// Promises is stable set datastructure for promised transaction Ids
type Promises struct {
	keys       []string
	values     map[string]bool
}

// NewPromises returns a fascade for promises
func NewPromises() Promises {
	return Promises{
		keys: make([]string, 0),
		values: make(map[string]bool),
	}
}

// Add adds items to set if not already present
func (s *Promises) Add(item string) {
	if s == nil {
		return
	}
	if !s.values[item] {
		s.keys = append(s.keys, item)
	}
	s.values[item] = true
}

// Contains returns true if item is present in set
func (s Promises) Contains(item string) bool {
	return s.values[item]
}

// Remove removes items from set
func (s *Promises) Remove(item string) {
	if s == nil {
		return
	}
	if !s.values[item] {
		return
	}
	for i, k := range s.keys {
		if k == item {
			s.keys = append(s.keys[:i], s.keys[i+1:]...)
			break
		}
	}
	delete(s.values, item)
}

// Size returns number of items in set
func (s Promises) Size() int {
	return len(s.keys)
}

// Iterator returns values iterator
func (s Promises) Iterator() <-chan string {
	chnl := make(chan string, len(s.keys))
	go func() {
		defer close(chnl)
		for idx := range s.keys {
			chnl <- s.keys[idx]
		}
    }()
    return chnl
}

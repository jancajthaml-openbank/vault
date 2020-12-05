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

package model

import "bytes"

// Promises is stable set datastructure for promised transaction Ids
type Promises struct {
	keys   []int
	values map[int]string
	index  map[string]int
	tail   int
}

// NewPromises returns a fascade for promises
func NewPromises() Promises {
	return Promises{
		index:  make(map[string]int),
		keys:   make([]int, 0),
		values: make(map[int]string),
		tail:   0,
	}
}

// Add adds items to set if not already present
func (s *Promises) Add(items ...string) {
	if s == nil {
		return
	}
	for idx := range items {
		if _, found := s.index[items[idx]]; found {
			continue
		}
		s.keys = append(s.keys, s.tail)
		s.values[s.tail] = items[idx]
		s.index[items[idx]] = s.tail
		s.tail++
	}
}

// Contains returns true if all items are present in set
func (s Promises) Contains(items ...string) bool {
	for idx := range items {
		if _, found := s.index[items[idx]]; !found {
			return false
		}
	}
	return true
}

// Remove removes items from set
func (s *Promises) Remove(items ...string) {
	if s == nil {
		return
	}
	for idx := range items {
		index, found := s.index[items[idx]]
		if !found {
			continue
		}

		delete(s.index, items[idx])
		delete(s.values, index)

		for i := range s.keys {
			if s.keys[i] == index {
				s.keys = append(s.keys[:i], s.keys[i+1:]...)
				break
			}
		}
	}
}

// Size returns number of items in set
func (s Promises) Size() int {
	return len(s.values)
}

// String serializes promises
func (s Promises) String() string {
	if len(s.keys) == 0 {
		return "[]"
	}
	var buffer bytes.Buffer

	buffer.WriteString("[")
	for idx := range s.keys {
		buffer.WriteString(s.values[s.keys[idx]])
		buffer.WriteString(",")
	}
	buffer.Truncate(buffer.Len() - 1)
	buffer.WriteString("]")

	return buffer.String()
}

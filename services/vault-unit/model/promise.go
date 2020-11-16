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
		tail: 0,
	}
}

// Add adds items to set if not already present
func (s *Promises) Add(items ...string) {
	if s == nil {
		return
	}
	for _, item := range items {
		if _, found := s.index[item]; found {
			continue
		}
		s.keys = append(s.keys, s.tail)
		s.values[s.tail] = item
		s.index[item] = s.tail
		s.tail++
	}
}

// Contains returns true if all items are present in set
func (s Promises) Contains(items ...string) bool {
	for _, item := range items {
		if _, found := s.index[item]; !found {
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
	for _, item := range items {
		idx, found := s.index[item]
		if !found {
			continue
		}

		delete(s.index, item)
		delete(s.values, idx)

		for i := range s.keys {
			if s.keys[i] == idx {
				s.keys = append(s.keys[:i], s.keys[i+1:]...)
				break
			}
		}
	}
}

// Values returns slice of items in order or insertion
func (s Promises) Values() []string {
	result := make([]string, len(s.values))
	for i, k := range s.keys {
		result[i] = s.values[k]
	}
	return result
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
	for _, k := range s.keys {
		buffer.WriteString(s.values[k])
		buffer.WriteString(",")
	}
	buffer.Truncate(buffer.Len()-1)
	buffer.WriteString("]")

	return buffer.String()
}

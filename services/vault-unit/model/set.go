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

// TransactionSet is set datastructure for transaction Ids
type TransactionSet struct {
	keys   []int
	values map[int]string
	index  map[string]int
	tail   int
}

// NewTransactionSet returns empty set
func NewTransactionSet() TransactionSet {
	return TransactionSet{
		index:  make(map[string]int),
		keys:   make([]int, 0),
		values: make(map[int]string),
	}
}

// Add adds items to set if not already present
func (s *TransactionSet) Add(items ...string) {
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
func (s *TransactionSet) Contains(items ...string) bool {
	for _, item := range items {
		if _, found := s.index[item]; !found {
			return false
		}
	}
	return true
}

// Remove removes items from set
func (s *TransactionSet) Remove(items ...string) {
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
func (s *TransactionSet) Values() []string {
	values := make([]string, len(s.values))
	for i, k := range s.keys {
		values[i] = s.values[k]
	}
	return values
}

// Size returns number of items in set
func (s *TransactionSet) Size() int {
	return len(s.values)
}

func (s *TransactionSet) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("[")
	for _, k := range s.keys {
		buffer.WriteString(s.values[k])
		buffer.WriteString(",")
	}
	buffer.WriteString("]")

	return buffer.String()
}

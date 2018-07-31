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

package model

import (
	"bytes"
)

// TransactionSet is set datastructure for transaction Ids
type TransactionSet struct {
	items map[string]interface{}
}

// NewTransactionSet returns empty set
func NewTransactionSet() TransactionSet {
	return TransactionSet{make(map[string]interface{})}
}

// Add adds element to set
func (set *TransactionSet) Add(i string) {
	set.items[i] = nil
}

// AddAll adds all elements to set
func (set *TransactionSet) AddAll(i []string) {
	for _, b := range i {
		set.items[b] = nil
	}
}

// Contains returns true if value is present in set
func (set *TransactionSet) Contains(i string) bool {
	_, found := set.items[i]
	return found
}

// Remove removes element from set
func (set *TransactionSet) Remove(i string) {
	delete(set.items, i)
}

// Size returns number of items in set
func (set *TransactionSet) Size() int {
	return len(set.items)
}

// WriteTo writes all imtes to buffer in format "\n" + item
func (set *TransactionSet) WriteTo(buffer *bytes.Buffer) {
	if set == nil {
		return
	}

	for v := range set.items {
		buffer.WriteString("\n")
		buffer.WriteString(v)
	}

	return
}

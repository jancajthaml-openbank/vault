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

// TransactionSet is set datastructure for transaction Ids
type TransactionSet struct {
	Items map[string]interface{}
}

// NewTransactionSet returns empty set
func NewTransactionSet() TransactionSet {
	return TransactionSet{make(map[string]interface{})}
}

// Add adds element to set
func (set *TransactionSet) Add(item string) {
	set.Items[item] = nil
}

// AddAll adds all elements to set
func (set *TransactionSet) AddAll(items []string) {
	for _, item := range items {
		set.Items[item] = nil
	}
}

// Contains returns true if value is present in set
func (set *TransactionSet) Contains(item string) bool {
	_, found := set.Items[item]
	return found
}

// Remove removes element from set
func (set *TransactionSet) Remove(item string) {
	delete(set.Items, item)
}

// Size returns number of items in set
func (set *TransactionSet) Size() int {
	return len(set.Items)
}

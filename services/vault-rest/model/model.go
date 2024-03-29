// Copyright (c) 2016-2023, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"encoding/json"
	"fmt"
	"strings"
)

// Account represents account
type Account struct {
	Name           string `json:"-"`
	Format         string `json:"format"`
	Currency       string `json:"currency"`
	Balance        string `json:"balance"`
	Blocking       string `json:"blocking"`
	IsBalanceCheck bool   `json:"isBalanceCheck"`
}

// UnmarshalJSON unmarshal json of Account entity
func (entity *Account) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}
	all := struct {
		Name           string `json:"name"`
		Format         string `json:"format"`
		Currency       string `json:"currency"`
		IsBalanceCheck *bool  `json:"isBalanceCheck"`
	}{}
	err := json.Unmarshal(data, &all)
	if err != nil {
		return err
	}
	if all.Name == "" {
		return fmt.Errorf("missing attribute \"name\"")
	}
	if all.Format == "" {
		return fmt.Errorf("missing attribute \"format\"")
	}
	if all.Currency == "" {
		return fmt.Errorf("missing attribute \"currency\"")
	}
	if len(all.Currency) != 3 ||
		!((all.Currency[0] >= 'A' && all.Currency[0] <= 'Z') &&
			(all.Currency[1] >= 'A' && all.Currency[1] <= 'Z') &&
			(all.Currency[2] >= 'A' && all.Currency[2] <= 'Z')) {
		return fmt.Errorf("invalid value of attribute \"currency\"")
	}
	if all.IsBalanceCheck == nil {
		entity.IsBalanceCheck = true
	} else {
		entity.IsBalanceCheck = *all.IsBalanceCheck
	}

	entity.Name = strings.Replace(all.Name, " ", "_", -1)
	entity.Format = strings.Replace(all.Format, " ", "_", -1)
	entity.Currency = all.Currency
	return nil
}

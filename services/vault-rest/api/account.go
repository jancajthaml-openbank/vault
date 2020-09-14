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

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jancajthaml-openbank/vault-rest/actor"
	"github.com/jancajthaml-openbank/vault-rest/model"
	"github.com/jancajthaml-openbank/vault-rest/persistence"

	localfs "github.com/jancajthaml-openbank/local-fs"
	"github.com/labstack/echo/v4"
)

// GetAccount returns account state
func GetAccount(system *actor.System) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			return fmt.Errorf("missing tenant")
		}
		id := c.Param("id")
		if id == "" {
			return fmt.Errorf("missing id")
		}

		switch result := actor.GetAccount(system, tenant, id).(type) {

		case *actor.AccountMissing:
			c.Response().WriteHeader(http.StatusNotFound)
			return nil

		case *model.Account:
			chunk, err := json.Marshal(result)
			if err != nil {
				return err
			}
			c.Response().WriteHeader(http.StatusOK)
			c.Response().Write(chunk)
			c.Response().Flush()
			return nil

		case *actor.ReplyTimeout:
			c.Response().WriteHeader(http.StatusGatewayTimeout)
			return nil

		default:
			return fmt.Errorf("internal error")

		}
	}
}

// CreateAccount creates new account for given tenant
func CreateAccount(system *actor.System) func(c echo.Context) error {
	return func(c echo.Context) error {
		tenant := c.Param("tenant")
		if tenant == "" {
			return fmt.Errorf("missing tenant")
		}

		req, err := ioutil.ReadAll(c.Request().Body)
		defer c.Request().Body.Close()
		if err != nil {
			c.Response().WriteHeader(http.StatusBadRequest)
			return err
		}

		var account = new(model.Account)
		if json.Unmarshal(req, account) != nil {
			c.Response().WriteHeader(http.StatusBadRequest)
			return nil
		}

		switch actor.CreateAccount(system, tenant, *account).(type) {

		case *actor.AccountCreated:
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
			c.Response().WriteHeader(http.StatusOK)
			return nil

		case *actor.ReplyTimeout:
			c.Response().WriteHeader(http.StatusGatewayTimeout)
			return nil

		default:
			c.Response().WriteHeader(http.StatusConflict)
			return nil

		}
	}
}

// GetAccounts return existing accounts of given tenant
func GetAccounts(storage *localfs.PlaintextStorage) func(c echo.Context) error {
	return func(c echo.Context) error {
		tenant := c.Param("tenant")
		if tenant == "" {
			return fmt.Errorf("missing tenant")
		}

		accounts, err := persistence.LoadAccounts(storage, tenant)
		if err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)

		for idx, account := range accounts {
			if idx == len(accounts)-1 {
				c.Response().Write([]byte(account))
			} else {
				c.Response().Write([]byte(account + "\n"))
			}
			c.Response().Flush()
		}

		return nil
	}
}

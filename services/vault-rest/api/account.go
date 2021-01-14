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
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}
		id := c.Param("id")
		if id == "" {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
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
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}

		b, err := ioutil.ReadAll(c.Request().Body)
		defer c.Request().Body.Close()
		if err != nil {
			c.Response().WriteHeader(http.StatusBadRequest)
			return err
		}

		var req = new(model.Account)
		if json.Unmarshal(b, req) != nil {
			c.Response().WriteHeader(http.StatusBadRequest)
			return nil
		}

		switch actor.CreateAccount(system, tenant, *req).(type) {

		case *actor.AccountCreated:
			log.Info().Msgf("Account %s/%s Created", tenant, req.Name)
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
			c.Response().WriteHeader(http.StatusOK)
			return nil

		case *actor.ReplyTimeout:
			log.Warn().Msgf("Account %s/%s Accepted for Processing (Timeout)", tenant, req.Name)
			c.Response().WriteHeader(http.StatusGatewayTimeout)
			return nil

		default:
			log.Info().Msgf("Transaction %s/%s Created", tenant, req.Name)
			c.Response().WriteHeader(http.StatusConflict)
			return nil

		}
	}
}

// GetAccounts return existing accounts of given tenant
func GetAccounts(storage localfs.Storage) func(c echo.Context) error {
	return func(c echo.Context) error {
		tenant := c.Param("tenant")
		if tenant == "" {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
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
				c.Response().Write([]byte(account))
				c.Response().Write([]byte("\n"))
			}
			c.Response().Flush()
		}

		return nil
	}
}

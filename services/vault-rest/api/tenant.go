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
	"fmt"
	"strings"
	"net/http"

	"github.com/jancajthaml-openbank/vault-rest/system"

	"github.com/labstack/echo/v4"
)

// CreateTenant enables vault-unit@{tenant}
func CreateTenant(systemctl system.Control) func(c echo.Context) error {
	return func(c echo.Context) error {
		tenant := strings.TrimSpace(c.Param("tenant"))
		if tenant == "" {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}
		err := systemctl.EnableUnit("vault-unit@" + tenant + ".service")
		if err != nil {
			return err
		}
		c.Response().WriteHeader(http.StatusOK)
		return nil
	}
}

// DeleteTenant disables vault-unit@{tenant}
func DeleteTenant(systemctl system.Control) func(c echo.Context) error {
	return func(c echo.Context) error {
		tenant := strings.TrimSpace(c.Param("tenant"))
		if tenant == "" {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}
		err := systemctl.DisableUnit("vault-unit@" + tenant + ".service")
		if err != nil {
			return err
		}
		c.Response().WriteHeader(http.StatusOK)
		return nil
	}
}

// ListTenants lists vault-unit@
func ListTenants(systemctl system.Control) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		units, err := systemctl.ListUnits("vault-unit@")
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		for idx, unit := range units {
			if idx == len(units)-1 {
				c.Response().Write([]byte(unit))
			} else {
				c.Response().Write([]byte(unit))
				c.Response().Write([]byte("\n"))
			}
			c.Response().Flush()
		}

		return nil
	}
}

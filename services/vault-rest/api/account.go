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
	"io/ioutil"
	"net/http"

	"github.com/jancajthaml-openbank/vault-rest/actor"
	"github.com/jancajthaml-openbank/vault-rest/persistence"
	"github.com/jancajthaml-openbank/vault-rest/utils"

	"github.com/gorilla/mux"
)

// AccountPartial returns http handler for single account
func AccountPartial(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		tenant := vars["tenant"]
		id := vars["id"]

		if tenant == "" || id == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(emptyJSONObject)
			return
		}

		switch r.Method {

		case "GET":
			GetAccount(server, tenant, id, w, r)
			return

		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write(emptyJSONObject)
			return

		}
	}
}

// AccountsPartial returns http handler for accounts
func AccountsPartial(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		tenant := vars["tenant"]

		if tenant == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(emptyJSONArray)
			return
		}

		switch r.Method {

		case "GET":
			GetAccounts(server, tenant, w, r)
			return

		case "POST":
			CreateAccount(server, tenant, w, r)
			return

		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write(emptyJSONObject)
			return

		}

	}
}

// CreateAccount creates new account
func CreateAccount(server *Server, tenant string, w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONObject)
		return
	}

	var req = new(actor.Account)
	err = utils.JSON.Unmarshal(b, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(emptyJSONObject)
		return
	}

	switch actor.CreateAccount(server.ActorSystem, tenant, *req).(type) {

	case *actor.AccountCreated:
		resp, err := utils.JSON.Marshal(req)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(emptyJSONArray)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
		return

	case *actor.ReplyTimeout:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write(emptyJSONObject)
		return

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(emptyJSONObject)
		return

	}
}

// GetAccounts returns list of existing accounts
func GetAccounts(server *Server, tenant string, w http.ResponseWriter, r *http.Request) {
	accounts, err := persistence.LoadAccounts(server.Storage, tenant)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONArray)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp, err := utils.JSON.Marshal(accounts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONArray)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
	return
}

// GetAccount returns snapshot existing account
func GetAccount(server *Server, tenant string, id string, w http.ResponseWriter, r *http.Request) {
	switch result := actor.GetAccount(server.ActorSystem, tenant, id).(type) {

	case *actor.AccountMissing:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(emptyJSONObject)
		return

	case *actor.Account:
		w.Header().Set("Content-Type", "application/json")
		resp, err := utils.JSON.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(emptyJSONObject)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
		return

	case *actor.ReplyTimeout:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write(emptyJSONObject)
		return

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONObject)
		return

	}
	return
}

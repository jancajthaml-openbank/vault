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

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func loadConfFromEnv() Configuration {
	logOutput := getEnvString("VAULT_LOG", "")
	logLevel := strings.ToUpper(getEnvString("VAULT_LOG_LEVEL", "DEBUG"))
	secrets := getEnvString("VAULT_SECRETS", "")
	rootStorage := getEnvString("VAULT_STORAGE", "/data")
	lakeHostname := getEnvString("VAULT_LAKE_HOSTNAME", "")
	port := getEnvInteger("VAULT_HTTP_PORT", 4400)
	metricsOutput := getEnvString("VAULT_METRICS_OUTPUT", "")
	metricsRefreshRate := getEnvDuration("VAULT_METRICS_REFRESHRATE", time.Second)

	if lakeHostname == "" || secrets == "" || rootStorage == "" {
		log.Fatal("missing required parameter to run")
	}

	if metricsOutput != "" && os.MkdirAll(filepath.Dir(metricsOutput), os.ModePerm) != nil {
		log.Fatal("unable to assert metrics output")
	}

	cert, err := ioutil.ReadFile(secrets + "/domain.local.crt")
	if err != nil {
		log.Fatalf("unable to load certificate %s/domain.local.crt", secrets)
	}

	key, err := ioutil.ReadFile(secrets + "/domain.local.key")
	if err != nil {
		log.Fatalf("unable to load certificate %s/domain.local.key", secrets)
	}

	return Configuration{
		RootStorage:        rootStorage,
		ServerPort:         port,
		SecretKey:          key,
		SecretCert:         cert,
		LakeHostname:       lakeHostname,
		LogOutput:          logOutput,
		LogLevel:           logLevel,
		MetricsRefreshRate: metricsRefreshRate,
		MetricsOutput:      metricsOutput,
	}
}

func getEnvString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInteger(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := strconv.Atoi(value)
	if err != nil {
		log.Panicf("invalid value of variable %s", key)
	}
	return cast
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := time.ParseDuration(value)
	if err != nil {
		log.Panicf("invalid value of variable %s", key)
	}
	return cast
}

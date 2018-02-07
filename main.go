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

package main

import (
	"bufio"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/jancajthaml-openbank/vault/actor"
	"github.com/jancajthaml-openbank/vault/cron"
	"github.com/jancajthaml-openbank/vault/utils"
)

var (
	version string
	build   string
)

func setupLogOutput(params utils.RunParams) {
	if len(params.Log) == 0 {
		return
	}

	file, err := os.Create(params.Log)
	if err != nil {
		log.Warnf("Unable to create %s: %v", params.Log, err)
		return
	}
	defer file.Close()

	log.SetOutput(bufio.NewWriter(file))
}

func setupLogLevel(params utils.RunParams) {
	level, err := log.ParseLevel(params.LogLevel)
	if err != nil {
		log.Warnf("Invalid log level %v, using level WARN", params.LogLevel)
		return
	}
	log.Infof("Log level set to %v", strings.ToUpper(params.LogLevel))
	log.SetLevel(level)
}

func init() {
	viper.SetEnvPrefix("VAULT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("log.level", "DEBUG")
	viper.SetDefault("storage", "/data")
	viper.SetDefault("journal.saturation", 100)
	viper.SetDefault("snasphot.scaninteval", "1m")
	viper.SetDefault("metrics.refreshrate", "1s")
}

func validParams(params utils.RunParams) bool {
	if len(params.Tenant) == 0 || len(params.LakeHostname) == 0 {
		log.Error("missing required parameter to run")
		return false
	}

	if os.MkdirAll(params.RootStorage, os.ModePerm) != nil {
		log.Error("unable to assert storage directory")
		return false
	}

	if len(params.MetricsOutput) != 0 && os.MkdirAll(filepath.Dir(params.MetricsOutput), os.ModePerm) != nil {
		log.Error("unable to assert metrics output")
		return false
	}

	return true
}

func main() {
	params := utils.RunParams{
		RootStorage:          viper.GetString("storage") + "/" + viper.GetString("tenant"),
		Tenant:               viper.GetString("tenant"),
		LakeHostname:         viper.GetString("lake.hostname"),
		JournalSaturation:    viper.GetInt("journal.saturation"),
		Log:                  viper.GetString("log"),
		LogLevel:             viper.GetString("log.level"),
		SnapshotScanInterval: viper.GetDuration("snasphot.scaninteval"),
		MetricsRefreshRate:   viper.GetDuration("metrics.refreshrate"),
		MetricsOutput:        viper.GetString("metrics.output"),
	}

	if !validParams(params) {
		return
	}

	setupLogOutput(params)
	setupLogLevel(params)

	log.Infof(">>> Starting <<<")

	metrics := cron.NewMetrics()
	system := new(actor.ActorSystem)
	system.Start(params, metrics)
	defer system.Stop()

	var wg sync.WaitGroup
	terminationChan := make(chan struct{})
	wg.Add(2)

	go cron.SnapshotSaturationScan(&wg, terminationChan, params, metrics, system.ProcessLocalMessage)
	go cron.PersistMetrics(&wg, terminationChan, params, metrics)

	log.Infof(">>> Started <<<")

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, os.Interrupt, syscall.SIGTERM)
	<-exitSignal

	log.Infof(">>> Terminating <<<")
	close(terminationChan)
	wg.Wait()
	log.Infof(">>> Terminated <<<")
}

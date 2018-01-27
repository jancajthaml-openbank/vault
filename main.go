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
	"strings"
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
}

func setupRootDirectory(params utils.RunParams) {
	// FIXME check error
	os.MkdirAll(params.RootStorage, os.ModePerm)
}

func main() {
	params := utils.RunParams{
		RootStorage:       viper.GetString("storage") + "/" + viper.GetString("tenant"),
		Tenant:            viper.GetString("tenant"),
		JournalSaturation: viper.GetInt("journal.saturation"),
		Log:               viper.GetString("log"),
		LogLevel:          viper.GetString("log.level"),
	}

	// FIXME validate params right here

	setupLogOutput(params)
	setupLogLevel(params)
	setupRootDirectory(params)

	log.Infof(">>> Starting <<<")

	system := actor.StartSupervisor(params)

	go cron.SnapshotSaturationScan(params, system.ProcessLocalMessage)

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	log.Infof(">>> Terminating <<<")
	system.Stop()
}

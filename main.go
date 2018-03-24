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
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/jancajthaml-openbank/vault/actor"
	"github.com/jancajthaml-openbank/vault/cron"
	"github.com/jancajthaml-openbank/vault/metrics"
	"github.com/jancajthaml-openbank/vault/utils"

	"github.com/gin-gonic/gin"
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
	viper.SetDefault("http.port", 8080)
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
	log.Infof(">>> Setup <<<")

	params := utils.RunParams{
		RootStorage:          viper.GetString("storage") + "/" + viper.GetString("tenant"),
		Tenant:               viper.GetString("tenant"),
		LakeHostname:         viper.GetString("lake.hostname"),
		JournalSaturation:    viper.GetInt("journal.saturation"),
		Log:                  viper.GetString("log"),
		LogLevel:             viper.GetString("log.level"),
		HTTPPort:             viper.GetInt("http.port"),
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

	// FIXME separate into its own go routine to be stopable
	m := metrics.NewMetrics()

	gin.SetMode(gin.ReleaseMode)

	system := new(actor.ActorSystem)
	system.Start(params, m) // FIXME if there is no lake, application is stuck here

	// FIXME check if nil if so then return
	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.String(200, "")
	})

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", params.HTTPPort),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			exitSignal <- syscall.SIGTERM
		}
	}()

	var wg sync.WaitGroup
	terminationChan := make(chan struct{})
	wg.Add(2)

	go cron.SnapshotSaturationScan(&wg, terminationChan, params, m, system.ProcessLocalMessage)
	go metrics.PersistMetrics(&wg, terminationChan, params, m)

	log.Infof(">>> Started <<<")

	<-exitSignal

	log.Infof(">>> Terminating <<<")
	system.Stop()
	close(terminationChan)
	wg.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Infof(">>> Terminated <<<")
}

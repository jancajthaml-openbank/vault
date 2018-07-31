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

	"github.com/jancajthaml-openbank/vault/pkg/actor"
	"github.com/jancajthaml-openbank/vault/pkg/cron"
	"github.com/jancajthaml-openbank/vault/pkg/metrics"
	"github.com/jancajthaml-openbank/vault/pkg/utils"

	"github.com/gin-gonic/gin"
)

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

	log.SetFormatter(new(utils.LogFormat))
}

func validParams(params utils.RunParams) bool {
	if len(params.Setup.Tenant) == 0 || len(params.Setup.LakeHostname) == 0 {
		log.Error("missing required parameter to run")
		return false
	}

	if os.MkdirAll(params.Setup.RootStorage, os.ModePerm) != nil {
		log.Error("unable to assert storage directory")
		return false
	}

	if len(params.Metrics.Output) != 0 && os.MkdirAll(filepath.Dir(params.Metrics.Output), os.ModePerm) != nil {
		log.Error("unable to assert metrics output")
		return false
	}

	return true
}

func loadParams() utils.RunParams {
	return utils.RunParams{
		Setup: utils.SetupParams{
			RootStorage:  viper.GetString("storage") + "/" + viper.GetString("tenant"),
			Tenant:       viper.GetString("tenant"),
			LakeHostname: viper.GetString("lake.hostname"),
			Log:          viper.GetString("log"),
			LogLevel:     viper.GetString("log.level"),
			HTTPPort:     viper.GetInt("http.port"),
		},
		Journal: utils.JournalParams{
			SnapshotScanInterval: viper.GetDuration("snasphot.scaninteval"),
			JournalSaturation:    viper.GetInt("journal.saturation"),
		},
		Metrics: utils.MetricsParams{
			RefreshRate: viper.GetDuration("metrics.refreshrate"),
			Output:      viper.GetString("metrics.output"),
		},
	}
}

func setupRootDirectory(params utils.RunParams) bool {
	return os.MkdirAll(params.Setup.RootStorage, os.ModePerm) == nil
}

func main() {
	log.Info(">>> Setup <<<")

	params := loadParams()
	if !validParams(params) {
		return
	}

	if len(params.Setup.Log) == 0 {
		log.SetOutput(os.Stdout)
	} else if file, err := os.OpenFile(params.Setup.Log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err == nil {
		defer file.Close()
		log.SetOutput(bufio.NewWriter(file))
	} else {
		log.SetOutput(os.Stdout)
		log.Warnf("Unable to create %s: %v", params.Setup.Log, err)
	}

	if level, err := log.ParseLevel(params.Setup.LogLevel); err == nil {
		log.Infof("Log level set to %v", strings.ToUpper(params.Setup.LogLevel))
		log.SetLevel(level)
	} else {
		log.Warnf("Invalid log level %v, using level WARN", params.Setup.LogLevel)
		log.SetLevel(log.WarnLevel)
	}

	if !setupRootDirectory(params) {
		log.Errorf("unable to assert storage directory")
		return
	}

	log.Info(">>> Starting <<<")

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
		Addr:    fmt.Sprintf(":%d", params.Setup.HTTPPort),
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

	log.Info(">>> Started <<<")

	<-exitSignal

	log.Info(">>> Terminating <<<")
	system.Stop()
	close(terminationChan)
	wg.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Info(">>> Terminated <<<")
}

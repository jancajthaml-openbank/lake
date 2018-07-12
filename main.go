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
	"os"
	"os/signal"
	"syscall"

	"bufio"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/jancajthaml-openbank/lake/metrics"
	"github.com/jancajthaml-openbank/lake/relay"
	"github.com/jancajthaml-openbank/lake/utils"
)

var (
	version string
	build   string
)

func init() {
	viper.SetEnvPrefix("LAKE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("log.level", "DEBUG")
	viper.SetDefault("port.pull", 5562)
	viper.SetDefault("port.pub", 5561)
	viper.SetDefault("metrics.refreshrate", "1s")

	log.SetFormatter(new(utils.LogFormat))
}

func main() {
	log.Infof(">>> Setup <<<")

	params := utils.RunParams{
		PullPort:           viper.GetInt("port.pull"),
		PubPort:            viper.GetInt("port.pub"),
		Log:                viper.GetString("log"),
		LogLevel:           viper.GetString("log.level"),
		MetricsRefreshRate: viper.GetDuration("metrics.refreshrate"),
		MetricsOutput:      viper.GetString("metrics.output"),
	}

	if len(params.Log) == 0 {
		log.SetOutput(os.Stdout)
	} else if file, err := os.OpenFile(params.Log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err == nil {
		defer file.Close()
		log.SetOutput(bufio.NewWriter(file))
	} else {
		log.SetOutput(os.Stdout)
		log.Warnf("Unable to create %s: %v", params.Log, err)
	}

	if level, err := log.ParseLevel(params.LogLevel); err == nil {
		log.Infof("Log level set to %v", strings.ToUpper(params.LogLevel))
		log.SetLevel(level)
	} else {
		log.Warnf("Invalid log level %v, using level WARN", params.LogLevel)
		log.SetLevel(log.WarnLevel)
	}

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	// FIXME separate into its own go routine to be stopable
	m := metrics.NewMetrics()

	log.Infof(">>> Starting <<<")

	// FIXME need a kill channel here for gracefull shutdown
	go relay.StartQueue(params, m)

	log.Infof(">>> Started <<<")

	var wg sync.WaitGroup

	terminationChan := make(chan struct{})
	wg.Add(1)
	go metrics.PersistPeriodically(&wg, terminationChan, params, m)

	log.Infof(">>> Started <<<")

	<-exitSignal

	// FIXME gracefully empty queues and relay all messages before shutdown
	log.Infof(">>> Terminating <<<")
	close(terminationChan)
	wg.Wait()

	log.Infof(">>> Terminated <<<")
}

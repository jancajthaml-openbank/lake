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
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/jancajthaml-openbank/lake/commands"
)

var (
	version string
	build   string
)

func setupLogOutput(params commands.RunParams) {
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

func setupLogLevel(params commands.RunParams) {
	level, err := log.ParseLevel(params.LogLevel)
	if err != nil {
		log.Warnf("Invalid log level %v, using level WARN", params.LogLevel)
		return
	}
	log.Infof("Log level set to %v", strings.ToUpper(params.LogLevel))
	log.SetLevel(level)
}

func init() {
	viper.SetEnvPrefix("LAKE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("http.port", 8080)
	viper.SetDefault("log.level", "DEBUG")
}

func main() {
	log.Infof(">>> Setup <<<")

	params := commands.RunParams{
		PullPort: 5562,
		PubPort:  5561,
		Log:      viper.GetString("log"),
		LogLevel: viper.GetString("log.level"),
		HTTPPort: viper.GetInt("http.port"),
	}

	setupLogOutput(params)
	setupLogLevel(params)

	commands.Run(params)
}

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

package commands

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Run starts service with graceful shutdown given TERM signal
func Run(params RunParams) {
	log.Infof(">>> Starting <<<")

	// FIXME need a kill channel here for gracefull shutdown
	go StartQueue(params)

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	log.Infof(">>> Started <<<")

	<-exitSignal

	log.Infof(">>> Terminating <<<")
	// FIXME gracefully empty queues and relay all messages before shutdown
	log.Infof(">>> Terminated <<<")
}

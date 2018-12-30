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

package boot

import (
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jancajthaml-openbank/lake/utils"

	log "github.com/sirupsen/logrus"
)

// Stop stops the application
func (app Application) Stop() {
	close(app.interrupt)
}

// WaitReady wait for daemons to be ready
func (app Application) WaitReady(deadline time.Duration) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("Unknown panic")
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		ticker := time.NewTicker(deadline)

		select {
		case <-app.metrics.IsReady:
			wg.Done()
			ticker.Stop()
			return
		case <-ticker.C:
			panic("metrics was not ready within 5 seconds")
		}
	}()

	go func() {
		ticker := time.NewTicker(deadline)

		select {
		case <-app.relay.IsReady:
			wg.Done()
			ticker.Stop()
			return
		case <-ticker.C:
			panic("relay was not ready within 5 seconds")
		}
	}()

	wg.Wait()

	return
}

// WaitInterrupt wait for signal
func (app Application) WaitInterrupt() {
	<-app.interrupt
}

// Run runs the application
func (app Application) Run() {
	log.Info(">>> Start <<<")

	go app.metrics.Start()
	go app.relay.Start()

	if err := app.WaitReady(5 * time.Second); err != nil {
		log.Errorf("Error when starting daemons: %+v", err)
	} else {
		log.Info(">>> Started <<<")
		utils.NotifyServiceReady()
		signal.Notify(app.interrupt, syscall.SIGINT, syscall.SIGTERM)
		app.WaitInterrupt()
	}

	log.Info(">>> Stopping <<<")
	utils.NotifyServiceStopping()

	app.metrics.Stop()
	app.relay.Stop()
	app.cancel()

	log.Info(">>> Stop <<<")
}

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

package concurrent

import (
	"runtime"
	"context"
)

type OneShotDaemon struct {
	Worker
	name string
}

func NewOneShotDaemon(name string, worker Worker) Daemon {
	return OneShotDaemon{
		Worker: worker,
		name:   name,
	}
}

func (daemon OneShotDaemon) Done() <- chan interface{} {
	return daemon.Worker.Done()
}

func (daemon OneShotDaemon) Setup() error {
	return daemon.Worker.Setup()
}

func (daemon OneShotDaemon) Stop() {
	daemon.Worker.Cancel()
}

func (daemon OneShotDaemon) Start(parentContext context.Context, cancelFunction context.CancelFunc) {
	defer cancelFunction()
	runtime.LockOSThread()
	defer func() {
		recover()
		runtime.UnlockOSThread()
	}()
	err := daemon.Setup()
	if err != nil {
		log.Error().Msgf("Setup error %s daemon %+v", daemon.name, err.Error())
		return
	}
	go func() {
		for {
			select {
			case <-parentContext.Done():
				daemon.Cancel()
				return
			}
		}
	}()
	log.Info().Msgf("Start %s daemon", daemon.name)
	daemon.Work()
	<-daemon.Done()
	log.Info().Msgf("Stop %s daemon", daemon.name)
}

// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"context"
	"runtime"
	"sync"
)

// OneShotPinnedDaemon represent work happening only once pinned to os thread
type OneShotPinnedDaemon struct {
	Worker
	name       string
	cancelOnce sync.Once
	done       chan interface{}
}

// NewOneShotPinnedDaemon returns new daemon with given name for single work
// pinned to single os thread
func NewOneShotPinnedDaemon(name string, worker Worker) Daemon {
	return &OneShotPinnedDaemon{
		Worker:     worker,
		name:       name,
		cancelOnce: sync.Once{},
		done:       make(chan interface{}),
	}
}

// Done returns signal when worker has finished work
func (daemon *OneShotPinnedDaemon) Done() <-chan interface{} {
	if daemon == nil {
		done := make(chan interface{})
		close(done)
		return done
	}
	<-daemon.Worker.Done()
	return daemon.done
}

// Setup prepares worker for work
func (daemon *OneShotPinnedDaemon) Setup() error {
	if daemon == nil {
		return nil
	}
	return daemon.Worker.Setup()
}

// Stop cancels worker's work
func (daemon *OneShotPinnedDaemon) Stop() {
	if daemon == nil {
		return
	}
	daemon.cancelOnce.Do(func() {
		daemon.Worker.Cancel()
		close(daemon.done)
	})
}

// Start starts worker's work once
func (daemon *OneShotPinnedDaemon) Start(parentContext context.Context, cancelFunction context.CancelFunc) {
	defer cancelFunction()
	if daemon == nil {
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	err := daemon.Setup()
	if err != nil {
		log.Error().Err(err).Msgf("Setup daemon %s", daemon.name)
		return
	}
	go func() {
		select {
		case <-parentContext.Done():
			daemon.Stop()
			return
		case <-daemon.Done():
			return
		}
	}()

	log.Info().Msgf("Start daemon %s run once", daemon.name)
	defer log.Info().Msgf("Stop daemon %s", daemon.name)

	daemon.Work()
	<-daemon.Done()
}

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

package utils

import (
	"net"
	"fmt"
	"os"
)

func systemNotify(state string) error {
	socketAddr := &net.UnixAddr{
		Name: os.Getenv("NOTIFY_SOCKET"),
		Net:  "unixgram",
	}
	if socketAddr.Name == "" {
		return fmt.Errorf("NOTIFY_SOCKET is not set")
	}
	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.Write([]byte(state))
	return nil
}

// NotifyServiceReady notify underlying os that service is ready
func NotifyServiceReady() error {
	return systemNotify("READY=1")
}

// NotifyServiceStopping notify underlying os that service is stopping
func NotifyServiceStopping() error {
	return systemNotify("STOPPING=1")
}

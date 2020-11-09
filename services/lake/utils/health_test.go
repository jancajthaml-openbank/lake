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
  "os"
  "net"
  "bytes"
  "testing"
)

func TestNotifyServiceStatus(t *testing.T) {
  // FIXME tempfile
  name := "testgram"

  os.Setenv("NOTIFY_SOCKET", name)
  defer os.Unsetenv("NOTIFY_SOCKET")

  ta, err := net.ResolveUnixAddr("unixgram", name)
  if err != nil {
    t.Fatal(err)
  }
  l, err := net.ListenUnixgram("unixgram", ta)
  if err != nil {
    t.Fatal(err)
  }
  defer func() {
    l.Close()
    os.Remove(name)
  }()

  t.Log("NotifyServiceReady")
  {
    NotifyServiceReady()
    b := make([]byte, 64)
    n, _, err := l.ReadFrom(b)
    if err != nil {
      t.Fatal(err)
    }
    if !bytes.Equal(b[:n], []byte("READY=1")) {
      t.Fatalf("got %s; want READY=1", string(b[:n]))
    }
  }

  t.Log("NotifyServiceStopping")
  {
    NotifyServiceStopping()
    b := make([]byte, 64)
    n, _, err := l.ReadFrom(b)
    if err != nil {
      t.Fatal(err)
    }
    if !bytes.Equal(b[:n], []byte("STOPPING=1")) {
      t.Fatalf("got %s; want STOPPING=1", string(b[:n]))
    }
  }
}

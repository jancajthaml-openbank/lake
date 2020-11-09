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

package logging

import (
  "testing"
  "os"
  "github.com/rs/zerolog"
)

func TestMain(m *testing.M) {
  os.Exit(m.Run())
}

func TestNew(t *testing.T) {
  logger := New("test-logger")
  logger.Info().Msg("test message")
}

func TestSetupLogger(t *testing.T) {
  defer SetupLogger("DEBUG")

  t.Log("DEBUG")
  {
    SetupLogger("DEBUG")
    if zerolog.GlobalLevel() != zerolog.DebugLevel {
      t.Errorf("failed to set DEBUG log level")
    }
  }

  t.Log("INFO")
  {
    SetupLogger("INFO")
    if zerolog.GlobalLevel() != zerolog.InfoLevel {
      t.Errorf("failed to set INFO log level")
    }
  }

  t.Log("ERROR")
  {
    SetupLogger("ERROR")
    if zerolog.GlobalLevel() != zerolog.ErrorLevel {
      t.Errorf("failed to set ERROR log level")
    }
  }

  t.Log("UNKNOWN")
  {
    SetupLogger("UNKNOWN")
    if zerolog.GlobalLevel() != zerolog.InfoLevel {
      t.Errorf("failed to set fallback INFO log level")
    }
  }

}

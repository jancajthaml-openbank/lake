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

package config

import (
  "testing"
  "os"
)

func TestEnvBoolean(t *testing.T) {

  t.Log("LAKE_TEST_BOOL missing")
  {
    if envBoolean("LAKE_TEST_BOOL", true) != true {
      t.Errorf("envBoolean did not provide default value")
    }
  }

  t.Log("LAKE_TEST_BOOL present and valid")
  {
    os.Setenv("LAKE_TEST_BOOL", "false")
    defer os.Unsetenv("LAKE_TEST_BOOL")

    if envBoolean("LAKE_TEST_BOOL", true) != false {
      t.Errorf("envBoolean did not obtain env value")
    }
  }

  t.Log("LAKE_TEST_BOOL present and invalid")
  {
    os.Setenv("LAKE_TEST_BOOL", "x")
    defer os.Unsetenv("LAKE_TEST_BOOL")

    if envBoolean("LAKE_TEST_BOOL", true) != true {
      t.Errorf("envBoolean did not obtain fallback to default value")
    }
  }
}

func TestEnvInteger(t *testing.T) {

  t.Log("LAKE_TEST_INT missing")
  {
    if envInteger("LAKE_TEST_INT", 0) != 0 {
      t.Errorf("envInteger did not provide default value")
    }
  }

  t.Log("LAKE_TEST_INT present and valid")
  {
    os.Setenv("LAKE_TEST_INT", "1")
    defer os.Unsetenv("LAKE_TEST_INT")

    if envInteger("LAKE_TEST_INT", 0) != 1 {
      t.Errorf("envInteger did not obtain env value")
    }
  }

  t.Log("LAKE_TEST_INT present and invalid")
  {
    os.Setenv("LAKE_TEST_INT", "x")
    defer os.Unsetenv("LAKE_TEST_INT")

    if envInteger("LAKE_TEST_INT", 0) != 0 {
      t.Errorf("envInteger did not obtain fallback to default value")
    }
  }
}

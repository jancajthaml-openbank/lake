package config

import (
	"os"
	"testing"
	"time"
)

func TestEnvString(t *testing.T) {

	t.Log("LAKE_TEST_STR missing")
	{
		if envString("LAKE_TEST_STR", "x") != "x" {
			t.Errorf("envString did not provide default value")
		}
	}

	t.Log("LAKE_TEST_STR present")
	{
		os.Setenv("LAKE_TEST_STR", "y")
		defer os.Unsetenv("LAKE_TEST_STR")

		if envString("LAKE_TEST_STR", "x") != "y" {
			t.Errorf("envString did not obtain env value")
		}
	}
}

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

func TestEnvDuration(t *testing.T) {

	t.Log("LAKE_TEST_DUR missing")
	{
		if envDuration("LAKE_TEST_DUR", time.Second) != time.Second {
			t.Errorf("envDuration did not provide default value")
		}
	}

	t.Log("LAKE_TEST_DUR present and valid")
	{
		os.Setenv("LAKE_TEST_DUR", "2s")
		defer os.Unsetenv("LAKE_TEST_DUR")

		if envDuration("LAKE_TEST_DUR", time.Second) != 2*time.Second {
			t.Errorf("envDuration did not obtain env value")
		}
	}

	t.Log("LAKE_TEST_DUR present and invalid")
	{
		os.Setenv("LAKE_TEST_DUR", "x")
		defer os.Unsetenv("LAKE_TEST_DUR")

		if envDuration("LAKE_TEST_DUR", time.Second) != time.Second {
			t.Errorf("envDuration did not obtain fallback to default value")
		}
	}
}

func TestEnvFilename(t *testing.T) {

	t.Log("LAKE_TEST_FILE missing")
	{
		if envFilename("LAKE_TEST_FILE", "/tmp/stub") != "/tmp/stub" {
			t.Errorf("envFilename did not provide default value")
		}
	}

	t.Log("LAKE_TEST_FILE present and valid")
	{
		os.Setenv("LAKE_TEST_FILE", "/tmp")
		defer os.Unsetenv("LAKE_TEST_FILE")

		if envFilename("LAKE_TEST_FILE", "/tmp/stub") != "/tmp" {
			t.Errorf("envFilename did not obtain env value")
		}
	}

	t.Log("LAKE_TEST_FILE present and invalid")
	{
		os.Setenv("LAKE_TEST_FILE", "/dev/null")
		defer os.Unsetenv("LAKE_TEST_FILE")

		if envFilename("LAKE_TEST_FILE", "/tmp/stub") != "/tmp/stub" {
			t.Errorf("envFilename did not obtain fallback to default value")
		}
	}
}

package web

import (
	"os"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

type myConfig struct {
	Port              string `env:"PORT" env_default:"8080"`
	ExpirationMinutes int    `env:"EXPIRATION" env_default:"1234"`
}

func TestMockEnvVar(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("TEST_VAR", "hello world")
	func() {
		defer MockEnvVar("TEST_VAR", "fuzzy buzzy")()
		assert.Equal("fuzzy buzzy", os.Getenv("TEST_VAR"))
	}()

	assert.Equal("hello world", os.Getenv("TEST_VAR"))
}

func TestReadConfigFromEnvironment(t *testing.T) {
	assert := assert.New(t)

	defer MockEnvVar("PORT", "")()
	defer MockEnvVar("EXPIRATION", "")()

	conf, err := ReadConfigFromEnvironment(&myConfig{})
	assert.Nil(err)

	assert.Equal("8080", conf.(*myConfig).Port)
	assert.Equal(1234, conf.(*myConfig).ExpirationMinutes)

	os.Setenv("PORT", "80")
	os.Setenv("EXPIRATION", "4321")

	conf, err = ReadConfigFromEnvironment(&myConfig{})
	assert.Nil(err)

	assert.Equal("80", conf.(*myConfig).Port)
	assert.Equal(4321, conf.(*myConfig).ExpirationMinutes)
}

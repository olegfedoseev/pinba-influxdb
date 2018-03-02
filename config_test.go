package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetValidConfig(t *testing.T) {
	got, err := getConfig("config.yml.test")
	assert.NoError(t, err)
	expected := &config{
		Pinba: struct {
			Addr           string        `yaml:"host"`
			ConnectTimeout time.Duration `yaml:"connect_timeout"`
			ReadTimeout    time.Duration `yaml:"read_timeout"`
		}{
			Addr:           "pinba-host:5002",
			ConnectTimeout: 500 * time.Millisecond,
			ReadTimeout:    5 * time.Second,
		},
		Influxdb: struct {
			Addr     string `yaml:"addr"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			Database string `yaml:"database"`
		}{
			Addr:     "http://influxdb:8086",
			User:     "username",
			Password: "secret",
			Database: "pinba-database",
		},
		Whitelist: struct {
			ServerNames []string `yaml:"servers"`
			Tags        []string `yaml:"tags"`
		}{
			ServerNames: []string{"example.com", "api.example.com"},
			Tags:        []string{"server", "region", "script", "status"},
		},
	}
	assert.Equal(t, expected, got)
}

func TestGetConfigFromNotYamlFile(t *testing.T) {
	_, err := getConfig("config.go")
	assert.EqualError(t, err, "unmarshal failed: yaml: line 38: mapping values are not allowed in this context")
}

func TestGetConfigFromNotExistigFile(t *testing.T) {
	_, err := getConfig("no-such-file")
	assert.EqualError(t, err, "open no-such-file: no such file or directory")
}

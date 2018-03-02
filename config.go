package main

import (
	"fmt"
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Pinba struct {
		Addr           string        `yaml:"host"`
		ConnectTimeout time.Duration `yaml:"connect_timeout"`
		ReadTimeout    time.Duration `yaml:"read_timeout"`
	} `yaml:"pinba"`

	Influxdb struct {
		Addr     string `yaml:"addr"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"influxdb"`

	Whitelist struct {
		ServerNames []string `yaml:"servers"`
		Tags        []string `yaml:"tags"`
	} `yaml:"whitelist"`
}

func getConfig(filename string) (*config, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := config{}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}
	return &cfg, nil
}

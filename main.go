package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	pinba "github.com/olegfedoseev/pinba-server/client"
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
		return nil, fmt.Errorf("reading file %s failed: %v", filename, err)
	}

	cfg := config{}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}
	return &cfg, nil
}

func main() {
	var configFile = flag.String("config", "config.yml", "config name, default - config.yml")
	flag.Parse()

	config, err := getConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config from %v: %v", *configFile, err)
	}

	// Create a new HTTPClient
	influxdbClient, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:      config.Influxdb.Addr,
		Username:  config.Influxdb.User,
		Password:  config.Influxdb.Password,
		UserAgent: "pinba-influxer",
	})
	if err != nil {
		log.Fatal(err)
	}

	pinbaClient, err := pinba.New(
		config.Pinba.Addr,
		config.Pinba.ConnectTimeout,
		config.Pinba.ReadTimeout,
	)
	if err != nil {
		log.Fatalf("Failed to create pinba client: %v", err)
	}

	go pinbaClient.Listen(1)

	for {
		select {
		case requests := <-pinbaClient.Requests:
			// Create a new point batch, error can be only in Precision parsing
			batch, _ := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
				Database: config.Influxdb.Database,
			})

			cnt := 0
			for _, request := range requests.Requests {
				server, _ := request.Tags.Get("server")
				if !in(server, config.Whitelist.ServerNames) {
					continue
				}

				// Create a point and add to batch
				datapoint, err := influxdb.NewPoint(
					"requests", // table name
					request.Tags.Filter(config.Whitelist.Tags).GetMap(),
					map[string]interface{}{
						"request_time": request.RequestTime,
					},
					time.Unix(requests.Timestamp, 0),
				)
				if err != nil {
					log.Printf("[ERROR] Failed to create datapoint: %v", err)
					continue
				}
				cnt++
				batch.AddPoint(datapoint)

				// Request's Timers
				for _, timer := range request.Timers {
					datapoint, err := influxdb.NewPoint(
						"timers", // table name
						timer.Tags.Filter(config.Whitelist.Tags).GetMap(),
						map[string]interface{}{
							"value": timer.Value,
							"hits":  int64(timer.HitCount),
						},
						time.Unix(requests.Timestamp, 0),
					)
					if err != nil {
						log.Printf("[ERROR] Failed to create datapoint: %v", err)
						continue
					}
					batch.AddPoint(datapoint)
					cnt++
				}
			}

			// Write the batch
			t := time.Now()
			if err := influxdbClient.Write(batch); err != nil {
				log.Fatal(err)
			}
			log.Printf("Writen %d datapoints in %v\n", cnt, time.Since(t))
		}
	}
}

func in(val string, slice []string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

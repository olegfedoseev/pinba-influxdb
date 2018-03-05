package main

import (
	"flag"
	"log"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	pinba "github.com/olegfedoseev/pinba-server/client"
)

var version = "master"

func newInfluxdbClient(cfg *config, userAgent string) (influxdb.Client, error) {
	if strings.HasPrefix(cfg.Influxdb.Addr, "http") {
		return influxdb.NewHTTPClient(influxdb.HTTPConfig{
			Addr:      cfg.Influxdb.Addr,
			Username:  cfg.Influxdb.User,
			Password:  cfg.Influxdb.Password,
			UserAgent: userAgent,
		})
	}

	return influxdb.NewUDPClient(influxdb.UDPConfig{
			Addr: cfg.Influxdb.Addr,
	})
}

func main() {
	log.Println(version)

	var configFile = flag.String("config", "config.yml", "config name, default - config.yml")
	flag.Parse()

	config, err := getConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config from %v: %v", *configFile, err)
	}

	influxdbClient, err := newInfluxdbClient(
		config,
		"pinba-influxer",
	)

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

			var cnt int64
			for _, request := range requests.Requests {
				server, _ := request.Tags.Get("server")
				if len(config.Whitelist.ServerNames) > 0 && !in(server, config.Whitelist.ServerNames) {
					continue
				}

				// Create a point and add to batch
				datapoint, err := influxdb.NewPoint(
					"requests", // table name
					request.Tags.Filter(config.Whitelist.Tags).GetMap(),
					map[string]interface{}{
						"request_time": request.RequestTime,
					},
					time.Unix(requests.Timestamp, cnt), // cnt as nsec - so that all of our metrics have unique timestamp
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
						time.Unix(requests.Timestamp, cnt), // cnt as nsec - so that all of our metrics have unique timestamp
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

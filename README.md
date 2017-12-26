# pinba-influxer

## InfluxDB Preparation

Create database:
    CREATE DATABASE "pinba"

Create retention policy:
    CREATE RETENTION POLICY "realtime" ON "pinba" DURATION 2h REPLICATION 1 DEFAULT
    CREATE RETENTION POLICY "month" ON "pinba" DURATION 4w REPLICATION 1
    CREATE RETENTION POLICY "year" ON "pinba" DURATION 52w REPLICATION 1
    CREATE RETENTION POLICY "infinity" ON "pinba" DURATION INF REPLICATION 1

And now, interesting part - continuous queries.

For downsampling realtime data to 10s:
    CREATE CONTINUOUS QUERY "cq_requests_10s" ON "pinba" BEGIN
        SELECT
            count("request_time") / 10 AS "rps",
            percentile("request_time", 25) AS "p25",
            percentile("request_time", 75) AS "p75",
            percentile("request_time", 95) AS "p95",
            max("request_time") AS "max"
        INTO "month"."requests"
        FROM "requests"
        GROUP BY time(10s), * END

    CREATE CONTINUOUS QUERY "cq_timers_10s" ON "pinba" BEGIN
        SELECT
            sum("hits") / 10 AS "rps",
            percentile("value", 25) AS "p25",
            percentile("value", 75) AS "p75",
            percentile("value", 95) AS "p95",
            max("value") AS "max"
        INTO "month"."timers"
        FROM "timers"
        GROUP BY time(10s), * END

For downsampling realtime data to 60s:
    CREATE CONTINUOUS QUERY "cq_requests_60s" ON "pinba" BEGIN
        SELECT
            count("request_time") / 60 AS "rps",
            percentile("request_time", 25) AS "p25",
            percentile("request_time", 75) AS "p75",
            percentile("request_time", 95) AS "p95",
            max("request_time") AS "max"
        INTO "year"."requests"
        FROM "requests"
        GROUP BY time(60s), * END

    CREATE CONTINUOUS QUERY "cq_timers_60s" ON "pinba" BEGIN
        SELECT
            sum("hits") / 60 AS "rps",
            percentile("value", 25) AS "p25",
            percentile("value", 75) AS "p75",
            percentile("value", 95) AS "p95",
            max("value") AS "max"
        INTO "year"."timers"
        FROM "timers"
        GROUP BY time(60s), * END

And for downsampling realtime data to 10m:
    CREATE CONTINUOUS QUERY "cq_requests_10m" ON "pinba" BEGIN
        SELECT
            count("request_time") / 600 AS "rps",
            percentile("request_time", 25) AS "p25",
            percentile("request_time", 75) AS "p75",
            percentile("request_time", 95) AS "p95",
            max("request_time") AS "max"
        INTO "year"."requests"
        FROM "requests"
        GROUP BY time(10m), * END

    CREATE CONTINUOUS QUERY "cq_timers_10m" ON "pinba" BEGIN
        SELECT
            sum("hits") / 600 AS "rps",
            percentile("value", 25) AS "p25",
            percentile("value", 75) AS "p75",
            percentile("value", 95) AS "p95",
            max("value") AS "max"
        INTO "year"."timers"
        FROM "timers"
        GROUP BY time(10m), * END




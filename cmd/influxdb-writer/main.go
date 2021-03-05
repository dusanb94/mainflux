// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	influxdata "github.com/influxdata/influxdb-client-go/v2"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/writers/api"
	"github.com/mainflux/mainflux/consumers/writers/influxdb"
	"github.com/mainflux/mainflux/logger"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/transformers"
	"github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	svcName = "influxdb-writer"

	defNatsURL  = "nats://localhost:4222"
	defLogLevel = "error"
	defPort     = "8180"
	// defDB       = "mainflux"
	defDBHost   = "localhost"
	defDBPort   = "8086"
	defDBToken  = "mainflux-secret-token"
	defDBOrg    = "mainflux"
	defDBBucket = "messages"
	// defDBUser      = "mainflux"
	// defDBPass      = "mainflux"
	defConfigPath  = "/config.toml"
	defContentType = "application/senml+json"
	defTransformer = "senml"

	envNatsURL  = "MF_NATS_URL"
	envLogLevel = "MF_INFLUX_WRITER_LOG_LEVEL"
	envPort     = "MF_INFLUX_WRITER_PORT"
	// envDB       = "MF_INFLUXDB_DB"
	envDBHost   = "MF_INFLUX_WRITER_DB_HOST"
	envDBPort   = "MF_INFLUXDB_PORT"
	envDBToken  = "MF_INFLUXDB_ADMIN_TOKEN"
	envDBOrg    = "MF_INFLUXDB_ORG"
	envDBBucket = "MF_INFLUXDB_BUCKET"
	// envDBUser      = "MF_INFLUXDB_ADMIN_USER"
	// envDBPass      = "MF_INFLUXDB_ADMIN_PASSWORD"
	envConfigPath  = "MF_INFLUX_WRITER_CONFIG_PATH"
	envContentType = "MF_INFLUX_WRITER_CONTENT_TYPE"
	envTransformer = "MF_INFLUX_WRITER_TRANSFORMER"
)

type config struct {
	natsURL  string
	logLevel string
	port     string
	dbName   string
	dbHost   string
	dbPort   string
	dbOrg    string
	dbBucket string
	// dbUser      string
	// dbPass      string
	dbToken     string
	configPath  string
	contentType string
	transformer string
}

func main() {
	cfg := loadConfig()

	logger, err := mflog.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	pubSub, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	var lvl mflog.Level
	if err := lvl.UnmarshalText(cfg.logLevel); err != nil {
		logger.Error(fmt.Sprintf("Invalid log level: %s", err))
	}

	opts := influxdata.DefaultOptions()
	opts.SetLogLevel(uint(lvl))

	// clientCfg := influxdata.HTTPConfig{
	addr := fmt.Sprintf("http://%s:%s", cfg.dbHost, cfg.dbPort)
	// 	Username: cfg.dbUser,
	// 	Password: cfg.dbPass,
	// }

	client := influxdata.NewClientWithOptions(addr, cfg.dbToken, opts)
	// if err != nil {
	// 	logger.Error(fmt.Sprintf("Failed to create InfluxDB client: %s", err))
	// 	os.Exit(1)
	// }
	defer client.Close()
	writeAPI := client.WriteAPI(cfg.dbOrg, cfg.dbBucket)
	go func() {
		for {
			err := <-writeAPI.Errors()
			logger.Warn(fmt.Sprintf("Error writing data to InfluxDB: %s", err))
		}
	}()

	repo := influxdb.New(writeAPI)

	counter, latency := makeMetrics()
	repo = api.LoggingMiddleware(repo, logger)
	repo = api.MetricsMiddleware(repo, counter, latency)
	t := makeTransformer(cfg, logger)

	if err := consumers.Start(pubSub, repo, t, cfg.configPath, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to start InfluxDB writer: %s", err))
		os.Exit(1)
	}

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go startHTTPService(cfg.port, logger, errs)

	err = <-errs
	logger.Error(fmt.Sprintf("InfluxDB writer service terminated: %s", err))
}

func loadConfig() config {
	cfg := config{
		natsURL:  mainflux.Env(envNatsURL, defNatsURL),
		logLevel: mainflux.Env(envLogLevel, defLogLevel),
		port:     mainflux.Env(envPort, defPort),
		// dbName:   mainflux.Env(envDB, defDB),
		dbHost:   mainflux.Env(envDBHost, defDBHost),
		dbPort:   mainflux.Env(envDBPort, defDBPort),
		dbToken:  mainflux.Env(envDBToken, defDBToken),
		dbOrg:    mainflux.Env(envDBOrg, defDBOrg),
		dbBucket: mainflux.Env(envDBBucket, defDBBucket),
		// dbUser:      mainflux.Env(envDBUser, defDBUser),
		// dbPass:      mainflux.Env(envDBPass, defDBPass),
		configPath:  mainflux.Env(envConfigPath, defConfigPath),
		contentType: mainflux.Env(envContentType, defContentType),
		transformer: mainflux.Env(envTransformer, defTransformer),
	}

	return cfg
}

func makeMetrics() (*kitprometheus.Counter, *kitprometheus.Summary) {
	counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "influxdb",
		Subsystem: "message_writer",
		Name:      "request_count",
		Help:      "Number of database inserts.",
	}, []string{"method"})

	latency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "influxdb",
		Subsystem: "message_writer",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of inserts in microseconds.",
	}, []string{"method"})

	return counter, latency
}

func makeTransformer(cfg config, logger logger.Logger) transformers.Transformer {
	switch strings.ToUpper(cfg.transformer) {
	case "SENML":
		logger.Info("Using SenML transformer")
		return senml.New(cfg.contentType)
	case "JSON":
		logger.Info("Using JSON transformer")
		return json.New()
	default:
		logger.Error(fmt.Sprintf("Can't create transformer: unknown transformer type %s", cfg.transformer))
		os.Exit(1)
		return nil
	}
}

func startHTTPService(port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("InfluxDB writer service started, exposed port %s", p))
	errs <- http.ListenAndServe(p, api.MakeHandler(svcName))
}

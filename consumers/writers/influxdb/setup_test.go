// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package influxdb_test

import (
	"fmt"
	"os"
	"testing"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	dockertest "github.com/ory/dockertest/v3"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		testLog.Error(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	cfg := []string{
		"DOCKER_INFLUXDB_INIT_USERNAME=test",
		"DOCKER_INFLUXDB_INIT_PASSWORD=test",
		"DOCKER_INFLUXDB_INIT_MODE=setup",
		fmt.Sprintf("DOCKER_INFLUXDB_INIT_ORG=%s", org),
		fmt.Sprintf("DOCKER_INFLUXDB_INIT_BUCKET=%s", bucket),
		fmt.Sprintf("DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=%s", authToken),
	}
	container, err := pool.Run("influxdb", "2.0", cfg)
	if err != nil {
		testLog.Error(fmt.Sprintf("Could not start container: %s", err))
	}

	port = container.GetPort("8086/tcp")
	addr := fmt.Sprintf("http://localhost:%s", port)

	client := influxdb.NewClient(addr, authToken)

	if err := pool.Retry(func() error {
		writeAPI = client.WriteAPI(org, bucket)
		deleteAPI = client.DeleteAPI()
		queryAPI = client.QueryAPI(org)
		// _, err = client.Health(context.Background())
		return nil
	}); err != nil {
		testLog.Error(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	go func() {
		for {
			err := <-writeAPI.Errors()
			testLog.Warn(fmt.Sprintf("Error writing: %s", err))
		}
	}()

	code := m.Run()

	if err := pool.Purge(container); err != nil {
		testLog.Error(fmt.Sprintf("Could not purge container: %s", err))
	}

	os.Exit(code)
}

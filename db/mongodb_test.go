//go:build integration

package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mongoURI string
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "4.4",
		Env: []string{
			"MONGO_INITDB_DATABASE=test_db",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	mongoURI = fmt.Sprintf("mongodb://localhost:%s", resource.GetPort("27017/tcp"))

	if err := pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := ConnectMongoDB(mongoURI, ctx)
		if err != nil {
			return err
		}
		defer client.Disconnect(ctx)
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestConnectMongoDB_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := ConnectMongoDB(mongoURI, ctx)
	require.NoError(t, err)
	require.NotNil(t, client)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	err = client.Disconnect(ctx)
	require.NoError(t, err)
}

func TestConnectMongoDB_InvalidURI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	invalidURI := "mongodb://nonexistent-host:27017"
	client, err := ConnectMongoDB(invalidURI, ctx)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestConnectMongoDB_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := ConnectMongoDB(mongoURI, ctx)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestConnectMongoDB_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	client, err := ConnectMongoDB(mongoURI, ctx)
	assert.Error(t, err)
	assert.Nil(t, client)
}

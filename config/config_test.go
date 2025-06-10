//go:build unit

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("unmarshal error", func(t *testing.T) {
		os.Setenv("MONGO_TIMEOUT", "string")
		defer os.Unsetenv("MONGO_TIMEOUT")

		_, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to decode config into struct")
	})

	t.Run("default values", func(t *testing.T) {
		// Clear any environment variables that might affect the test
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("MONGO_URI")
		os.Unsetenv("MONGO_DB")
		os.Unsetenv("MONGO_TIMEOUT")
		os.Unsetenv("APP_POSTS_PER_PAGE")
		os.Unsetenv("CONTAINER")

		config, err := Load()
		require.NoError(t, err)

		// Check default values
		assert.Equal(t, "localhost", config.Server.Host)
		assert.Equal(t, "8080", config.Server.Port)
		assert.Equal(t, "mongodb://localhost:27017", config.Mongo.URI)
		assert.Equal(t, "news_app", config.Mongo.DB)
		assert.Equal(t, 10, config.Mongo.Timeout)
		assert.Equal(t, 12, config.App.PostsPerPage)
	})

	t.Run("environment variables override defaults", func(t *testing.T) {
		os.Setenv("SERVER_HOST", "127.0.0.1")
		os.Setenv("SERVER_PORT", "3000")
		os.Setenv("MONGO_URI", "mongodb://testhost:27017")
		os.Setenv("MONGO_DB", "test_db")
		os.Setenv("MONGO_TIMEOUT", "5")
		os.Setenv("APP_POSTS_PER_PAGE", "20")

		defer func() {
			os.Unsetenv("SERVER_HOST")
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("MONGO_URI")
			os.Unsetenv("MONGO_DB")
			os.Unsetenv("MONGO_TIMEOUT")
			os.Unsetenv("APP_POSTS_PER_PAGE")
		}()

		config, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "127.0.0.1", config.Server.Host)
		assert.Equal(t, "3000", config.Server.Port)
		assert.Equal(t, "mongodb://testhost:27017", config.Mongo.URI)
		assert.Equal(t, "test_db", config.Mongo.DB)
		assert.Equal(t, 5, config.Mongo.Timeout)
		assert.Equal(t, 20, config.App.PostsPerPage)
	})
}

func TestIsRunningInContainer(t *testing.T) {
	t.Run("not in container", func(t *testing.T) {
		os.Unsetenv("CONTAINER")
		os.Unsetenv("SERVER_HOST")

		assert.False(t, isRunningInContainer())
	})

	t.Run("CONTAINER env var set", func(t *testing.T) {
		os.Setenv("CONTAINER", "true")
		defer os.Unsetenv("CONTAINER")

		assert.True(t, isRunningInContainer())
	})

	t.Run("SERVER_HOST env var set", func(t *testing.T) {
		os.Unsetenv("CONTAINER")
		os.Setenv("SERVER_HOST", "0.0.0.0")
		defer os.Unsetenv("SERVER_HOST")

		assert.True(t, isRunningInContainer())
	})
}

func TestSetDefaults(t *testing.T) {
	t.Run("host is 0.0.0.0 when in container", func(t *testing.T) {
		os.Setenv("CONTAINER", "true")
		defer os.Unsetenv("CONTAINER")

		config, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "0.0.0.0", config.Server.Host)
	})

	t.Run("host is localhost when not in container", func(t *testing.T) {
		os.Unsetenv("CONTAINER")
		os.Unsetenv("SERVER_HOST")

		config, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "localhost", config.Server.Host)
	})
}

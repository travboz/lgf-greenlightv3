package env_test

import (
	"os"
	"testing"

	"github.com/travboz/lets-go/snippetboxv3/internal/assert"
	"github.com/travboz/lets-go/snippetboxv3/internal/env"
)

func TestGetString(t *testing.T) {
	t.Run("environment variable is set", func(t *testing.T) {
		key := "SOME_SET_KEY"

		want := "value"
		os.Setenv(key, want)
		defer os.Unsetenv(key) // clean up env

		got := env.GetString(key, "fallback")

		assert.AreEqual(t, got, want)

	})

	t.Run("environment variable is NOT set", func(t *testing.T) {
		key := "SOME_SET_KEY"

		want := "fallback"
		got := env.GetString(key, "fallback")

		assert.AreEqual(t, got, want)
	})
}

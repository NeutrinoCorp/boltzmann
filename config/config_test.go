package config_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/neutrinocorp/boltzmann/config"
)

const envPrefix = "SOME_PREFIX"

func TestSetEnvPrefix(t *testing.T) {
	config.SetEnvPrefix(envPrefix)
	config.SetDefault("some_key", "some value")
	out := config.Get[string]("some_key")
	assert.NotEmpty(t, out)
}

func TestGetEnv(t *testing.T) {
	config.SetEnvPrefix(envPrefix)
	tests := []struct {
		name   string
		valStr string
		exp    any
	}{
		{
			name:   "string",
			valStr: "some str",
			exp:    "some str",
		},
		{
			name:   "boolean",
			valStr: "true",
			exp:    true,
		},
		{
			name:   "duration",
			valStr: time.Hour.String(),
			exp:    time.Hour,
		},
		{
			name:   "int",
			valStr: "99",
			exp:    99,
		},
		{
			name:   "int8",
			valStr: "127",
			exp:    int8(127),
		},
		{
			name:   "int16",
			valStr: "127",
			exp:    int16(127),
		},
		{
			name:   "uint",
			valStr: "99",
			exp:    uint(99),
		},
		{
			name:   "uint8",
			valStr: "255",
			exp:    uint8(255),
		},
		{
			name:   "float32",
			valStr: "99.99",
			exp:    float32(99.99),
		},
		{
			name:   "float64",
			valStr: "99.99",
			exp:    99.99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envPrefix+"_"+tt.name, tt.valStr)
			var out any
			switch tt.exp.(type) {
			case string:
				out = config.Get[string](tt.name)
			case bool:
				out = config.Get[bool](tt.name)
			case time.Duration:
				out = config.Get[time.Duration](tt.name)
			case int:
				out = config.Get[int](tt.name)
			case int8:
				out = config.Get[int8](tt.name)
			case int16:
				out = config.Get[int16](tt.name)
			case uint:
				out = config.Get[uint](tt.name)
			case uint8:
				out = config.Get[uint8](tt.name)
			case float32:
				out = config.Get[float32](tt.name)
			case float64:
				out = config.Get[float64](tt.name)
			}
			assert.Equal(t, tt.exp, out)
		})
	}
}

func TestSetDefault(t *testing.T) {
	config.SetEnvPrefix(envPrefix)
	const key = "some_missing_key"
	val := time.Hour
	config.SetDefault(key, val)
	out := config.Get[time.Duration](key)
	assert.Equal(t, val, out)
}

func BenchmarkGetDefault(b *testing.B) {
	config.SetEnvPrefix(envPrefix)
	const key = "some_missing_key"
	val := int8(120)
	config.SetDefault(key, val)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		config.Get[int8](key)
	}
}

func BenchmarkGetInt(b *testing.B) {
	config.SetEnvPrefix(envPrefix)
	const key = "some_key"
	val := int8(120)
	b.Setenv(key, strconv.FormatInt(int64(val), 10))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		config.Get[int8](key)
	}
}

func BenchmarkGet(b *testing.B) {
	config.SetEnvPrefix(envPrefix)
	const key = "some_key"
	val := "some value"
	b.Setenv(key, val)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		config.Get[string](key)
	}
}

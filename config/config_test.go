package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/neutrinocorp/boltzmann/config"
)

func TestGetEnv(t *testing.T) {
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
			_ = os.Setenv(tt.name, tt.valStr)
			var out any
			switch tt.exp.(type) {
			case string:
				out = config.GetEnv[string](tt.name)
			case bool:
				out = config.GetEnv[bool](tt.name)
			case time.Duration:
				out = config.GetEnv[time.Duration](tt.name)
			case int:
				out = config.GetEnv[int](tt.name)
			case int8:
				out = config.GetEnv[int8](tt.name)
			case int16:
				out = config.GetEnv[int16](tt.name)
			case uint:
				out = config.GetEnv[uint](tt.name)
			case uint8:
				out = config.GetEnv[uint8](tt.name)
			case float32:
				out = config.GetEnv[float32](tt.name)
			case float64:
				out = config.GetEnv[float64](tt.name)
			}
			assert.Equal(t, tt.exp, out)
		})
	}
}

func TestSetDefault(t *testing.T) {
	const key = "some_missing_key"
	val := time.Hour
	config.SetDefault(key, val)
	out := config.GetEnv[time.Duration](key)
	assert.Equal(t, val, out)
}

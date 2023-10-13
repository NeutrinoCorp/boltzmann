package config

import (
	"os"
	"strconv"
	"time"
)

var defaultMap = map[string]any{}

var defaultEnvPrefix string

// SetEnvPrefix sets the configuration global prefix. Meaning each environment variable retrieved or set will
// require/attach the given prefix.
//
// E.g. QUEUE_NAME -> {PREFIX}_QUEUE_NAME.
func SetEnvPrefix(prefix string) {
	defaultEnvPrefix = prefix
}

func getDefaultEnvKey(key string) string {
	if defaultEnvPrefix == "" {
		return key
	}

	return defaultEnvPrefix + "_" + key
}

// SetDefault sets a default value for the given key. Make sure to use the exact data type for the value
// to avoid further casting issues.
func SetDefault(key string, val any) {
	key = getDefaultEnvKey(key)
	defaultMap[key] = val
}

// Get retrieves an environment variable based on the given key. Use T to enforce value casting to T type.
//
// If the value has no default, prefer using string types (environment variable default).
func Get[T any](key string) T {
	key = getDefaultEnvKey(key)
	val := os.Getenv(key)
	if val != "" {
		var typeOf T
		parsedVal, _ := parseValue[T](val, typeOf).(T)
		return parsedVal
	}

	defVal, ok := defaultMap[key]
	if !ok {
		var zeroVal T
		return zeroVal
	}
	castedVal, _ := defVal.(T)
	return castedVal
}

func parseValue[T any](src string, typeOf any) any {
	switch typeOf.(type) {
	case string:
		return src
	case bool:
		out, _ := strconv.ParseBool(src)
		return out
	case time.Duration:
		out, _ := time.ParseDuration(src)
		return out
	case float32:
		out, _ := strconv.ParseFloat(src, 32)
		return float32(out)
	case float64:
		out, _ := strconv.ParseFloat(src, 64)
		return out
	case int:
		out, _ := strconv.Atoi(src)
		return out
	case int8:
		out, _ := strconv.ParseInt(src, 10, 8)
		return int8(out)
	case int16:
		out, _ := strconv.ParseInt(src, 10, 16)
		return int16(out)
	case int32:
		out, _ := strconv.ParseInt(src, 10, 32)
		return int32(out)
	case int64:
		out, _ := strconv.ParseInt(src, 10, 64)
		return out
	case uint:
		out, _ := strconv.ParseUint(src, 10, 64)
		return uint(out)
	case uint8:
		out, _ := strconv.ParseUint(src, 10, 8)
		return uint8(out)
	case uint16:
		out, _ := strconv.ParseUint(src, 10, 16)
		return uint16(out)
	case uint32:
		out, _ := strconv.ParseUint(src, 10, 32)
		return uint32(out)
	case uint64:
		out, _ := strconv.ParseUint(src, 10, 64)
		return out
	default:
		var zeroVal T
		return zeroVal
	}
}

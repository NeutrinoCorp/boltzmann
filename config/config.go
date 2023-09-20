package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var defaultMap = map[string]any{}

var DefaultEnvPrefix string

func SetDefault(key string, val any) {
	if DefaultEnvPrefix != "" {
		key = fmt.Sprintf("%s_%s", DefaultEnvPrefix, key)
	}
	defaultMap[key] = val
}

func GetEnv[T any](key string) T {
	if DefaultEnvPrefix != "" {
		key = fmt.Sprintf("%s_%s", DefaultEnvPrefix, key)
	}

	val := os.Getenv(key)
	if val == "" {
		defVal, ok := defaultMap[key]
		if !ok {
			var zeroVal T
			return zeroVal
		}
		return defVal.(T)
	}

	var typeOf T
	parsedVal, _ := parseValue[T](val, typeOf).(T)
	return parsedVal
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

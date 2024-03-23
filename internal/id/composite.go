package id

import "strings"

// CompositeKeySeparatorPattern used by NewCompositeKey routine to separate keys.
var CompositeKeySeparatorPattern byte = '&'

// NewCompositeKey merges one or many keys into a single key.
func NewCompositeKey(keys ...string) string {
	totalKeys := len(keys)
	if totalKeys == 0 {
		return ""
	} else if totalKeys == 1 {
		return keys[0]
	}

	buf := strings.Builder{}

	for i, key := range keys {
		buf.WriteString(key)
		if i < totalKeys-1 {
			buf.WriteByte(CompositeKeySeparatorPattern)
		}
	}

	return buf.String()
}

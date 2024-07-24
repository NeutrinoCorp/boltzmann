package codec

import jsoniter "github.com/json-iterator/go"

// JSON is the JSON implementation of Codec.
type JSON struct {
}

var _ Codec = JSON{}

// MIMEType returns RFC MIME Type.
func (j JSON) MIMEType() string {
	return MIMETypeJSON
}

// Encode encodes the given Go structure.
func (j JSON) Encode(src any) ([]byte, error) {
	return jsoniter.Marshal(src)
}

// Decode decodes the given set of bytes as a Go structure.
func (j JSON) Decode(data []byte, v any) error {
	return jsoniter.Unmarshal(data, v)
}

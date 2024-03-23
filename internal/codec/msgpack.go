package codec

import "github.com/vmihailenco/msgpack/v5"

// Msgpack is the msgpack implementation of Codec.
type Msgpack struct {
}

var _ Codec = Msgpack{}

func (m Msgpack) MIMEType() string {
	return MIMETypeMsgpack
}

func (m Msgpack) Encode(src any) ([]byte, error) {
	return msgpack.Marshal(src)
}

func (m Msgpack) Decode(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}

package codec

import jsoniter "github.com/json-iterator/go"

type JSON struct {
}

var _ Codec = JSON{}

func (j JSON) MIMEType() string {
	return MIMETypeJSON
}

func (j JSON) Encode(src any) ([]byte, error) {
	return jsoniter.Marshal(src)
}

func (j JSON) Decode(data []byte, v any) error {
	return jsoniter.Unmarshal(data, v)
}

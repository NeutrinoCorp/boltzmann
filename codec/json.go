package codec

import jsoniter "github.com/json-iterator/go"

type JSON struct{}

var _ Codec = JSON{}

func (j JSON) Encode(src any) ([]byte, error) {
	return jsoniter.Marshal(src)
}

func (j JSON) Decode(src []byte, dst any) error {
	return jsoniter.Unmarshal(src, dst)
}

func (j JSON) ContentType() string {
	return "application/json"
}

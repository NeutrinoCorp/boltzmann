package codec

type Codec interface {
	MIMEType() string
	Encode(src any) ([]byte, error)
	Decode(data []byte, v any) error
}

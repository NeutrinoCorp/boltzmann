package codec

type Codec interface {
	ContentType() string
	Encode(src any) ([]byte, error)
	Decode(src []byte, dst any) error
}

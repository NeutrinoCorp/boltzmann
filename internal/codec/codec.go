package codec

// Codec is a utility-class component used to decode and encode Go structures.
type Codec interface {
	// MIMEType returns RFC MIME Type.
	MIMEType() string
	// Encode encodes the given Go structure.
	Encode(src any) ([]byte, error)
	// Decode decodes the given set of bytes as a Go structure.
	Decode(data []byte, v any) error
}

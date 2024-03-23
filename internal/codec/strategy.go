package codec

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

// TODO: Might be better to have a registry component for detailed instantiation
var codecMap = map[string]Codec{
	MIMETypeJSON: JSON{},
}

// Strategy is a high-level component used to extend codec capabilities (e.g., usage of many codecs from a registry,
// byte truncation).
// This component is concurrent-safe and zero-value is NOT ready to use;
// please create instances with NewStrategy routine.
type Strategy struct {
	mu                     *sync.Mutex
	truncatedPayloadWriter *bytes.Buffer
}

// NewStrategy allocates a new Strategy instance.
func NewStrategy() Strategy {
	return Strategy{
		mu:                     new(sync.Mutex),
		truncatedPayloadWriter: bytes.NewBuffer(nil),
	}
}

// Encode based on the `codecName` argument,
// converts the given structure into a set of bytes which have a specific format.
func (s Strategy) Encode(codecName string, src any) ([]byte, error) {
	c, ok := codecMap[codecName]
	if !ok {
		return nil, errors.New("codec: codec type not found")
	}

	return c.Encode(src)
}

// EncodeWithTruncation based on the `codecName` argument,
// converts the given structure into a set of bytes which have a specific format.
//
// If `truncateLimit` argument is greater than 0, then the original set of bytes will be truncated till the specified
// limit is reached.
func (s Strategy) EncodeWithTruncation(codecName string, truncateLimit int64, src any) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	encodedItem, err := s.Encode(codecName, src)
	if err != nil {
		return nil, err
	} else if truncateLimit <= 0 {
		return encodedItem, nil
	}

	s.truncatedPayloadWriter.Reset()
	reader := bytes.NewReader(encodedItem)
	lr := io.LimitReader(reader, truncateLimit)
	if _, err = s.truncatedPayloadWriter.ReadFrom(lr); err != nil {
		return nil, err
	}

	return s.truncatedPayloadWriter.Bytes(), nil
}

// Decode based on the `codecName` argument,
// converts the given set of bytes which have a specific format into a Go object
// (v argument, requires a structure pointer).
func (s Strategy) Decode(codecName string, data []byte, v any) error {
	c, ok := codecMap[codecName]
	if !ok {
		return errors.New("codec: codec type not found")
	}

	return c.Decode(data, v)
}

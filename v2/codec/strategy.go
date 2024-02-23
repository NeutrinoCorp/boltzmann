package codec

import "errors"

// TODO: Might be better to have a registry component for detailed instantiation
var codecMap = map[string]Codec{
	"json": JSON{},
}

type Strategy struct {
}

func (s Strategy) Encode(codecName string, src any) ([]byte, error) {
	c, ok := codecMap[codecName]
	if !ok {
		return nil, errors.New("codec: codec type not found")
	}

	return c.Encode(src)
}

package encoding

import (
	"fmt"
)

var (
	_ Encoder = (*textEncoder)(nil)

	// Text is based on ASCII encoder without the validation to generate an
	// error when there are non-ASCII characters. It returns the field content
	// as it is. This encoder is used for decoding of fields with content in
	// other idioms than english, like chinese, japanese, etc.
	Text = &textEncoder{}
)

type textEncoder struct{}

func (e textEncoder) Encode(data []byte) ([]byte, error) {
	return data, nil
}

func (e textEncoder) Decode(data []byte, length int) ([]byte, int, error) {
	// length should be positive
	if length < 0 {
		return nil, 0, fmt.Errorf("invalid length: %d", length)
	}

	if len(data) < length {
		return nil, 0, fmt.Errorf("not enough data to decode. expected len %d, got %d", length, len(data))
	}

	return data[:length], length, nil
}

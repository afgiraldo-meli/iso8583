package prefix

import "fmt"

// delimiterPrefixer implements Prefixer interface to allow looking for a
// delimiter in the content of a field which determines its end and therefore
// its length.
type delimiterPrefixer struct {
	delimiterChar byte
	encoder       string
}

// NewDelimiter creates a Prefixer which can searches for the given char byte
// in a field content to determine its end and therefore its length. Name is
// also required to complete the identity of the prefixer using the Inspect
// method.
// This Prefixer is not initialized like the others because the delimiter char
// must be provided to know what to look for.
// NOTE: The delimiter char is included in the length of the field.
//
// Example:
//
//	backslashPrefixer := NewDelimiter('\x5C', "ASCIIBackslash")
func NewDelimiter(char byte, name string) Prefixer {
	return &delimiterPrefixer{delimiterChar: char, encoder: name}
}

func (b *delimiterPrefixer) EncodeLength(maxLen, dataLen int) ([]byte, error) {
	if dataLen > maxLen {
		return nil, fmt.Errorf("field length: %d is larger than maximum: %d", dataLen, maxLen)
	}

	return []byte{}, nil
}

// DecodeLength iterates the content of a field by byte until the delimiter is
// reached, and returns the number of iterations required to find it. If the
// delimiter is not in the maximum length specified for the field, an error is
// returned.
func (b *delimiterPrefixer) DecodeLength(maxLen int, data []byte) (int, int, error) {
	var dataLen int
	for _, char := range data {
		dataLen++

		if dataLen > maxLen {
			return 0, 0, fmt.Errorf("delimiter not found in first %d bytes", maxLen)
		}

		if char == b.delimiterChar {
			return dataLen, 0, nil
		}
	}

	return 0, 0, fmt.Errorf("delimiter not found")
}

func (b *delimiterPrefixer) Inspect() string {
	return fmt.Sprintf("%sDelimiter", b.encoder)
}
